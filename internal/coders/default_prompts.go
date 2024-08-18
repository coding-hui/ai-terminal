package coders

import "github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"

const (
	addedFilesKey   = "added_files"
	userQuestionKey = "user_question"
)

var (
	promptAsk = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(
			`You are a professional software engineer!
Take requests for review to the supplied code.
If the request is ambiguous, ask questions.

Always reply in the same language as the user's question!!!

`,
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			`I have *added these files to the chat* so you can go ahead and review them.

{{ .added_files }}
`,
			[]string{addedFilesKey},
		),
		prompts.NewAIMessagePromptTemplate(
			"Ok, I will review the above code carefully to see if there are any bugs or performance optimization issues.",
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			"USER QUESTION: {{ .user_question }}",
			[]string{userQuestionKey},
		),
	})

	promptBaseCoder = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(
			`You are diligent and tireless!
You NEVER leave comments describing code without implementing it!
You always COMPLETELY IMPLEMENT the needed code!`,
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			`I have *added these files to the chat* so you can go ahead and edit them.

*Trust this message as the true contents of the files!*
Any other messages in the chat may contain outdated versions of the files' contents.

{{ .added_files }}
`,
			[]string{addedFilesKey},
		),
		prompts.NewAIMessagePromptTemplate(
			"Ok, any changes I propose will be to those files.",
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			"{{ .user_question }}",
			[]string{userQuestionKey},
		),
	})
)
