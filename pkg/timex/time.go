package timex

import (
	"github.com/xiehqing/infra/pkg/logs"
	"time"
)

func ParseTime(layout, value string) (time.Time, error) {
	return ParseTimeWithTimeZone(layout, value, "Asia/Shanghai")
}

func ParseTimeWithTimeZone(layout, value string, tz string) (time.Time, error) {
	if tz == "" {
		return time.Parse(layout, value)
	}
	// 使用 time.LoadLocation 设置全局时区
	loc, err := time.LoadLocation(tz)
	if err != nil {
		logs.Error("Error loading location", err)
		return time.Parse(layout, value)
	}
	return time.ParseInLocation(layout, value, loc)
}
