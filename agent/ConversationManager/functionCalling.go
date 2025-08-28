package ConversationManager

import (
	"GoAgent/agent/general"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// RegisterFunction 注册函数
func (cm *ConversationManager) RegisterFunctionSimple(name, description string, fn interface{}) error {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// 检查是否是函数类型
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("注册的对象不是函数类型")
	}

	// 验证参数类型
	numParams := fnType.NumIn()
	properties := make(map[string]interface{})
	required := make([]string, 0)
	paramNames := make([]string, numParams)

	for i := 0; i < numParams; i++ {
		paramType := fnType.In(i)
		if !IsValidParameterType(paramType) {
			return fmt.Errorf("参数 %d 类型 %s 不受支持", i, paramType.String())
		}

		paramName := fmt.Sprintf("param%d", i)
		paramNames[i] = paramName
		properties[paramName] = map[string]interface{}{
			"type":        ConvertToJSONSchemaType(paramType),
			"description": fmt.Sprintf("参数 %d (%s)", i, paramType.String()),
		}
		required = append(required, paramName)
	}

	// 验证返回值类型
	numReturns := fnType.NumOut()
	for i := 0; i < numReturns; i++ {
		returnType := fnType.Out(i)
		if !IsValidParameterTypeReturn(returnType) {
			return fmt.Errorf("返回值 %d 类型 %s 不受支持", i, returnType.String())
		}
	}

	// 创建工具定义
	tool := general.Tool{
		Type: "function",
		Function: general.FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		},
	}

	// 保存函数和工具定义
	cm.registeredFuncs[name] = fnValue
	cm.funcSchemas[name] = tool
	cm.funcParamNames[name] = paramNames
	cm.tools = append(cm.tools, tool)

	return nil
}

func (cm *ConversationManager) RegisterFunction(name, description string, fn interface{}, paramNames, paraDescriptions []string) error {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// 检查是否是函数类型
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("注册的对象不是函数类型")
	}
	// 验证参数类型
	numParams := fnType.NumIn()
	properties := make(map[string]interface{})
	required := make([]string, 0)

	if numParams != len(paramNames) || numParams != len(paraDescriptions) {
		return fmt.Errorf("参数数量不匹配")
	}

	for i := 0; i < numParams; i++ {
		paramType := fnType.In(i)
		if !IsValidParameterType(paramType) {
			return fmt.Errorf("参数 %d 类型 %s 不受支持", i, paramType.String())
		}
		paramName := paramNames[i]
		properties[paramName] = map[string]interface{}{
			"type":        ConvertToJSONSchemaType(paramType),
			"description": paraDescriptions[i],
		}
		required = append(required, paramName)
	}

	// 验证返回值类型
	numReturns := fnType.NumOut()
	for i := 0; i < numReturns; i++ {
		returnType := fnType.Out(i)
		if !IsValidParameterTypeReturn(returnType) {
			return fmt.Errorf("返回值 %d 类型 %s 不受支持", i, returnType.String())
		}
	}

	// 创建工具定义
	tool := general.Tool{
		Type: "function",
		Function: general.FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		},
	}

	// 保存函数和工具定义
	cm.registeredFuncs[name] = fnValue
	cm.funcSchemas[name] = tool
	cm.funcParamNames[name] = paramNames
	cm.tools = append(cm.tools, tool)

	return nil
}

func (cm *ConversationManager) ModifyFunctionParaDescription(name string, paraNames, paraDescriptions []string) error {
	fnValue, exists := cm.registeredFuncs[name]
	if !exists {
		return fmt.Errorf("未找到注册的函数: %s", name)
	}
	fnType := fnValue.Type()
	numParams := fnType.NumIn()

	// 验证参数数量是否匹配
	if numParams != len(paraNames) || numParams != len(paraDescriptions) {
		return fmt.Errorf("参数数量不匹配: 函数有 %d 个参数，但提供了 %d 个参数名和 %d 个参数描述",
			numParams, len(paraNames), len(paraDescriptions))
	}

	// 获取现有的工具定义
	tool, exists := cm.funcSchemas[name]
	if !exists {
		return fmt.Errorf("未找到函数的工具定义: %s", name)
	}

	// 重新构建参数属性
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for i := 0; i < numParams; i++ {
		paramType := fnType.In(i)
		paramName := paraNames[i]
		paramDescription := paraDescriptions[i]

		properties[paramName] = map[string]interface{}{
			"type":        ConvertToJSONSchemaType(paramType),
			"description": paramDescription,
		}
		required = append(required, paramName)
	}

	// 更新工具定义
	tool.Function.Parameters = map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}

	// 保存更新后的工具定义
	cm.funcSchemas[name] = tool

	// 更新tools切片中的对应工具
	for i, existingTool := range cm.tools {
		if existingTool.Function.Name == name {
			cm.tools[i] = tool
			break
		}
	}

	return nil
}

