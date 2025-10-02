package coders

import (
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"
)

const (
	addedFilesKey   = "added_files"
	userQuestionKey = "user_question"
	lazyPromptKey   = "lazy_prompt"
	openFenceKey    = "open_fence"
	closeFenceKey   = "close_fence"

	lazyPrompt = `You are diligent and tireless!
You NEVER leave comments describing code without implementing it!
You always COMPLETELY IMPLEMENT the needed code!
`

	systemReminderPrompt = `# *SEARCH/REPLACE block* Rules:

Every *SEARCH/REPLACE block* must use this format:
1. The file path alone on a line, verbatim. No bold asterisks, no quotes around it, no escaping of characters, etc.
2. The opening fence and code language, eg: {{ .open_fence }}python
3. The start of search block: <<<<<<< SEARCH
4. A contiguous chunk of lines to search for in the existing source code
5. The dividing line: =======
6. The lines to replace into the source code
7. The end of the replace block: >>>>>>> REPLACE
8. The closing fence: {{ .close_fence }}

Every *SEARCH* section must *EXACTLY MATCH* the existing source code, character for character, including all comments, docstrings, etc.


*SEARCH/REPLACE* blocks will replace *all* matching occurrences.
Include enough lines to make the SEARCH blocks uniquely match the lines to change.

Keep *SEARCH/REPLACE* blocks concise.
Break large *SEARCH/REPLACE* blocks into a series of smaller blocks that each change a small portion of the file.
Include just the changing lines, and a few surrounding lines if needed for uniqueness.
Do not include long runs of unchanging lines in *SEARCH/REPLACE* blocks.

Only create *SEARCH/REPLACE* blocks for files that the user has added to the chat!

To move code within a file, use 2 *SEARCH/REPLACE* blocks: 1 to delete it from its current location, 1 to insert it in the new location.

If you want to put code in a new file, use a *SEARCH/REPLACE block* with:
- A new file path, including dir name if needed
- An empty SEARCH section
- The new file's contents in the REPLACE section

ONLY EVER RETURN CODE IN A *SEARCH/REPLACE BLOCK*!
`
)

var (
	promptDesign = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(
			`Act as an expert architect engineer and provide direction to your editor engineer.
Study the change request and the current code.
Describe how to modify the code to complete the request.
The editor engineer will rely solely on your instructions, so make them unambiguous and complete.
Explain all needed code changes clearly and completely, but concisely.
Just show the changes needed.

DO NOT show the entire updated function/file/etc!
`,
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			`I have *added these files to the chat* so you see all of their contents.
*Trust this message as the true contents of the files!*
Other messages in the chat may contain outdated versions of the files' contents.

{{ .added_files }}
`,
			[]string{addedFilesKey},
		),
		prompts.NewAIMessagePromptTemplate(
			"Ok, I will use that as the true, current contents of the files.",
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			"USER QUESTION: {{ .user_question }}",
			[]string{userQuestionKey},
		),
	})

	promptAskWithFiles = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(
			`You are a professional software engineer!
Take requests for review to the supplied code.
If the request is ambiguous, ask questions.
Always respond in the same language as the user's question.`,
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			`I have *added these files to the chat* so you can go ahead and review them.

{{ .added_files }}`,
			[]string{addedFilesKey},
		),
		prompts.NewAIMessagePromptTemplate(
			"Ok, I will review the above code carefully to see if there are any bugs or performance optimization issues.",
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			"{{ .user_question }}",
			[]string{userQuestionKey},
		),
	})

	promptAskGeneral = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(
			`You are a helpful AI assistant that can help with various tasks.
Always respond in the same language as the user's question.`,
			nil,
		),
		prompts.NewHumanMessagePromptTemplate(
			"{{ .user_question }}",
			[]string{userQuestionKey},
		),
	})

	promptBaseCoder = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(
			`Act as an expert software developer.
Always use best practices when coding.
Respect and use existing conventions, libraries, etc that are already present in the code base.
{{ .lazy_prompt }}
Take requests for changes to the supplied code.
If the request is ambiguous, ask questions.

Always reply to the user in the same language they are using.

Once you understand the request you MUST:
1. Decide if you need to propose *SEARCH/REPLACE* edits to any files that haven't been added to the chat. You can create new files without asking. But if you need to propose edits to existing files not already added to the chat, you *MUST* tell the user their full path names and ask them to *add the files to the chat*. End your reply and wait for their approval. You can keep asking if you then decide you need to edit more files.
2. Think step-by-step and explain the needed changes with a numbered list of short sentences.
3. Describe each change with a *SEARCH/REPLACE block* per the examples below. All changes to files must use this *SEARCH/REPLACE block* format. ONLY EVER RETURN CODE IN A *SEARCH/REPLACE BLOCK*!

All changes to files must use the *SEARCH/REPLACE block* format.`+systemReminderPrompt,
			[]string{lazyPromptKey, openFenceKey, closeFenceKey},
		),
		prompts.NewHumanMessagePromptTemplate(
			`Change get_factorial() to use math.factorial`,
			nil,
		),
		prompts.NewAIMessagePromptTemplate(
			`To make this change we need to modify **mathweb/flask/app.py** to:

1. Import the math package.
2. Remove the existing factorial() function.
3. Update get_factorial() to call math.factorial instead.

Here are the *SEARCH/REPLACE* blocks:

{{ .open_fence }}python
mathweb/flask/app.py
<<<<<<< SEARCH
from flask import Flask
=======
import math
from flask import Flask
>>>>>>> REPLACE
{{ .close_fence }}

{{ .open_fence }}python
mathweb/flask/app.py
<<<<<<< SEARCH
def factorial(n):
    "compute factorial"

    if n == 0:
        return 1
    else:
        return n * factorial(n-1)

=======
>>>>>>> REPLACE
{{ .close_fence }}

{{ .open_fence }}python
mathweb/flask/app.py
<<<<<<< SEARCH
    return str(factorial(n))
=======
    return str(math.factorial(n))
>>>>>>> REPLACE
{{ .close_fence }}`,
			[]string{openFenceKey, closeFenceKey},
		),
		prompts.NewHumanMessagePromptTemplate(
			`Refactor hello() into its own file.`,
			nil,
		),
		prompts.NewAIMessagePromptTemplate(
			`To make this change we need to modify **main.py** and make a new file **hello.py**:

1. Make a new hello.py file with hello() in it.
2. Remove hello() from main.py and replace it with an import.

Here are the *SEARCH/REPLACE* blocks:

{{ .open_fence }}python
hello.py
<<<<<<< SEARCH
=======
def hello():
    "print a greeting"

    print("hello")
>>>>>>> REPLACE
{{ .close_fence }}

{{ .open_fence }}python
main.py
<<<<<<< SEARCH
def hello():
    "print a greeting"

    print("hello")
=======
from hello import hello
>>>>>>> REPLACE
{{ .close_fence }}`,
			[]string{openFenceKey, closeFenceKey},
		),
		prompts.NewHumanMessagePromptTemplate(
			`I switched to a new code base. Please don't consider the above files or try to edit them any longer.`,
			nil,
		),
		prompts.NewAIMessagePromptTemplate(
			"OK.",
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

func formatPrompt(promptTemplate prompts.ChatPromptTemplate, values map[string]any) ([]llms.ChatMessage, error) {
	messages, err := promptTemplate.FormatMessages(values)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
