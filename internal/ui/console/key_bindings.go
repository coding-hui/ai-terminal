package console

import (
	"fmt"
	"os"

	"github.com/coding-hui/go-prompt"
)

var (
	altBackspace = []byte{27, 127}
	altLeft      = []byte{27, 98}
	altRight     = []byte{27, 102}
	altUp        = []byte{27, 27, 91, 65}
	altDown      = []byte{27, 27, 91, 66}
	altD         = []byte{226, 136, 130}
)

func makeConsoleKeyBindings() []prompt.ASCIICodeBind {
	return []prompt.ASCIICodeBind{deletePreviousWord, moveToPreviousWord, moveToNextWord, moveToLineBeginning, moveToLineEnd, deleteWholeLine}
}

var deletePreviousWord = prompt.ASCIICodeBind{
	ASCIICode: altBackspace,
	Fn: func(buffer *prompt.Buffer) {
		str := buffer.Document().Text
		cursorPosition := buffer.Document().CursorPositionCol()
		str = str[:cursorPosition]
		buffer.DeleteBeforeCursor(lengthLastWord(str))
	},
}

var moveToPreviousWord = prompt.ASCIICodeBind{
	ASCIICode: altLeft,
	Fn: func(buffer *prompt.Buffer) {
		str := buffer.Document().Text
		cursorPosition := buffer.Document().CursorPositionCol()
		str = str[:cursorPosition]
		buffer.CursorLeft(lengthLastWord(str))
	},
}

var moveToNextWord = prompt.ASCIICodeBind{
	ASCIICode: altRight,
	Fn: func(buffer *prompt.Buffer) {
		str := buffer.Document().Text
		cursorPosition := buffer.Document().CursorPositionCol()
		str = str[cursorPosition:]
		buffer.CursorRight(lengthFirstWord(str))
	},
}

var moveToLineBeginning = prompt.ASCIICodeBind{
	ASCIICode: altUp,
	Fn: func(buffer *prompt.Buffer) {
		str := buffer.Document().Text
		cursorPosition := buffer.Document().CursorPositionCol()
		str = str[:cursorPosition]
		buffer.CursorLeft(len(str))
	},
}

var moveToLineEnd = prompt.ASCIICodeBind{
	ASCIICode: altDown,
	Fn: func(buffer *prompt.Buffer) {
		str := buffer.Document().Text
		cursorPosition := buffer.Document().CursorPositionCol()
		str = str[cursorPosition:]
		buffer.CursorRight(len(str))
	},
}

var deleteWholeLine = prompt.ASCIICodeBind{
	ASCIICode: altD,
	Fn: func(buffer *prompt.Buffer) {
		str := buffer.Document().Text
		cursorPosition := buffer.Document().CursorPositionCol()
		str = str[:cursorPosition]
		buffer.DeleteBeforeCursor(len(str))
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
			Fn: func(b *prompt.Buffer) {
				fmt.Println("Bye!")
				os.Exit(0)
			},
		},
	}

	return keyBinds
}
