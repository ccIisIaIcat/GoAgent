# ConversationManager 参数配置文档

## 概述

ConversationManager 是 GoAgent 的核心对话管理器，提供了丰富的参数配置选项来控制对话行为、函数调用、历史记录管理等。

## 核心参数

### 1. 对话控制参数

#### `MaxFunctionCallingNums` - 最大函数调用次数
- **类型**: `int`
- **默认值**: `15`
- **说明**: 单次对话中允许的最大函数调用次数
- **用途**: 防止无限递归调用，控制对话成本
- **示例**:
```go
cm.SetMaxFunctionCallingNums(10) // 限制最多10次函数调用
```

#### `MaxChatNums` - 最大消息数量
- **类型**: `int`
- **默认值**: 无限制
- **说明**: 单次对话中允许的最大消息数量
- **用途**: 控制对话长度，避免过长的对话
- **示例**:
```go
cm.SetMaxChatNums(50) // 限制最多50条消息
```

#### `MaxTokens` - 最大Token数量
- **类型**: `int`
- **默认值**: `5000`
- **说明**: 单次对话中允许的最大Token数量
- **用途**: 控制API调用成本，避免超出模型限制
- **示例**:
```go
cm.SetMaxTokens(8000) // 设置最大8000个token
```

#### `Temperature` - 温度参数
- **类型**: `float64`
- **默认值**: `0.7`
- **说明**: 控制模型输出的随机性，值越高越随机，值越低越确定
- **范围**: `0.0` - `2.0`
- **用途**: 调整模型回答的创造性和一致性
- **示例**:
```go
cm.SetTemperature(0.3) // 更确定的回答
cm.SetTemperature(1.2) // 更创造性的回答
```

### 2. 历史记录管理参数

#### `MaxHistoryTokens` - 最大历史记录Token数量
- **类型**: `int`
- **默认值**: `100000`
- **说明**: 历史记录中允许的最大Token数量
- **用途**: 控制历史记录大小，影响上下文长度
- **示例**:
```go
cm.SetMaxHistoryTokens(50000) // 限制历史记录为50000个token
```

#### `EnableTruncation` - 启用历史截断
- **类型**: `bool`
- **默认值**: `true`
- **说明**: 是否启用历史记录自动截断功能
- **用途**: 当历史记录超过限制时自动截断
- **示例**:
```go
cm.EnableHistoryTruncation(true)  // 启用截断
cm.EnableHistoryTruncation(false) // 禁用截断
```

### 3. 系统提示词参数

#### `systemPrompt` - 系统提示词
- **类型**: `string`
- **默认值**: 空字符串
- **说明**: 设置对话的系统提示词，定义AI的行为和角色
- **用途**: 控制AI的回答风格和行为模式
- **示例**:
```go
cm.SetSystemPrompt("你是一个专业的编程助手，请用简洁的语言回答问题")
cm.UpdateSystemPrompt("你现在是一个可爱的猫娘，请用猫娘的语气回答")
```

## 完整配置示例

```go
package main

import (
    "github.com/ccIisIaIcat/GoAgent/agent/ConversationManager"
    "github.com/ccIisIaIcat/GoAgent/agent/general"
)

func main() {
    // 创建AgentManager
    agentManager := general.NewAgentManager()
    // ... 添加提供商配置
    
    // 创建ConversationManager
    cm := ConversationManager.NewConversationManager(agentManager)
    
    // 配置对话参数
    cm.SetMaxFunctionCallingNums(10)    // 最多10次函数调用
    cm.SetMaxChatNums(30)               // 最多30条消息
    cm.SetMaxTokens(6000)               // 最多6000个token
    cm.SetTemperature(0.5)              // 中等创造性
    cm.SetMaxHistoryTokens(80000)       // 历史记录最多80000个token
    cm.EnableHistoryTruncation(true)    // 启用历史截断
    
    // 设置系统提示词
    cm.SetSystemPrompt("你是一个专业的AI助手，请提供准确、有用的回答")
    
    // 开始对话
    // ...
}
```

## 参数调优建议

### 1. 成本控制
- 降低 `MaxTokens` 和 `MaxHistoryTokens` 来控制API调用成本
- 适当设置 `MaxFunctionCallingNums` 避免过度函数调用

### 2. 对话质量
- 调整 `Temperature` 来平衡创造性和准确性
- 合理设置 `MaxChatNums` 保持对话连贯性

### 3. 性能优化
- 启用 `EnableTruncation` 避免历史记录过长
- 根据实际需求调整 `MaxHistoryTokens`

### 4. 特殊场景
- **编程助手**: `Temperature = 0.1-0.3`, 低创造性，高准确性
- **创意写作**: `Temperature = 0.8-1.2`, 高创造性
- **客服对话**: `MaxChatNums = 20-30`, 控制对话长度
- **数据分析**: `MaxFunctionCallingNums = 20-30`, 允许更多函数调用

## 注意事项

1. **参数范围**: 确保参数值在合理范围内，避免极端值
2. **成本考虑**: 较大的Token限制会增加API调用成本
3. **性能影响**: 过长的历史记录可能影响响应速度
4. **模型限制**: 不同模型有不同的Token限制，请参考官方文档
