package fileextract

import (
	"path/filepath"
	"strings"
)

const (
	DOCX string = ".docx"
	XLSX string = ".xlsx"
)

// getBaseFilename 获取文件名称
func getBaseFilename(filename string) string {
	// "第一章 投标邀请书及招标公告(312601).docx"
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(base, ext)
}
