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
	var args []string
	switch editor {
	case "vim":
		args = []string{editor, "+normal G$", "+startinsert!", filename}
	case "nano":
		args = []string{editor, "+99999999", filename}
	default:
		args = []string{editor, filename}
	}
	if isWindows() {
		return prepareWindowsCommand(args...)
	}
	return prepareUnixCommand(args...)
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
