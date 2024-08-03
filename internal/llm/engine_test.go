package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms/openai"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/session/simple"
)

func TestEngine_ExecCompletion(t *testing.T) {
	llm, err := openai.New()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := options.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	e := &Engine{
		mode:        ChatEngineMode,
		config:      cfg,
		llm:         llm,
		channel:     make(chan EngineChatStreamOutput),
		pipe:        "",
		running:     false,
		chatHistory: simple.NewChatMessageHistory(),
		execHistory: simple.NewChatMessageHistory(),
	}
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"t1",
			args{
				"hello?",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := e.ExecCompletion(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecCompletion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
			assert.NotEmpty(t, got.Explanation)
		})
	}
}
