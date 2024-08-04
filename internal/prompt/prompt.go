package prompt

import (
	"fmt"
)

func GetPromptStringByTemplateName(templateName string, vars map[string]any) (string, error) {
	t, ok := promptTemplates[templateName]
	if !ok {
		return "", fmt.Errorf("prompt template %s not found", templateName)
	}
	res, err := t.template.Format(vars)
	if err != nil {
		return "", err
	}
	return res, nil
}
