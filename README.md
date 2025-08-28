# 多厂商LLM Agent

这是一个支持多家厂商LLM的golang agent，提供统一的接口来调用OpenAI、Anthropic、Google、DeepSeek等厂商的API。

## 特性

- 🔄 **统一接口**: 提供一套统一的数据结构和接口，支持在不同厂商间无缝切换
- 🛠️ **函数调用**: 支持所有厂商的function calling功能
- 🖼️ **多模态**: 支持图片输入的多模态对话
- 🌊 **流式响应**: 支持流式和非流式两种响应模式
- 📦 **模块化**: 每个厂商的实现独立在单独的文件夹中

## 项目结构

```
agent/
├── types.go                 # 统一的数据结构定义
├── manager.go              # 智能体管理器
├── main.go                 # 示例代码
├── agent/
│   ├── openai/            # OpenAI实现
│   │   ├── types.go       # OpenAI特定数据结构
│   │   ├── converter.go   # 数据格式转换器
│   │   └── client.go      # OpenAI客户端
│   ├── anthropic/         # Anthropic实现
│   │   ├── types.go
│   │   ├── converter.go
│   │   └── client.go
│   ├── google/            # Google实现
│   │   ├── types.go
│   │   ├── converter.go
│   │   └── client.go
│   └── deepseek/          # DeepSeek实现
│       ├── types.go
│       ├── converter.go
│       └── client.go
└── README.md
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置API密钥

在 `main.go` 中替换对应的API密钥：

```go
configs := []*ProviderConfig{
    {
        Provider: ProviderOpenAI,
        APIKey:   "your-openai-api-key",
        Model:    "gpt-4o",
    },
    {
        Provider: ProviderAnthropic,
        APIKey:   "your-anthropic-api-key",
        Model:    "claude-sonnet-4-20250514",
    },
    // ... 其他配置
}
```

### 3. 运行示例

```bash
go run .
```

## 使用方法

### 基础对话

```go
// 创建管理器
manager := NewAgentManager()

// 添加提供商
manager.AddProvider(&ProviderConfig{
    Provider: ProviderOpenAI,
    APIKey:   "your-api-key",
    Model:    "gpt-4o",
})

// 发送聊天请求
req := &ChatRequest{
    Model: "gpt-4o",
    Messages: []Message{
        {
            Role: RoleUser,
            Content: []Content{
                {
                    Type: ContentTypeText,
                    Text: "Hello, how are you?",
                },
            },
        },
    },
    MaxTokens:   1000,
    Temperature: 0.7,
}

ctx := context.Background()
resp, err := manager.Chat(ctx, ProviderOpenAI, req)
```

### 多模态对话（图片）

```go
req := &ChatRequest{
    Model: "gpt-4o",
    Messages: []Message{
        {
            Role: RoleUser,
            Content: []Content{
                {
                    Type: ContentTypeText,
                    Text: "What do you see in this image?",
                },
                {
                    Type: ContentTypeImageURL,
                    ImageURL: &ImageURL{
                        URL:    "data:image/jpeg;base64,/9j/4AAQ...",
                        Detail: DetailHigh,
                    },
                },
            },
        },
    },
    MaxTokens: 1000,
}

resp, err := manager.Chat(ctx, ProviderOpenAI, req)
```

### 函数调用

```go
req := &ChatRequest{
    Model: "gpt-4o",
    Messages: []Message{
        {
            Role: RoleUser,
            Content: []Content{
                {
                    Type: ContentTypeText,
                    Text: "What's the weather like in Beijing?",
                },
            },
        },
    },
    Tools: []Tool{
        {
            Type: "function",
            Function: FunctionDefinition{
                Name:        "get_weather",
                Description: "Get weather information",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "location": map[string]interface{}{
                            "type": "string",
                            "description": "The city name",
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    },
    MaxTokens: 1000,
}

resp, err := manager.Chat(ctx, ProviderOpenAI, req)
```

### 流式响应

```go
ch, err := manager.ChatStream(ctx, ProviderOpenAI, req)
if err != nil {
    log.Fatal(err)
}

for response := range ch {
    // 处理流式响应
    fmt.Printf("Received: %+v\n", response)
}
```

## 支持的厂商

| 厂商 | 对话 | 函数调用 | 多模态 | 流式 |
|------|------|---------|--------|------|
| OpenAI | ✅ | ✅ | ✅ | ✅ |
| Anthropic | ✅ | ✅ | ✅ | ✅ |
| Google | ✅ | ✅ | ✅ | ✅ |
| DeepSeek | ✅ | ❓ | ❓ | ✅ |

## 扩展新厂商

要添加新的厂商支持，需要：

1. 在 `agent/` 目录下创建新厂商文件夹
2. 实现三个文件：
   - `types.go`: 厂商特定的数据结构
   - `converter.go`: 统一格式与厂商格式的转换
   - `client.go`: HTTP客户端实现
3. 在 `manager.go` 中添加对应的包装器

## 注意事项

- 每个 `.go` 文件限制在300行以内
- 各厂商的API格式差异较大，转换逻辑可能需要进一步优化
- 图片处理目前主要支持base64格式，URL格式需要额外处理
- 部分厂商的特殊功能可能无法完全映射到统一接口

## 后续增加模型
Grok
Qwen
kimi
llama

## 许可证

MIT License