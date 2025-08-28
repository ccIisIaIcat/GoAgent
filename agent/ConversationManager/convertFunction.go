package ConversationManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// convertReturnValueToString 将函数返回值安全转换为字符串
func ConvertReturnValueToString(value reflect.Value) string {
	if !value.IsValid() {
		return "<invalid>"
	}

	// 处理 nil 值
	if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return "<nil>"
		}
	}

	actualValue := value.Interface()

	switch v := actualValue.(type) {
	case string:
		return v
	case error:
		if v == nil {
			return "<nil>"
		}
		return "ERROR: " + v.Error()
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.6f", v)
	default:
		// 对于复合类型(slice, map, struct等)，转换为JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("<%T: %v>", v, v)
		}
		return string(jsonBytes)
	}
}

// isValidParameterType 检查参数类型是否为基础类型或者array/slice/map
func IsValidParameterType(t reflect.Type) bool {
	kind := t.Kind()

	// 基础类型
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	case reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return false
	}
}

// isValidParameterType 检查参数类型是否为基础类型或者array/slice/map,还允许返回Error类型
func IsValidParameterTypeReturn(t reflect.Type) bool {
	kind := t.Kind()

	// 基础类型
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	case reflect.Array, reflect.Slice, reflect.Map:
		return true
	case reflect.Interface:
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if t.Implements(errorType) {
			return true
		}
		return false
	default:
		return false
	}
}

// convertToJSONSchemaType 将 Go 类型转换为 JSON Schema 类型名称
func ConvertToJSONSchemaType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Complex64, reflect.Complex128:
		return "string" // 复数类型转换为字符串表示
	case reflect.String:
		return "string"
	case reflect.Array, reflect.Slice:
		return "array"
	case reflect.Map:
		return "object"
	default:
		return "string" // 默认为 string
	}
}

// convertInterfaceToType 将JSON解析后的interface{}转换为指定的Go类型
func ConvertInterfaceToType(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	// 如果值为nil，返回零值
	if value == nil {
		return reflect.Zero(targetType), nil
	}

	// 如果类型已经匹配，直接返回
	if reflect.TypeOf(value) == targetType {
		return reflect.ValueOf(value), nil
	}

	switch targetType.Kind() {
	case reflect.Bool:
		if v, ok := value.(bool); ok {
			return reflect.ValueOf(v), nil
		}
		return reflect.Value{}, errors.New("无法转换为 bool 类型")

	case reflect.String:
		if v, ok := value.(string); ok {
			return reflect.ValueOf(v), nil
		}
		return reflect.Value{}, errors.New("无法转换为 string 类型")

	// 处理所有整数类型 (JSON解析后都是float64)
	case reflect.Int:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(int(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 int 类型")

	case reflect.Int8:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(int8(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 int8 类型")

	case reflect.Int16:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(int16(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 int16 类型")

	case reflect.Int32:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(int32(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 int32 类型")

	case reflect.Int64:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(int64(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 int64 类型")

	// 处理所有无符号整数类型
	case reflect.Uint:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(uint(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 uint 类型")

	case reflect.Uint8:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(uint8(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 uint8 类型")

	case reflect.Uint16:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(uint16(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 uint16 类型")

	case reflect.Uint32:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(uint32(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 uint32 类型")

	case reflect.Uint64:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(uint64(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 uint64 类型")

	case reflect.Uintptr:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(uintptr(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 uintptr 类型")

	// 处理浮点数类型
	case reflect.Float32:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(float32(v)), nil
		}
		return reflect.Value{}, errors.New("无法转换为 float32 类型")

	case reflect.Float64:
		if v, ok := value.(float64); ok {
			return reflect.ValueOf(v), nil
		}
		return reflect.Value{}, errors.New("无法转换为 float64 类型")

	// 暂时不支持复合类型 (array, slice, map)，可以后续扩展
	default:
		return reflect.Value{}, errors.New("不支持的类型转换: " + targetType.String())
	}
}
