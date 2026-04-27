package fileextract

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func extractZip(src, dest string) ([]string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("打开ZIP文件失败: %w", err)
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		name := f.Name
		// ZIP 在 Windows 上创建时常使用 GBK 编码文件名
		// NonUTF8 为 true 时表示文件名非 UTF-8 编码
		if f.NonUTF8 {
			name = ensureUTF8(name)
		}
		name = filepath.FromSlash(name)

		target, err := sanitizePath(dest, name)
		if err != nil {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return files, fmt.Errorf("创建目录失败: %w", err)
		}

		if err := writeZipEntry(f, target); err != nil {
			return files, err
		}
		files = append(files, target)
	}

	return files, nil
}

func writeZipEntry(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("打开ZIP条目失败 %s: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("创建文件失败 %s: %w", target, err)
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}
