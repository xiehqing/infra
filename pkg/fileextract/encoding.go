package fileextract

import (
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// decodeGBK 将 GBK 编码的字符串转换为 UTF-8
func decodeGBK(s string) (string, error) {
	reader := transform.NewReader(
		strings.NewReader(s),
		simplifiedchinese.GBK.NewDecoder(),
	)
	b, err := io.ReadAll(reader)
	if err != nil {
		return s, err
	}
	return string(b), nil
}

// ensureUTF8 检测字符串编码，如果不是合法 UTF-8 则尝试从 GBK 转换
func ensureUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	if decoded, err := decodeGBK(s); err == nil {
		return decoded
	}
	return s
}