// CallRegisteredFunction 调用已注册的函数
func (cm *ConversationManager) CallRegisteredFunction(name string, arguments json.RawMessage) (string, error) {
	// 检查函数是否存在
	fnValue, exists := cm.registeredFuncs[name]
	if !exists {
		return "", fmt.Errorf("未找到注册的函数: %s", name)
	}

	fnType := fnValue.Type()

	// 解析参数
	var params map[string]interface{}
	if err := json.Unmarshal(arguments, &params); err != nil {
		// 尝试作为字符串解析（DeepSeek格式）
		var argsStr string
		if err2 := json.Unmarshal(arguments, &argsStr); err2 == nil {
			if err3 := json.Unmarshal([]byte(argsStr), &params); err3 != nil {
				return "", fmt.Errorf("解析参数失败: %w", err)
			}
		} else {
			return "", fmt.Errorf("解析参数失败: %w", err)
		}
	}

	// 获取注册时保存的参数名称
	savedParamNames, exists := cm.funcParamNames[name]
	if !exists {
		return "", fmt.Errorf("未找到函数 %s 的参数名称信息", name)
	}

	// 准备函数参数
	numIn := fnType.NumIn()
	args := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		paramName := savedParamNames[i]
		paramType := fnType.In(i)

		paramValue, exists := params[paramName]
		if !exists {
			// 如果参数不存在，使用零值
			args[i] = reflect.Zero(paramType)
		} else {
			// 转换参数类型
			convertedValue, err := ConvertInterfaceToType(paramValue, paramType)
			if err != nil {
				return "", fmt.Errorf("转换参数 %s 失败: %w", paramName, err)
			}
			args[i] = convertedValue
		}
	}

	// 调用函数
	results := fnValue.Call(args)

	// 处理返回值
	if len(results) == 0 {
		return "函数执行完成", nil
	}

	// 如果有多个返回值，拼接所有返回值
	var resultParts []string
	for i, result := range results {
		resultStr := ConvertReturnValueToString(result)
		if i == len(results)-1 && result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			// 如果最后一个返回值是错误类型且不为nil
			if !result.IsNil() {
				return "", fmt.Errorf("函数执行错误: %s", resultStr)
			}
			// 如果错误为nil，跳过这个返回值
			continue
		}
		resultParts = append(resultParts, resultStr)
	}

	if len(resultParts) == 0 {
		return "函数执行完成", nil
	}

	return fmt.Sprintf("函数返回: %s", resultParts[0]), nil
}

// HandleToolCall 处理工具调用（支持注册的函数）
func (cm *ConversationManager) HandleToolCall(ctx context.Context, provider general.Provider, toolCall general.ToolCall, info_chan chan general.Message) error {
	// 检查是否是注册的函数
	if _, exists := cm.registeredFuncs[toolCall.Function.Name]; exists {
		result, err := cm.CallRegisteredFunction(toolCall.Function.Name, toolCall.Function.Arguments)
		if err != nil {
			result = fmt.Sprintf("函数执行错误: %v", err)
		}

		// 添加工具结果到历史
		cm.AddMessage(general.RoleTool, []general.Content{
			{
				Type:   general.ContentTypeToolRes,
				Text:   result,
				ToolID: toolCall.ID,
			},
		})
		if info_chan != nil {
			info_chan <- general.Message{
				Role:    general.RoleTool,
				Content: []general.Content{{Type: general.ContentTypeToolRes, Text: result, ToolID: toolCall.ID}},
			}
		}

		return nil
	}

	// 如果不是注册的函数，返回错误
	return fmt.Errorf("未找到函数: %s", toolCall.Function.Name)
}

// hasToolCalls 检查消息是否包含工具调用
func (cm *ConversationManager) hasToolCalls(msg general.Message) bool {
	return len(msg.ToolCalls) > 0
}
