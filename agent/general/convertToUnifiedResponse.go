package general

import "encoding/json"

// convertToUnifiedResponse 将各种响应格式转换为统一响应格式
func convertToUnifiedResponse(resp interface{}) *ChatResponse {
	// 这里需要进行类型断言和转换
	// 简化实现，实际应该根据响应的具体结构进行转换
	// 由于各个provider的converter已经返回了统一格式的interface{}
	// 这里可以进行JSON序列化/反序列化来转换

	if resp == nil {
		return nil
	}

	// 使用JSON编码/解码来实现通用转换
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil
	}

	var unified ChatResponse
	if err := json.Unmarshal(respBytes, &unified); err != nil {
		return nil
	}

	return &unified
}
