package fileextract

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/xiehqing/infra/pkg/logs"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardlehane/mscfb"
)

// ExtractDocxAttachments 从 DOCX 文件中提取嵌入的附件和媒体文件
// DOCX 本质上是 ZIP 格式，嵌入对象在 word/embeddings/，媒体在 word/media/
func ExtractDocxAttachments(src, dest string) ([]string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("打开DOCX文件失败: %w", err)
	}
	defer r.Close()

	baseName := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	attachDir := filepath.Join(dest, baseName+"_attachments")

	var files []string
	var fileId = 1
	for _, f := range r.File {
		name := f.Name
		isEmbedding := strings.HasPrefix(name, "word/embeddings/")
		isMedia := strings.HasPrefix(name, "word/media/")

		if (!isEmbedding && !isMedia) || f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil || len(data) == 0 {
			continue
		}
		os.MkdirAll(attachDir, 0755)
		fileName := filepath.Base(name)

		// 对 word/embeddings/ 下的 .bin 文件尝试解析 OLE 获取真实嵌入文件
		if isEmbedding && strings.EqualFold(filepath.Ext(fileName), ".bin") {
			if extracted, extName, ok := tryExtractOLEPayload(data); ok {
				if extName == DOCX || extName == XLSX {
					fileName = fmt.Sprintf("%s[附件%d]%s", getBaseFilename(src), fileId, extName)
					fileId += 1
				} else {
					fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + extName
				}
				data = extracted
			}
		}

		target := filepath.Join(attachDir, ensureUTF8(fileName))
		target = deduplicatePath(target)
		if filepath.Ext(fileName) != ".bin" && filepath.Ext(fileName) != ".emf" {
			if err := os.WriteFile(target, data, 0644); err != nil {
				continue
			}
			files = append(files, target)
		}
	}

	return files, nil
}

// tryExtractOLEPayload attempts to locate the embedded file inside an OLE
// Compound Binary File (CFBF / structured storage). It returns the raw payload,
// a suggested file extension, and whether extraction succeeded.
//
// The implementation is intentionally minimal: it walks the FAT chain for the
// first non-root directory entry whose stream looks like a known file type and
// reads the entire stream. This covers the common case where Office embeds a
// single document (xlsx, pptx, docx, pdf …) as an OLE package.
func tryExtractOLEPayload(data []byte) (payload []byte, ext string, ok bool) {
	if len(data) < 512 {
		return nil, "", false
	}

	// CFBF magic: D0 CF 11 E0 A1 B1 1A E1
	magic := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	if !bytes.Equal(data[:8], magic) {
		return nil, "", false
	}

	sectorSize := 1 << binary.LittleEndian.Uint16(data[30:32])
	if sectorSize < 512 || sectorSize > 4096 {
		return nil, "", false
	}

	fatSectors := int(binary.LittleEndian.Uint32(data[44:48]))
	dirStartSector := int(binary.LittleEndian.Uint32(data[48:52]))

	fat, err := buildFAT(data, sectorSize, fatSectors)
	if err != nil {
		return nil, "", false
	}

	dirData := readStream(data, sectorSize, dirStartSector, fat, -1)
	if dirData == nil {
		return nil, "", false
	}

	const dirEntrySize = 128
	entryCount := len(dirData) / dirEntrySize

	for i := 1; i < entryCount; i++ {
		entry := dirData[i*dirEntrySize : (i+1)*dirEntrySize]
		objType := entry[66]
		if objType == 0 {
			continue
		}

		streamSize := int(binary.LittleEndian.Uint32(entry[120:124]))
		if streamSize == 0 {
			continue
		}

		startSect := int(binary.LittleEndian.Uint32(entry[116:120]))

		miniStreamCutoff := 0x1000
		if streamSize < miniStreamCutoff {
			continue
		}

		stream := readStream(data, sectorSize, startSect, fat, streamSize)
		if stream == nil || len(stream) < 4 {
			continue
		}

		if detectedExt := detectFileExtension(stream); detectedExt != "" {
			return stream, detectedExt, true
		}
	}

	return nil, "", false
}

