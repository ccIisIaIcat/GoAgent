package general

import (
	"context"
)

// QwenProviderWrapper Qwen提供商包装器
type QwenProviderWrapper struct {
	client interface {
		Chat(ctx context.Context, req interface{}) (interface{}, error)
		ChatStream(ctx context.Context, req interface{}) (<-chan interface{}, error)
		GetProvider() string
		ValidateRequest(req interface{}) error
	}
}

// Chat 发送聊天请求
func (w *QwenProviderWrapper) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	resp, err := w.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为统一响应格式
	return convertToUnifiedResponse(resp), nil
}

// ChatStream 发送流式聊天请求
func (w *QwenProviderWrapper) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
	ch, err := w.client.ChatStream(ctx, req)
	if err != nil {
		return nil, err
	}

	// 创建转换通道
	resultCh := make(chan *ChatResponse, 10)

	go func() {
		defer close(resultCh)
		for resp := range ch {
			// 转换为统一格式
			converted := convertToUnifiedResponse(resp)
			if converted != nil {
				resultCh <- converted
			}
		}
	}()

	return resultCh, nil
}

// GetProvider 获取提供商名称
func (w *QwenProviderWrapper) GetProvider() Provider {
	return ProviderQwen
}

// ValidateRequest 验证请求参数
func (w *QwenProviderWrapper) ValidateRequest(req *ChatRequest) error {
	return w.client.ValidateRequest(req)
}
