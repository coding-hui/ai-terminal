package console

import (
	"github.com/coding-hui/go-prompt"
	"github.com/coding-hui/go-prompt/completer"
)

const suggestLimit = 5

var (
	options = []prompt.Option{
		prompt.OptionPrefixTextColor(prompt.Turquoise),
		prompt.OptionCompletionOnDown(),
		prompt.OptionPreviewSuggestionTextColor(prompt.Fuchsia),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionDescriptionBGColor(prompt.LightGray),
		prompt.OptionDescriptionTextColor(prompt.Black),
		prompt.OptionSelectedSuggestionBGColor(prompt.Black),
		prompt.OptionSelectedSuggestionTextColor(prompt.White),
		prompt.OptionSelectedDescriptionBGColor(prompt.DarkGray),
		prompt.OptionScrollbarThumbColor(prompt.Black),
		prompt.OptionScrollbarBGColor(prompt.White),
		prompt.OptionMaxSuggestion(suggestLimit),
		prompt.OptionSwitchKeyBindMode(prompt.CommonKeyBind),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
		prompt.OptionAddASCIICodeBind(makeConsoleKeyBindings()...),
		prompt.OptionAddKeyBind(makeNecessaryKeyBindings()...),
	}
)

func NewPrompt(prefix string, enableColor bool, completer prompt.Completer, exec func(string)) *prompt.Prompt {
	promptOptions := append(options, prompt.OptionPrefix(prefix+" > "))
	if !enableColor {
		promptOptions = append(promptOptions, prompt.OptionPrefixTextColor(prompt.DefaultColor))
	}

	return prompt.New(exec, completer, promptOptions...)
}
