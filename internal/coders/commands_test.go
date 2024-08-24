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

}

func prepareTestCommands() *command {
	options.LoadConfig(viper.GetString(options.FlagAIConfig), "ai")
	coder := NewAutoCoder()
	coder.Init()

	// mock status
	select {
	case <-coder.checkpointChan:
	default:
	}

	return newCommand(coder)
}
