package runner

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("PrepareUnixEditSettingsCommand", testPrepareUnixEditSettingsCommand)
	t.Run("PrepareWinEditSettingsCommand", testPrepareWinEditSettingsCommand)
}

func testPrepareUnixEditSettingsCommand(t *testing.T) {
	cmd := prepareUnixCommand("ai.json")

	expectedCmd := exec.Command(
		"bash",
		"-c",
		"ai.json; echo \"\n\";",
	)

	assert.Equal(t, expectedCmd.Args, cmd.Args, "The command arguments should be the same.")
}

func testPrepareWinEditSettingsCommand(t *testing.T) {
	cmd := prepareWindowsCommand("notepad.exe ai.json")

	expectedCmd := exec.Command(
		"cmd.exe",
		"/c",
		"notepad.exe ai.json",
	)

	assert.Equal(t, expectedCmd.Args, cmd.Args, "The command arguments should be the same.")
}
