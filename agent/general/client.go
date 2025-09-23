package general

import (
	"context"
	"fmt"

	"github.com/ccIisIaIcat/GoAgent/agent/anthropic"
	"github.com/ccIisIaIcat/GoAgent/agent/deepseek"
	"github.com/ccIisIaIcat/GoAgent/agent/google"
	"github.com/ccIisIaIcat/GoAgent/agent/openai"
	"github.com/ccIisIaIcat/GoAgent/agent/qwen"
)

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Provider Provider `json:"provider"`
	APIKey   string   `json:"api_key"`
	BaseURL  string   `json:"base_url,omitempty"`
	Model    string   `json:"model,omitempty"`
}

// AgentManager 智能体管理器
type AgentManager struct {
	PC        ProviderConfig
	providers map[Provider]LLMProvider
}

// NewAgentManager 创建智能体管理器
func NewAgentManager() *AgentManager {
	return &AgentManager{
		providers: make(map[Provider]LLMProvider),
	}
}

// AddProvider 添加提供商
func (m *AgentManager) AddProvider(config *ProviderConfig) error {
	switch config.Provider {
	case ProviderOpenAI:
		client := openai.NewClient(&openai.Config{
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
		})
		m.providers[ProviderOpenAI] = &OpenAIProviderWrapper{client: client}

	case ProviderAnthropic:
		client := anthropic.NewClient(&anthropic.Config{
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
		})
		m.providers[ProviderAnthropic] = &AnthropicProviderWrapper{client: client}

	case ProviderGoogle:
		client := google.NewClient(&google.Config{
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
		})
		m.providers[ProviderGoogle] = &GoogleProviderWrapper{client: client}

	case ProviderDeepSeek:
		client := deepseek.NewClient(&deepseek.Config{
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
		})
		m.providers[ProviderDeepSeek] = &DeepSeekProviderWrapper{client: client}

	case ProviderQwen:
		client := qwen.NewClient(&qwen.Config{
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
		})
		m.providers[ProviderQwen] = &QwenProviderWrapper{client: client}

	default:
		return fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	return nil
}

// GetProvider 获取提供商
func (m *AgentManager) GetProvider(provider Provider) (LLMProvider, error) {
	if p, exists := m.providers[provider]; exists {
		return p, nil
	}
	return nil, fmt.Errorf("provider %s not found", provider)
}

// Chat 发送聊天请求
func (m *AgentManager) Chat(ctx context.Context, provider Provider, req *ChatRequest) (*ChatResponse, error) {
	p, err := m.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	// 对maxtokens进行检查
	if req.MaxTokens == 0 {
		req.MaxTokens = 3000
	}
	// 对模型进行赋值
	if req.Model == "" {
		req.Model = getDefaultModel(provider)
	}

	if err := p.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("validate request failed: %w", err)
	}

	return p.Chat(ctx, req)
}

// ChatStream 发送流式聊天请求
func (m *AgentManager) ChatStream(ctx context.Context, provider Provider, req *ChatRequest) (<-chan *ChatResponse, error) {
	p, err := m.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	if err := p.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("validate request failed: %w", err)
	}

	return p.ChatStream(ctx, req)
}

// ListProviders 列出所有已注册的提供商
func (m *AgentManager) ListProviders() []Provider {
	var providers []Provider
	for p := range m.providers {
		providers = append(providers, p)
	}
	return providers
}
