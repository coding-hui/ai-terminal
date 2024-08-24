package coders

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/coding-hui/ai-terminal/internal/cli/options"

	_ "github.com/coding-hui/ai-terminal/internal/session/mongo"
	_ "github.com/coding-hui/ai-terminal/internal/session/simple"
)

func TestCommands(t *testing.T) {
	t.Run("codingCmd", testCodingCmd)
}

func testCodingCmd(t *testing.T) {
	cmds := prepareTestCommands()

	cmds.addFiles("internal/coders/code_editor.go")

	cmds.coding("添加注释")
}

func prepareTestCommands() *command {
	options.LoadConfig(viper.GetString(options.FlagAIConfig), "ai")
	coder := NewAutoCoder()
	coder.Init()

	// mock status
	go func() {
		for {
			select {
			case <-coder.checkpointChan:
			}
		}
	}()

	return newCommand(coder)
}
