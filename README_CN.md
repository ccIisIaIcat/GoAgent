# 多厂商LLM Agent

这是一个支持多家厂商LLM的golang agent，提供统一的接口来调用OpenAI、Anthropic、Google、DeepSeek等厂商的API，支持多模态图片上传，function calling，MCP,对话过程中无缝切换

## 特性

- 🔄 **统一接口**: 提供一套统一的数据结构和接口，支持在不同厂商间无缝切换
- 🛠️ **函数调用**: 支持所有厂商的function calling功能
- 🖼️ **多模态**: 支持图片输入的多模态对话
- 🌊 **流式响应**: 支持流式和非流式两种响应模式
- 📦 **mcp**: 支持MCP调用


## 快速开始

### 1. 安装依赖

```bash
go get github.com/ccIisIaIcat/GoAgent@v1.0.0
go mod tidy
```

### 2. 配置API密钥

创建密钥文件 `LLMConfig.yaml` 

```yaml
AgentAPIKey:
  # OpenAI配置
  OpenAI:
    BaseUrl: https://api.openai.com/v1  # 官方地址，或国内代理地址
    APIKey: your-openai-api-key-here
    Model: gpt-4o  # 可选，默认 gpt-4o，也可用 gpt-4o-mini, gpt-3.5-turbo 等
  
  # Anthropic配置  
  Anthropic:
    BaseUrl: https://api.anthropic.com  # 官方地址，或国内代理地址
    APIKey: your-anthropic-api-key-here
    Model: claude-3-5-sonnet-20241022  # 可选，默认 claude-3-5-sonnet-20241022
  
  # DeepSeek配置
  DeepSeek:
    BaseUrl: https://api.deepseek.com
    APIKey: your-deepseek-api-key-here
    Model: deepseek-chat  # 可选，默认 deepseek-chat，也可用 deepseek-coder
  
  # Google配置
  GoogleKey:
    BaseUrl: https://generativelanguage.googleapis.com/v1beta  # 官方地址，或代理地址
    APIKey: your-google-api-key-here
    Model: gemini-2.5-pro  # 可选，默认 gemini-pro，也可用 gemini-pro-vision

```

### 3. 运行示例

#### 简答对话，切换厂商
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ccIisIaIcat/GoAgent/agent/ConversationManager"
	"github.com/ccIisIaIcat/GoAgent/agent/general"
)

func main() {
	config, err := general.LoadConfig("./LLMConfig.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	agentManager := general.NewAgentManager()
	for _, v := range config.ToProviderConfigs() {
		agentManager.AddProvider(v)
	}
	cm := ConversationManager.NewConversationManager(agentManager)
	cm.SetSystemPrompt("请扮演一只可爱的猫娘，用猫娘的语气回答问题")
	ret, finish_reason, err := cm.Chat(context.Background(), general.ProviderOpenAI, "你好", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)

	ret, finish_reason, err = cm.Chat(context.Background(), general.ProviderDeepSeek, "今天过得怎么样？", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)
}

```

#### 便捷的函数调用

可以使用RegisterFunctionSimple直接注册要使用的函数，也可以使用RegisterFunction对参数名称和描述进行修改

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ccIisIaIcat/GoAgent/agent/general"

	"github.com/ccIisIaIcat/GoAgent/agent/ConversationManager"
)

func AddNumber(a, b int) int {
	return a + b
}

func GetWeather(city string) string {
	return "天气晴朗"
}

func main() {
	config, err := general.LoadConfig("./LLMConfig.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	agentManager := general.NewAgentManager()
	for _, v := range config.ToProviderConfigs() {
		agentManager.AddProvider(v)
	}
	cm := ConversationManager.NewConversationManager(agentManager)
	cm.SetSystemPrompt("请扮演一只可爱的猫娘，用猫娘的语气回答问题")
	cm.RegisterFunctionSimple("AddNumber", "Add two numbers", AddNumber)
	cm.RegisterFunction("GetWeather", "Get the weather of a city", GetWeather, []string{"city"}, []string{"The city to get the weather of"})

	ret, finish_reason, err := cm.Chat(context.Background(), general.ProviderOpenAI, "请问现在几点了", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}
	fmt.Println(ret)
	fmt.Println(finish_reason)

	ret, finish_reason, err = cm.Chat(context.Background(), general.ProviderDeepSeek, "请问787加上859等于多少？", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}
	fmt.Println(ret)
	fmt.Println(finish_reason)
}

```

#### MCP接入

创建配置文件

```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": [
        "@playwright/mcp@latest",
        "--isolated"
      ]
    },
    "desktop-commander": {
      "command": "npx",
      "args": [
        "@modelcontextprotocol/server-everything"
      ]
    }
  }
}
```

初始化ConversationManager并加载MCP配置

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ccIisIaIcat/GoAgent/agent/general"

	"github.com/ccIisIaIcat/GoAgent/agent/ConversationManager"
)

func main() {
	config, err := general.LoadConfig("./LLMConfig.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	agentManager := general.NewAgentManager()
	for _, v := range config.ToProviderConfigs() {
		agentManager.AddProvider(v)
	}
	cm := ConversationManager.NewConversationManager(agentManager)
	cm.SetSystemPrompt("请扮演一只可爱的猫娘，用猫娘的语气回答问题")
	fmt.Println("加载MCP配置")
	if err := cm.LoadMCPConfig("./mcp/example_config.json"); err != nil {
		log.Printf("加载MCP配置失败: %v", err)
	}
	// 确保清理资源
	defer cm.CloseMCP()
	ret, finish_reason, err := cm.Chat(context.Background(), general.ProviderOpenAI, "请问现在几点了", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)

	ret, finish_reason, err = cm.Chat(context.Background(), general.ProviderDeepSeek, "请问787乘上859等于多少？", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)

}
```

#### 图片解析

```go
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/ccIisIaIcat/GoAgent/agent/general"

	"github.com/ccIisIaIcat/GoAgent/agent/ConversationManager"
)

func main() {
	config, err := general.LoadConfig("./LLMConfig.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	agentManager := general.NewAgentManager()
	for _, v := range config.ToProviderConfigs() {
		agentManager.AddProvider(v)
	}
	cm := ConversationManager.NewConversationManager(agentManager)
	cm.SetSystemPrompt("请扮演一只可爱的猫娘，用猫娘的语气回答问题")

	imageBytes, err := os.ReadFile("image.png")
	if err != nil {
		log.Fatalf("❌ Failed to read image file: %v", err)
	}

	// Convert to base64
	imageData := base64.StdEncoding.EncodeToString(imageBytes)
	ret, finish_reason, err := cm.Chat(context.Background(), general.ProviderOpenAI, "分析这张图片，并告诉我图片中有什么", []string{imageData}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)

}

```

## 支持的厂商

| 厂商 | 对话 | 函数调用 | 多模态 | 流式 |
|------|------|---------|--------|------|
| OpenAI | ✅ | ✅ | ✅ | ✅ |
| Anthropic | ✅ | ✅ | ✅ | ✅ |
| Google | ✅ | ✅ | ✅ | ✅ |
| DeepSeek | ✅ | ✅ | ❓ | ✅ |

## 扩展新厂商

要添加新的厂商支持，需要：

1. 在 `agent/` 目录下创建新厂商文件夹
2. 实现三个文件：
   - `types.go`: 厂商特定的数据结构
   - `converter.go`: 统一格式与厂商格式的转换
   - `client.go`: 客户端实现
3. 在 `manager.go` 中添加对应的包装器


## 后续希望增加模型
Grok
Qwen
kimi
llama
