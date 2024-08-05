package runner

import (
	"fmt"
	"os/exec"
	"runtime"
)

const (
	defaultShellUnix = "bash"
	defaultShellWin  = "cmd.exe"
)

func Run(cmd string, arg ...string) (string, error) {
	out, err := exec.Command(cmd, arg...).Output()
	if err != nil {
		return fmt.Sprintf("error: %v", err), err
	}

	return string(out), nil
}

func PrepareInteractiveCommand(input string) *exec.Cmd {
	if isWindows() {
		return prepareWindowsCommand(input)
	}
	return prepareUnixCommand(input)
}

func PrepareEditSettingsCommand(editor, filename string) *exec.Cmd {
	switch editor {
	case "vim":
		return exec.Command(editor, "+normal G$", "+startinsert!", filename)
	case "nano":
		return exec.Command(editor, "+99999999", filename)
	default:
		return exec.Command(editor, filename)
	}
}

// IsCommandAvailable checks whether a command is available in the PATH.
func IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func prepareWindowsCommand(args ...string) *exec.Cmd {
	return exec.Command(
		defaultShellWin,
		append([]string{"/c"}, args...)...,
	)
}

func prepareUnixCommand(args ...string) *exec.Cmd {
	return exec.Command(
		defaultShellUnix,
		append([]string{"-c"}, args...)...,
	)
}
