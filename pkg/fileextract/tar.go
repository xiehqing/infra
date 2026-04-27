package fileextract

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

func extractTarGz(src, dest string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("打开GZIP流失败: %w", err)
	}
	defer gr.Close()

	return extractTarFromReader(gr, dest)
}

func extractTarBz2(src, dest string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	return extractTarFromReader(bzip2.NewReader(f), dest)
}

func extractTarXz(src, dest string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	xr, err := xz.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("打开XZ流失败: %w", err)
	}

	return extractTarFromReader(xr, dest)
}

func extractTar(src, dest string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	return extractTarFromReader(f, dest)
}

func extractTarFromReader(r io.Reader, dest string) ([]string, error) {
	tr := tar.NewReader(r)
	var files []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return files, fmt.Errorf("读取TAR条目失败: %w", err)
		}

		name := ensureUTF8(header.Name)
		name = filepath.FromSlash(name)

		target, err := sanitizePath(dest, name)
		if err != nil {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return files, fmt.Errorf("创建目录失败: %w", err)
			}
			out, err := os.Create(target)
			if err != nil {
				return files, fmt.Errorf("创建文件失败: %w", err)
			}
			if _, err = io.Copy(out, tr); err != nil {
				out.Close()
				return files, fmt.Errorf("写入文件失败: %w", err)
			}
			out.Close()
			files = append(files, target)
		}
	}

	return files, nil
}

// ---- 单文件解压 (非 tar 归档) ----

func decompressGz(src, dest string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("打开GZIP流失败: %w", err)
	}
	defer gr.Close()

	name := gr.Name
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(src), ".gz")
	}

	target := filepath.Join(dest, ensureUTF8(name))
	out, err := os.Create(target)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	if _, err = io.Copy(out, gr); err != nil {
		return nil, err
	}
	return []string{target}, nil
}

func decompressBz2(src, dest string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	name := strings.TrimSuffix(filepath.Base(src), ".bz2")
	target := filepath.Join(dest, ensureUTF8(name))
	out, err := os.Create(target)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	if _, err = io.Copy(out, bzip2.NewReader(f)); err != nil {
		return nil, err
	}
	return []string{target}, nil
}

func decompressXz(src, dest string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	xr, err := xz.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("打开XZ流失败: %w", err)
	}

	name := strings.TrimSuffix(filepath.Base(src), ".xz")
	target := filepath.Join(dest, ensureUTF8(name))
	out, err := os.Create(target)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	if _, err = io.Copy(out, xr); err != nil {
		return nil, err
	}
	return []string{target}, nil
}
