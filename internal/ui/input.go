package ui

import (
	"strings"
)

type Input struct {
	runMode    RunMode
	promptMode PromptMode
	args       string
	pipe       string
}

func NewInput(runMode RunMode, promptMode PromptMode, pipe string, prompts []string) (*Input, error) {
	if len(prompts) > 0 {
		runMode = CliMode
	}

	return &Input{
		runMode:    runMode,
		promptMode: promptMode,
		args:       strings.Join(prompts, " "),
		pipe:       pipe,
	}, nil
}

func (i *Input) GetRunMode() RunMode {
	return i.runMode
}

func (i *Input) GetPromptMode() PromptMode {
	return i.promptMode
}

func (i *Input) GetArgs() string {
	return i.args
}

func (i *Input) GetPipe() string {
	return i.pipe
}
