package general

import (
	"context"

	"github.com/ccIisIaIcat/GoAgent/agent/anthropic"
)

// AnthropicProviderWrapper Anthropic提供商包装器
type AnthropicProviderWrapper struct {
	client *anthropic.Client
}

func (w *AnthropicProviderWrapper) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	resp, err := w.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	return convertToUnifiedResponse(resp), nil
}

func (w *AnthropicProviderWrapper) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
	ch, err := w.client.ChatStream(ctx, req)
	if err != nil {
		return nil, err
	}

	unifiedCh := make(chan *ChatResponse, 10)
	go func() {
		defer close(unifiedCh)
		for resp := range ch {
			if converted := convertToUnifiedResponse(resp); converted != nil {
				unifiedCh <- converted
			}
		}
	}()

	return unifiedCh, nil
}

func (w *AnthropicProviderWrapper) GetProvider() Provider {
	return ProviderAnthropic
}

func (w *AnthropicProviderWrapper) ValidateRequest(req *ChatRequest) error {
	return w.client.ValidateRequest(req)
}
