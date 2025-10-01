package console

import (
	"fmt"
	"os"

	"github.com/elk-language/go-prompt"
	pstrings "github.com/elk-language/go-prompt/strings"
)

var deleteWholeLine = func(p *prompt.Prompt) bool {
	p.DeleteBeforeCursorRunes(pstrings.RuneNumber(len(p.Buffer().Document().Text)))
	return true
}

var bye = func(p *prompt.Prompt) bool {
	fmt.Println("Bye!")
	os.Exit(0)
	return true
}

func makeNecessaryKeyBindings() []prompt.KeyBind {
	keyBinds := []prompt.KeyBind{
		{
			Key: prompt.ControlH,
			Fn:  prompt.DeleteBeforeChar,
		},
		{
			Key: prompt.ControlU,
			Fn:  deleteWholeLine,
		},
		{
			Key: prompt.ControlC,
			Fn:  bye,
		},
	}

	return keyBinds
}
