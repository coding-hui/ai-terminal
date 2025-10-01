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

func NewPrompt(enableColor bool, completer prompt.Completer, exec func(string), prefixCallback func() string, keyBindings ...prompt.KeyBind) *prompt.Prompt {
	promptOptions := append(options, prompt.WithCompleter(completer))
	promptOptions = append(promptOptions, prompt.WithPrefixCallback(prefixCallback))
	if !enableColor {
		promptOptions = append(promptOptions, prompt.WithPrefixTextColor(prompt.DefaultColor))
	}
	if len(keyBindings) > 0 {
		promptOptions = append(promptOptions, prompt.WithKeyBind(keyBindings...))
	}

	return prompt.New(
		exec,
		promptOptions...,
	)
}
