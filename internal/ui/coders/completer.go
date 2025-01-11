package coders

import (
	"strings"

	"github.com/coding-hui/go-prompt"

	"github.com/coding-hui/ai-terminal/internal/git"
)

type CommandCompleter struct {
	cmds []string
	repo *git.Command
}

func NewCommandCompleter(repo *git.Command) CommandCompleter {
	return CommandCompleter{
		cmds: getSupportedCommands(),
		repo: repo,
	}
}

func (c CommandCompleter) Complete(d prompt.Document) []prompt.Suggest {
	t := d.GetWordBeforeCursor()

	// if the input starts with "/", then we use the command completer
	if strings.HasPrefix(t, "/") {
		var completions []prompt.Suggest
		for _, v := range c.cmds {
			completions = append(completions, prompt.Suggest{Text: v})
		}
		t = strings.TrimPrefix(strings.TrimPrefix(t, "/"), "!")
		return prompt.FilterHasPrefix(completions, t, true)
	}

	// if the input starts with "@", then we use the file completer
	if strings.HasPrefix(t, "@") {
		files, _ := c.repo.ListAllFiles()
		var completions []prompt.Suggest
		for _, v := range files {
			completions = append(completions, prompt.Suggest{Text: v})
		}
		t = strings.TrimPrefix(t, "@")
		return prompt.FilterFuzzy(completions, t, true)
	}

	return []prompt.Suggest{}
}
