package pinyin

import (
	"github.com/mozillazg/go-pinyin"
	"strings"
)

// 常见多音姓氏映射表
var surnameMap = map[string]string{
	"单": "shan",
	"仇": "qiu",
	"区": "ou",
	"查": "zha",
	"盖": "ge",
	"黑": "he",
	"任": "ren",
	"华": "hua",
	"解": "xie",
	"折": "she",
	"朴": "piao",
	"繁": "po",
	"召": "shao",
	"种": "chong",
	"员": "yun",
	"曾": "zeng",
	"沈": "shen",
	"尉": "yu",
	"乐": "yue",
	"秘": "bi",
}

var pinyinArgs = pinyin.NewArgs()

// Generate 生成拼音
func Generate(content string) string {
	if len(content) == 0 {
		return ""
	}
	result := pinyin.LazyPinyin(content, pinyinArgs)
	return strings.Join(result, "")
}

// GenerateSurname 生成姓氏拼音
func GenerateSurname(name string) string {
	if len(name) == 0 {
		return ""
	}
	runes := []rune(name)
	surname := string(runes[0])
	// 检查是否是特殊多音姓氏
	if correctPinyin, exists := surnameMap[surname]; exists {
		if len(runes) > 1 {
			remaining := string(runes[1:])
			remainingPinyin := pinyin.LazyPinyin(remaining, pinyinArgs)
			return correctPinyin + strings.Join(remainingPinyin, "")
		}
		return correctPinyin
	}
	// 普通姓名直接转换
	return Generate(name)
}
