package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Input struct {
	runMode    RunMode
	promptMode PromptMode
	args       string
	pipe       string
}

func NewInput(runMode RunMode, promptMode PromptMode, args []string) (*Input, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Error getting stat:", err)
		return nil, err
	}

	pipe := ""
	if !(stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0) {
		reader := bufio.NewReader(os.Stdin)
		var builder strings.Builder

		for {
			r, _, err := reader.ReadRune()
			if err != nil && err == io.EOF {
				break
			}
			_, err = builder.WriteRune(r)
			if err != nil {
				fmt.Println("Error getting input:", err)
				return nil, err
			}
		}

		pipe = strings.TrimSpace(builder.String())
	}

	if len(args) > 0 {
		runMode = CliMode
	}

	return &Input{
		runMode:    runMode,
		promptMode: promptMode,
		args:       strings.Join(args, " "),
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
