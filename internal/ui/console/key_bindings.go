package console

import (
	"fmt"
	"os"

	"github.com/elk-language/go-prompt"
	pstrings "github.com/elk-language/go-prompt/strings"
)

var (
	altD = []byte{226, 136, 130}
)

var deleteWholeLine = prompt.ASCIICodeBind{
	ASCIICode: altD,
	Fn: func(p *prompt.Prompt) bool {
		p.DeleteBeforeCursorRunes(pstrings.RuneNumber(len(p.Buffer().Document().Text)))
		return true
	},
}

func makeNecessaryKeyBindings() []prompt.KeyBind {
	keyBinds := []prompt.KeyBind{
		{
			Key: prompt.ControlH,
			Fn:  prompt.DeleteBeforeChar,
		},
		{
			Key: prompt.ControlU,
			Fn:  deleteWholeLine.Fn,
		},
		{
			Key: prompt.ControlC,
			Fn: func(b *prompt.Prompt) bool {
				fmt.Println("Bye!")
				os.Exit(0)
				return true
			},
		},
	}

	return keyBinds
}
