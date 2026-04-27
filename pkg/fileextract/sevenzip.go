package fileextract

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bodgit/sevenzip"
)

func extractSevenZip(src, dest string) ([]string, error) {
	r, err := sevenzip.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("打开7z文件失败: %w", err)
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		name := ensureUTF8(f.Name)
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

		rc, err := f.Open()
		if err != nil {
			return files, fmt.Errorf("打开7z条目失败 %s: %w", f.Name, err)
		}

		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return files, fmt.Errorf("创建文件失败: %w", err)
		}

		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()

		if err != nil {
			return files, fmt.Errorf("写入文件失败: %w", err)
		}
		files = append(files, target)
	}

	return files, nil
}
