package cryptox

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
)

type Md5 struct {
}

func NewMd5() *Md5 {
	return &Md5{}
}

func (m *Md5) Encrypt(data string) (string, error) {
	hash := md5.New()
	io.WriteString(hash, data)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (m *Md5) Decrypt(data string) (string, error) {
	return "", fmt.Errorf("md5: decrypt not supported")
}
