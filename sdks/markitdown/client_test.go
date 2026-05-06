package markitdown

import (
	"os"
	"testing"
	"time"
)

func TestOpenFile(t *testing.T) {
	filepath := "C:\\zorktech\\国网吉林电力2026年第一次服务公开招标采购招标文件-服务采购文件.docx"
	t.Log(filepath)
	file, err := os.Open(filepath)
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()
	t.Log(file.Name())
}

func TestUploadFile(t *testing.T) {
	filepath := "C:\\zorktech\\国网吉林电力2026年第一次服务公开招标采购招标文件-服务采购文件.docx"
	client := NewClient("http://192.168.12.34:5001", WithTimeout(time.Duration(1000)*time.Second))
	convert, err := client.Convert(filepath)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(convert)
}