// tryExtractOLE 尝试解析 OLE 复合文档 (.bin)，提取其中嵌入的实际文件
func tryExtractOLE(data []byte, destDir, originalName string) []string {
	reader := bytes.NewReader(data)
	doc, err := mscfb.New(reader)
	if err != nil {
		return nil
	}

	var files []string
	for {
		entry, err := doc.Next()
		if err != nil {
			break
		}

		entryData, err := io.ReadAll(entry)
		if err != nil || len(entryData) == 0 {
			continue
		}

		switch entry.Name {
		case "\x01Ole10Native":
			logs.Infof("Ole10Native: %s", entry.Name)
			// Ole10Native 流包含原始文件名和实际文件数据
			name, fileData, err := parseOle10Native(entryData)
			if err != nil || len(fileData) == 0 {
				continue
			}
			if name == "" {
				name = guessFilename(fileData, originalName)
			}
			name = filepath.Base(ensureUTF8(name))
			target := deduplicatePath(filepath.Join(destDir, name))
			if os.WriteFile(target, fileData, 0644) == nil {
				files = append(files, target)
			}

		case "CONTENTS", "Package":
			logs.Infof("CONTENTS: %s", entry.Name)
			name := guessFilename(entryData, originalName)
			target := deduplicatePath(filepath.Join(destDir, name))
			if os.WriteFile(target, entryData, 0644) == nil {
				files = append(files, target)
			}
		}
	}

	return files
}

// parseOle10Native 解析 Ole10Native 二进制流格式：
//
//	[4B totalSize] [2B flags] [label\0] [filename\0] [4B reserved] [temppath\0] [4B dataSize] [data...]
func parseOle10Native(data []byte) (string, []byte, error) {
	if len(data) < 6 {
		return "", nil, fmt.Errorf("数据过短")
	}

	r := bytes.NewReader(data)

	var totalSize uint32
	if err := binary.Read(r, binary.LittleEndian, &totalSize); err != nil {
		return "", nil, err
	}

	var flags uint16
	if err := binary.Read(r, binary.LittleEndian, &flags); err != nil {
		return "", nil, err
	}

	label, _ := readNullString(r)
	filename, _ := readNullString(r)

	// reserved / unknown 4 bytes
	var reserved uint32
	binary.Read(r, binary.LittleEndian, &reserved)

	// temp path
	readNullString(r)

	var dataSize uint32
	if err := binary.Read(r, binary.LittleEndian, &dataSize); err != nil {
		return "", nil, err
	}

	if dataSize == 0 || int64(dataSize) > int64(r.Len()) {
		return "", nil, fmt.Errorf("无效的数据长度: %d", dataSize)
	}

	fileData := make([]byte, dataSize)
	if _, err := io.ReadFull(r, fileData); err != nil {
		return "", nil, err
	}

	name := filename
	if name == "" {
		name = label
	}
	return name, fileData, nil
}

func readNullString(r *bytes.Reader) (string, error) {
	var buf []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			return string(buf), err
		}
		if b == 0 {
			return string(buf), nil
		}
		buf = append(buf, b)
	}
}

// guessFilename 根据文件头魔数推测文件扩展名
func guessFilename(data []byte, fallback string) string {
	ext := detectExtByMagic(data)
	if ext != "" {
		base := strings.TrimSuffix(fallback, filepath.Ext(fallback))
		return base + ext
	}
	return fallback
}

