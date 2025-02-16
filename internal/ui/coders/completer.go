package coders

import (
	"strings"

	"github.com/elk-language/go-prompt"
	pstrings "github.com/elk-language/go-prompt/strings"

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

func (c CommandCompleter) Complete(d prompt.Document) (suggestions []prompt.Suggest, startChar, endChar pstrings.RuneNumber) {
	endIndex := d.CurrentRuneIndex()
	w := d.GetWordBeforeCursor()
	startIndex := endIndex - pstrings.RuneCount([]byte(w))

	// if the input starts with "/", then we use the command completer
	if strings.HasPrefix(w, "/") {
		var completions []prompt.Suggest
		for _, v := range c.cmds {
			completions = append(completions, prompt.Suggest{Text: v})
		}
		return prompt.FilterHasPrefix(completions, w, true), startIndex, endIndex
	}

	// if the input starts with "@", then we use the file completer
	if strings.HasPrefix(w, "@") {
		files, _ := c.repo.ListAllFiles()
		var completions []prompt.Suggest
		for _, v := range files {
			completions = append(completions, prompt.Suggest{Text: v})
		}
		w = strings.TrimPrefix(w, "@")
		return prompt.FilterFuzzy(completions, w, true), startIndex, endIndex
	}

	// if the input starts with "--", then we use the flag completer
	if strings.HasPrefix(w, "--") {
		completions := []prompt.Suggest{
			{Text: "--verbose"},
			{Text: "--help"},
		}
		return prompt.FilterContains(completions, w, true), startIndex, endIndex
	}

	return prompt.FilterHasPrefix([]prompt.Suggest{}, w, true), startIndex, endIndex
}
