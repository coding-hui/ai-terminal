package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) ExecCompletion(input string) (*EngineChatStreamOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*EngineChatStreamOutput), args.Error(1)
}

func TestEngine_ExecCompletion(t *testing.T) {
	mockEngine := new(MockEngine)

	mockOutput := &EngineChatStreamOutput{
		content: "mock explanation",
	}

	mockEngine.On("ExecCompletion", "hello?").Return(mockOutput, nil)

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

			got, err := mockEngine.ExecCompletion(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecCompletion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
			assert.NotEmpty(t, got.content)
		})
	}
}
