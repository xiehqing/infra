package filex

import "os"

// CheckAndCreateDir 检查并创建目录
func CheckAndCreateDir(dir string) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dir, os.ModePerm)
		}
	}
}
