package jsonx

import (
	"encoding/json"
	"github.com/xiehqing/infra/pkg/logs"
)

// ToJson 对象转换为json
func ToJson(o interface{}) (string, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ToJsonIgnoreError 对象转换为json，忽略错误
func ToJsonIgnoreError(o interface{}) string {
	if o == nil {
		logs.Errorf("[ToJsonIgnoreError]对象为nil")
		return ""
	}
	b, err := json.Marshal(o)
	if err != nil {
		logs.Errorf("[ToJsonIgnoreError]对象转换为json失败：%s", err.Error())
		return ""
	}
	return string(b)
}
