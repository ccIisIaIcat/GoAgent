package general

import (
	"GoAgent/agent/google"
	"context"
)

// GoogleProviderWrapper Google提供商包装器
type GoogleProviderWrapper struct {
	client *google.Client
}

func (w *GoogleProviderWrapper) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	resp, err := w.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	return convertToUnifiedResponse(resp), nil
}

func (w *GoogleProviderWrapper) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
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

func (w *GoogleProviderWrapper) GetProvider() Provider {
	return ProviderGoogle
}

func (w *GoogleProviderWrapper) ValidateRequest(req *ChatRequest) error {
	return w.client.ValidateRequest(req)
}
