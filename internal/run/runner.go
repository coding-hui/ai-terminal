package run

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
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

func PrepareEditSettingsCommand(input string) *exec.Cmd {
	if isWindows() {
		return prepareWindowsCommand(input)
	}
	return prepareUnixCommand(input)
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func prepareWindowsCommand(input string) *exec.Cmd {
	return exec.Command(
		defaultShellWin,
		"/c",
		strings.TrimRight(input, "&"),
	)
}

func prepareUnixCommand(input string) *exec.Cmd {
	return exec.Command(
		defaultShellUnix,
		"-c",
		fmt.Sprintf("%s; echo \"\n\";", strings.TrimRight(input, ";")),
	)
}
