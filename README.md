# Multi-Vendor LLM Agent

This is a Golang agent that supports multiple LLM vendors, providing a unified interface to call APIs from OpenAI, Anthropic, Google, DeepSeek, and other vendors. It supports multimodal image uploads, function calling, MCP, and seamless switching between vendors during conversations.

## Features

- 🔄 **Unified Interface**: Provides a unified data structure and interface, supporting seamless switching between different vendors
- 🛠️ **Function Calling**: Supports function calling capabilities for all vendors
- 🖼️ **Multimodal**: Supports multimodal conversations with image input
- 🌊 **Streaming Response**: Supports both streaming and non-streaming response modes
- 📦 **MCP**: Supports MCP (Model Context Protocol) integration

## Quick Start

### 1. Install Dependencies

```bash
go get github.com/ccIisIaIcat/GoAgent@v1.0.2
go env -w GOTOOLCHAIN=auto
go mod tidy
```

### 2. Configure API Keys

Create a configuration file `LLMConfig.yaml`

```yaml
AgentAPIKey:
  # OpenAI Configuration
  OpenAI:
    BaseUrl: https://api.openai.com/v1  # Official URL, or domestic proxy address
    APIKey: your-openai-api-key-here
    Model: gpt-4o  # Optional, default gpt-4o, can also use gpt-4o-mini, gpt-3.5-turbo, etc.
  
  # Anthropic Configuration  
  Anthropic:
    BaseUrl: https://api.anthropic.com  # Official URL, or domestic proxy address
    APIKey: your-anthropic-api-key-here
    Model: claude-3-5-sonnet-20241022  # Optional, default claude-3-5-sonnet-20241022
  
  # DeepSeek Configuration
  DeepSeek:
    BaseUrl: https://api.deepseek.com
    APIKey: your-deepseek-api-key-here
    Model: deepseek-chat  # Optional, default deepseek-chat, can also use deepseek-coder
  
  # Google Configuration
  GoogleKey:
    BaseUrl: https://generativelanguage.googleapis.com/v1beta  # Official URL, or proxy address
    APIKey: your-google-api-key-here
    Model: gemini-2.5-pro  # Optional, default gemini-pro, can also use gemini-pro-vision

```

### 3. Run Examples

#### Simple Conversation with Vendor Switching
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
	cm.SetSystemPrompt("Please act as a cute cat girl and answer questions in a cat girl's tone")
	ret, finish_reason, err := cm.Chat(context.Background(), general.ProviderOpenAI, "Hello", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)

	ret, finish_reason, err = cm.Chat(context.Background(), general.ProviderDeepSeek, "How was your day today?", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)
}

```

#### Convenient Function Calling

You can use RegisterFunctionSimple to directly register functions to use, or use RegisterFunction to modify parameter names and descriptions

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
	return "Sunny weather"
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
	cm.SetSystemPrompt("Please act as a cute cat girl and answer questions in a cat girl's tone")
	cm.RegisterFunctionSimple("AddNumber", "Add two numbers", AddNumber)
	cm.RegisterFunction("GetWeather", "Get the weather of a city", GetWeather, []string{"city"}, []string{"The city to get the weather of"})

	ret, finish_reason, err := cm.Chat(context.Background(), general.ProviderOpenAI, "What time is it now?", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}
	fmt.Println(ret)
	fmt.Println(finish_reason)

	ret, finish_reason, err = cm.Chat(context.Background(), general.ProviderDeepSeek, "What is 787 plus 859?", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}
	fmt.Println(ret)
	fmt.Println(finish_reason)
}

```

#### MCP Integration

Create a configuration file

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

Initialize ConversationManager and load MCP configuration

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
	cm.SetSystemPrompt("Please act as a cute cat girl and answer questions in a cat girl's tone")
	fmt.Println("Loading MCP configuration")
	if err := cm.LoadMCPConfig("./mcp/example_config.json"); err != nil {
		log.Printf("Failed to load MCP configuration: %v", err)
	}
	// Ensure cleanup of resources
	defer cm.CloseMCP()
	ret, finish_reason, err := cm.Chat(context.Background(), general.ProviderOpenAI, "What time is it now?", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)

	ret, finish_reason, err = cm.Chat(context.Background(), general.ProviderDeepSeek, "What is 787 multiplied by 859?", []string{}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)

}
```

#### Image Analysis

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
	cm.SetSystemPrompt("Please act as a cute cat girl and answer questions in a cat girl's tone")

	imageBytes, err := os.ReadFile("image.png")
	if err != nil {
		log.Fatalf("❌ Failed to read image file: %v", err)
	}

	// Convert to base64
	imageData := base64.StdEncoding.EncodeToString(imageBytes)
	ret, finish_reason, err := cm.Chat(context.Background(), general.ProviderOpenAI, "Analyze this image and tell me what's in it", []string{imageData}, nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Println(ret)
	fmt.Println(finish_reason)

}

```

## Supported Vendors

| Vendor | Chat | Function Calling | Multimodal | Streaming |
|--------|------|------------------|------------|-----------|
| OpenAI | ✅ | ✅ | ✅ | ✅ |
| Anthropic | ✅ | ✅ | ✅ | ✅ |
| Google | ✅ | ✅ | ✅ | ✅ |
| DeepSeek | ✅ | ✅ | ❓ | ✅ |
| Qwen | ✅ | ✅ | ❓ | ✅ |

## Extending New Vendors

To add support for new vendors, you need to:

1. Create a new vendor folder under the `agent/` directory
2. Implement three files:
   - `types.go`: Vendor-specific data structures
   - `converter.go`: Conversion between unified format and vendor format
   - `client.go`: Client implementation
3. Add corresponding wrapper in `manager.go`

## Future Models to Add
Grok
kimi
llama
