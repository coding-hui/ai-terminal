package ai

import (
	"context"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

const (
	ModelTypeOpenAI = "openai"
	ModelTypeARK    = "ark"
)

type Model interface {
	GenerateContent(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error)
}
