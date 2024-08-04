package prompt

import (
	"embed"
	"io/fs"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"
	"k8s.io/klog/v2"
)

var (
	//go:embed templates/*
	templatesFS embed.FS

	templatesDir    = "templates"
	promptTemplates map[string]prompts.PromptTemplate
)

// Initializes the prompt package by loading the templates from the embedded file system.
func init() { //nolint:gochecknoinits
	if err := LoadPromptTemplates(templatesFS); err != nil {
		klog.Fatal(err)
	}
}

// LoadPromptTemplates loads all the prompt templates found in the templates directory.
func LoadPromptTemplates(files embed.FS) error {
	if promptTemplates == nil {
		promptTemplates = make(map[string]*prompts.PromptTemplate)
	}

	fs.ReadFile()

	return nil
}
