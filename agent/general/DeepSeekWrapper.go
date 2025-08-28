package general

import (
	"context"

	"github.com/ccIisIaIcat/GoAgent/agent/deepseek"
)

// DeepSeekProviderWrapper DeepSeek提供商包装器
type DeepSeekProviderWrapper struct {
	client *deepseek.Client
}

func (w *DeepSeekProviderWrapper) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	resp, err := w.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	return convertToUnifiedResponse(resp), nil
}

func (w *DeepSeekProviderWrapper) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
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

func (w *DeepSeekProviderWrapper) GetProvider() Provider {
	return ProviderDeepSeek
}

func (w *DeepSeekProviderWrapper) ValidateRequest(req *ChatRequest) error {
	return w.client.ValidateRequest(req)
}
