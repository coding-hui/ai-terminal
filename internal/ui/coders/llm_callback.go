package coders

import (
	"context"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/callbacks"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

type llmCallback struct{}

var _ callbacks.Handler = llmCallback{}

func (l llmCallback) HandleLLMGenerateContentStart(_ context.Context, ms []llms.MessageContent) {
}

func (l llmCallback) HandleLLMGenerateContentEnd(_ context.Context, res *llms.ContentResponse) {
}

func (l llmCallback) HandleStreamingFunc(_ context.Context, chunk []byte) {
}

func (l llmCallback) HandleText(_ context.Context, text string) {
}

func (l llmCallback) HandleLLMStart(_ context.Context, prompts []string) {
}

func (l llmCallback) HandleLLMError(_ context.Context, err error) {
}

func (l llmCallback) HandleChainStart(_ context.Context, inputs map[string]any) {
}

func (l llmCallback) HandleChainEnd(_ context.Context, outputs map[string]any) {
}

func (l llmCallback) HandleChainError(_ context.Context, err error) {
}

func (l llmCallback) HandleToolStart(_ context.Context, input string) {
}

func (l llmCallback) HandleToolEnd(_ context.Context, output string) {
}

func (l llmCallback) HandleToolError(_ context.Context, err error) {
}
