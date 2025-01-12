package prompt

import (
	"fmt"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"
)

func GetPromptStringByTemplate(promptTemplate string, vars map[string]any) (llms.PromptValue, error) {
	tpl := prompts.NewPromptTemplate(promptTemplate, nil)
	return tpl.FormatPrompt(vars)
}

func GetPromptStringByTemplateName(templateName string, vars map[string]any) (llms.PromptValue, error) {
	t, ok := promptTemplates[templateName]
	if !ok {
		return nil, fmt.Errorf("prompt template %s not found", templateName)
	}
	res, err := t.template.FormatPrompt(vars)
	if err != nil {
		return nil, err
	}
	return res, nil
}
