package general

import (
	"GoAgent/agent/openai"
	"context"
)

// OpenAIProviderWrapper OpenAI提供商包装器
type OpenAIProviderWrapper struct {
	client *openai.Client
}

func (w *OpenAIProviderWrapper) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	resp, err := w.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	return convertToUnifiedResponse(resp), nil
}

func (w *OpenAIProviderWrapper) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
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

func (w *OpenAIProviderWrapper) GetProvider() Provider {
	return ProviderOpenAI
}

func (w *OpenAIProviderWrapper) ValidateRequest(req *ChatRequest) error {
	return w.client.ValidateRequest(req)
}
