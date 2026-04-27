package fileextract

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileType 文件类型枚举
type FileType int

const (
	TypeUnknown FileType = iota
	TypeZip
	TypeTar
	TypeTarGz
	TypeTarBz2
	TypeTarXz
	TypeRar
	Type7z
	TypeGz
	TypeBz2
	TypeXz
	TypeDocx
	TypeXlsx
	TypePptx
)

// Extractor 解压器，支持递归解压嵌套压缩文件、DOCX 附件提取
type Extractor struct {
	MaxDepth int // 最大递归深度，防止压缩炸弹
}

var extractor = New(WithMaxDepth(10))

func Default() *Extractor {
	return extractor
}

type Option func(*Extractor)

func WithMaxDepth(depth int) Option {
	return func(e *Extractor) { e.MaxDepth = depth }
}

func New(opts ...Option) *Extractor {
	e := &Extractor{MaxDepth: 10}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// ExtractAll 将压缩文件 src 解压到 dest 目录，递归处理嵌套压缩文件和 DOCX 附件
// 返回最终提取到的所有文件路径列表
func (e *Extractor) ExtractAll(src, dest string) ([]string, error) {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return nil, fmt.Errorf("创建目标目录失败: %w", err)
	}
	return e.extract(src, dest, 0)
}

// extract 执行实际的解压操作
func (e *Extractor) extract(src, dest string, depth int) ([]string, error) {
	if depth > e.MaxDepth {
		return nil, fmt.Errorf("超过最大递归深度 %d", e.MaxDepth)
	}

	ft := DetectFileType(src)

	var extractedFiles []string
	var err error

	switch ft {
	case TypeZip:
		extractedFiles, err = extractZip(src, dest)
	case TypeTarGz:
		extractedFiles, err = extractTarGz(src, dest)
	case TypeTarBz2:
		extractedFiles, err = extractTarBz2(src, dest)
	case TypeTarXz:
		extractedFiles, err = extractTarXz(src, dest)
	case TypeTar:
		extractedFiles, err = extractTar(src, dest)
	case TypeRar:
		extractedFiles, err = extractRar(src, dest)
	case Type7z:
		extractedFiles, err = extractSevenZip(src, dest)
	case TypeGz:
		extractedFiles, err = decompressGz(src, dest)
	case TypeBz2:
		extractedFiles, err = decompressBz2(src, dest)
	case TypeXz:
		extractedFiles, err = decompressXz(src, dest)
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s", src)
	}

	if err != nil {
		return extractedFiles, fmt.Errorf("解压失败 [%s]: %w", filepath.Base(src), err)
	}

	var allFiles []string
	for _, f := range extractedFiles {
		processed, _ := e.processFile(f, depth)
		allFiles = append(allFiles, processed...)
	}

	return allFiles, nil
}

// processFile 检查已解压的文件，如果是压缩文件则递归解压，如果是 DOCX 则提取附件
func (e *Extractor) processFile(path string, depth int) ([]string, error) {
	if depth > e.MaxDepth {
		return []string{path}, nil
	}

	ft := DetectFileType(path)

	if isArchiveType(ft) {
		archiveName := trimArchiveExt(filepath.Base(path))
		subDest := filepath.Join(filepath.Dir(path), archiveName)
		os.MkdirAll(subDest, 0755)

		subFiles, err := e.extract(path, subDest, depth+1)
		if err != nil {
			return []string{path}, nil
		}
		os.Remove(path)
		return subFiles, nil
	}

	if ft == TypeDocx {
		var result []string
		result = append(result, path)

		attachments, err := ExtractDocxAttachments(path, filepath.Dir(path))
		if err == nil {
			for _, att := range attachments {
				processed, _ := e.processFile(att, depth+1)
				result = append(result, processed...)
			}
		}
		return result, nil
	}

	return []string{path}, nil
}

// DetectFileType 通过文件扩展名和魔数字节检测文件类型
func DetectFileType(path string) FileType {
	ext := strings.ToLower(filepath.Ext(path))
	name := strings.ToLower(filepath.Base(path))

	// 复合扩展名优先检测
	if strings.HasSuffix(name, ".tar.gz") || ext == ".tgz" {
		return TypeTarGz
	}
	if strings.HasSuffix(name, ".tar.bz2") || ext == ".tbz2" || ext == ".tbz" {
		return TypeTarBz2
	}
	if strings.HasSuffix(name, ".tar.xz") || ext == ".txz" {
		return TypeTarXz
	}

	switch ext {
	case ".zip":
		return TypeZip
	case ".tar":
		return TypeTar
	case ".rar":
		return TypeRar
	case ".7z":
		return Type7z
	case ".gz":
		return TypeGz
	case ".bz2":
		return TypeBz2
	case ".xz":
		return TypeXz
	case ".docx":
		return TypeDocx
	case ".xlsx":
		return TypeXlsx
	case ".pptx":
		return TypePptx
	}

	return detectByMagic(path)
}

// detectByMagic 通过文件头魔数判断文件类型
func detectByMagic(path string) FileType {
	f, err := os.Open(path)
	if err != nil {
		return TypeUnknown
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return TypeUnknown
	}
	buf = buf[:n]

	switch {
	case bytes.HasPrefix(buf, []byte("PK\x03\x04")):
		return TypeZip
	case bytes.HasPrefix(buf, []byte("Rar!\x1a\x07")):
		return TypeRar
	case bytes.HasPrefix(buf, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}):
		return Type7z
	case bytes.HasPrefix(buf, []byte{0x1f, 0x8b}):
		return TypeGz
	case bytes.HasPrefix(buf, []byte("BZh")):
		return TypeBz2
	case bytes.HasPrefix(buf, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}):
		return TypeXz
	case len(buf) > 262 && string(buf[257:262]) == "ustar":
		return TypeTar
	}

	return TypeUnknown
}

func isArchiveType(ft FileType) bool {
	switch ft {
	case TypeZip, TypeTar, TypeTarGz, TypeTarBz2, TypeTarXz,
		TypeRar, Type7z, TypeGz, TypeBz2, TypeXz:
		return true
	}
	return false
}

// trimArchiveExt 移除压缩文件扩展名，支持 .tar.gz 等复合扩展名
func trimArchiveExt(name string) string {
	lower := strings.ToLower(name)
	for _, suffix := range []string{".tar.gz", ".tar.bz2", ".tar.xz"} {
		if strings.HasSuffix(lower, suffix) {
			return name[:len(name)-len(suffix)]
		}
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// sanitizePath 防止 Zip Slip 路径穿越攻击
func sanitizePath(dest, name string) (string, error) {
	p := filepath.Join(dest, name)
	absP, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	absD, err := filepath.Abs(dest)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absP, absD+string(os.PathSeparator)) {
		return "", fmt.Errorf("非法路径(可能存在路径穿越): %s", name)
	}
	return p, nil
}
