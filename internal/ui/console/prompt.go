package console

import (
	"github.com/elk-language/go-prompt"
)

const suggestLimit = 5

var (
	options = []prompt.Option{
		prompt.WithPrefixTextColor(prompt.Turquoise),
		prompt.WithCompletionOnDown(),
		prompt.WithSuggestionBGColor(prompt.DarkGray),
		prompt.WithSuggestionTextColor(prompt.White),
		prompt.WithDescriptionBGColor(prompt.LightGray),
		prompt.WithDescriptionTextColor(prompt.Black),
		prompt.WithSelectedSuggestionBGColor(prompt.Black),
		prompt.WithSelectedSuggestionTextColor(prompt.White),
		prompt.WithSelectedDescriptionBGColor(prompt.DarkGray),
		prompt.WithScrollbarThumbColor(prompt.Black),
		prompt.WithScrollbarBGColor(prompt.White),
		prompt.WithMaxSuggestion(suggestLimit),
		prompt.WithKeyBind(makeNecessaryKeyBindings()...),
	}
)

func NewPrompt(prefix string, enableColor bool, completer prompt.Completer, exec func(string)) *prompt.Prompt {
	promptOptions := append(options, prompt.WithPrefix(prefix+" > "))
	promptOptions = append(promptOptions, prompt.WithCompleter(completer))
	if !enableColor {
		promptOptions = append(promptOptions, prompt.WithPrefixTextColor(prompt.DefaultColor))
	}

	return prompt.New(
		exec,
		promptOptions...,
	)
}
