package convert

import (
	"fmt"
	"reflect"
	"strconv"
)

// ToInt64 将任意数字类型转换为int64
func ToInt64(input interface{}) (int64, error) {
	if input == nil {
		return 0, fmt.Errorf("unsupported type: %T", input)
	}
	kind := reflect.TypeOf(input).Kind()
	switch kind {
	case reflect.Float64:
		return int64(input.(float64)), nil
	case reflect.Float32:
		return int64(input.(float32)), nil
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Int16:
		return int64(reflect.ValueOf(input).Int()), nil
	case reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Uint16:
		return int64(reflect.ValueOf(input).Uint()), nil
	case reflect.String:
		return strconv.ParseInt(input.(string), 10, 64)
	case reflect.Bool:
		if input.(bool) {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported number type: %T", input)
	}
	return 0, nil
}
