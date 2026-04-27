package fileextract

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nwaples/rardecode"
)

func extractRar(src, dest string) ([]string, error) {
	r, err := rardecode.OpenReader(src, "")
	if err != nil {
		return nil, fmt.Errorf("打开RAR文件失败: %w", err)
	}
	defer r.Close()

	var files []string
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return files, fmt.Errorf("读取RAR条目失败: %w", err)
		}

		name := ensureUTF8(header.Name)
		name = filepath.FromSlash(name)

		target, err := sanitizePath(dest, name)
		if err != nil {
			continue
		}

		if header.IsDir {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return files, fmt.Errorf("创建目录失败: %w", err)
		}

		out, err := os.Create(target)
		if err != nil {
			return files, fmt.Errorf("创建文件失败: %w", err)
		}
		if _, err = io.Copy(out, r); err != nil {
			out.Close()
			return files, fmt.Errorf("写入文件失败: %w", err)
		}
		out.Close()

		files = append(files, target)
	}

	return files, nil
}