func detectExtByMagic(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	switch {
	case bytes.HasPrefix(data, []byte("%PDF")):
		return ".pdf"
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		return ".zip"
	case bytes.HasPrefix(data, []byte("\x89PNG")):
		return ".png"
	case bytes.HasPrefix(data, []byte("\xFF\xD8\xFF")):
		return ".jpg"
	case bytes.HasPrefix(data, []byte("GIF8")):
		return ".gif"
	case bytes.HasPrefix(data, []byte("BM")):
		return ".bmp"
	case bytes.HasPrefix(data, []byte("RIFF")):
		if len(data) > 11 && string(data[8:12]) == "WEBP" {
			return ".webp"
		}
		return ".avi"
	case bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0}):
		return ".ole"
	default:
		return ""
	}
}

// deduplicatePath 处理同名文件冲突，自动追加 _1, _2 ... 后缀
func deduplicatePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	for i := 1; ; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

func buildFAT(data []byte, sectorSize, fatSectors int) ([]int32, error) {
	entriesPerSector := sectorSize / 4
	fat := make([]int32, 0, fatSectors*entriesPerSector)

	for i := 0; i < fatSectors; i++ {
		if 76+i*4+4 > len(data) {
			return nil, fmt.Errorf("FAT sector index out of range")
		}
		sectIdx := int(binary.LittleEndian.Uint32(data[76+i*4 : 76+i*4+4]))
		offset := 512 + sectIdx*sectorSize
		if offset+sectorSize > len(data) {
			return nil, fmt.Errorf("FAT sector data out of range")
		}
		for j := 0; j < entriesPerSector; j++ {
			v := int32(binary.LittleEndian.Uint32(data[offset+j*4 : offset+j*4+4]))
			fat = append(fat, v)
		}
	}
	return fat, nil
}

func readStream(data []byte, sectorSize, startSector int, fat []int32, streamSize int) []byte {
	var buf bytes.Buffer
	sect := startSector
	const endOfChain int32 = -2
	maxIter := len(fat) + 1

	for i := 0; sect >= 0 && int32(sect) != endOfChain && i < maxIter; i++ {
		offset := 512 + sect*sectorSize
		if offset+sectorSize > len(data) {
			break
		}
		buf.Write(data[offset : offset+sectorSize])
		if sect >= len(fat) {
			break
		}
		sect = int(fat[sect])
	}

	result := buf.Bytes()
	if streamSize > 0 && streamSize < len(result) {
		result = result[:streamSize]
	}
	return result
}

var fileSignatures = []struct {
	magic  []byte
	offset int
	ext    string
}{
	{[]byte{0x50, 0x4B, 0x03, 0x04}, 0, ".zip"},
	{[]byte("%PDF"), 0, ".pdf"},
	{[]byte{0x89, 0x50, 0x4E, 0x47}, 0, ".png"},
	{[]byte{0xFF, 0xD8, 0xFF}, 0, ".jpg"},
	{[]byte("GIF87a"), 0, ".gif"},
	{[]byte("GIF89a"), 0, ".gif"},
	{[]byte{0x42, 0x4D}, 0, ".bmp"},
	{[]byte("RIFF"), 0, ".wav"},
	{[]byte{0x00, 0x00, 0x00}, 0, ""},
}

func detectFileExtension(data []byte) string {
	// PK archive — further classify as xlsx/pptx/docx
	if len(data) > 4 && bytes.Equal(data[:4], []byte{0x50, 0x4B, 0x03, 0x04}) {
		return classifyOfficeOpenXML(data)
	}
	for _, sig := range fileSignatures {
		if sig.ext == "" || sig.ext == ".zip" {
			continue
		}
		end := sig.offset + len(sig.magic)
		if end > len(data) {
			continue
		}
		if bytes.Equal(data[sig.offset:end], sig.magic) {
			return sig.ext
		}
	}
	return ""
}

func classifyOfficeOpenXML(data []byte) string {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ".zip"
	}
	for _, f := range r.File {
		switch {
		case strings.HasPrefix(f.Name, "xl/"):
			return ".xlsx"
		case strings.HasPrefix(f.Name, "ppt/"):
			return ".pptx"
		case strings.HasPrefix(f.Name, "word/"):
			return ".docx"
		}
	}
	return ".zip"
}
