package options

import (
	"github.com/spf13/pflag"

	"github.com/coding-hui/ai-terminal/internal/ui/display"
)

const (
	FlagLogFlushFrequency = "log-flush-frequency"
)

// AddBasicFlags binds client configuration flags to a given flagset.
func AddBasicFlags(flags *pflag.FlagSet, cfg *Config) {
	flags.StringVarP(&cfg.Model, "model", "m", cfg.Model, display.StdoutStyles().FlagDesc.Render(help["model"]))
	flags.StringVarP(&cfg.API, "api", "a", cfg.API, display.StdoutStyles().FlagDesc.Render(help["api"]))
	//flags.StringVarP(&cfg.HTTPProxy, "http-proxy", "x", cfg.HTTPProxy, display.StdoutStyles().FlagDesc.Render(help["http-proxy"]))
	//flags.BoolVarP(&cfg.Format, "format", "f", cfg.Format, display.StdoutStyles().FlagDesc.Render(help["format"]))
	flags.StringVar(&cfg.FormatAs, "format-as", cfg.FormatAs, display.StdoutStyles().FlagDesc.Render(help["format-as"]))
	flags.BoolVarP(&cfg.Raw, "raw", "r", cfg.Raw, display.StdoutStyles().FlagDesc.Render(help["raw"]))
	//flags.IntVarP(&cfg.IncludePrompt, "prompt", "P", cfg.IncludePrompt, display.StdoutStyles().FlagDesc.Render(help["prompt"]))
	//flags.BoolVarP(&cfg.IncludePromptArgs, "prompt-args", "p", cfg.IncludePromptArgs, display.StdoutStyles().FlagDesc.Render(help["prompt-args"]))
	//flags.StringVarP(&cfg.Continue, "continue", "c", "", display.StdoutStyles().FlagDesc.Render(help["continue"]))
	//flags.BoolVarP(&cfg.ContinueLast, "continue-last", "C", false, display.StdoutStyles().FlagDesc.Render(help["continue-last"]))
	//flags.BoolVarP(&cfg.List, "list", "l", cfg.List, display.StdoutStyles().FlagDesc.Render(help["list"]))
	//flags.StringVarP(&cfg.Title, "title", "t", cfg.Title, display.StdoutStyles().FlagDesc.Render(help["title"]))
	//flags.StringVarP(&cfg.Delete, "delete", "d", cfg.Delete, display.StdoutStyles().FlagDesc.Render(help["delete"]))
	//flags.Var(newDurationFlag(cfg.DeleteOlderThan, &cfg.DeleteOlderThan), "delete-older-than", display.StdoutStyles().FlagDesc.Render(help["delete-older-than"]))
	//flags.StringVarP(&cfg.Show, "show", "s", cfg.Show, display.StdoutStyles().FlagDesc.Render(help["show"]))
	//flags.BoolVarP(&cfg.ShowLast, "show-last", "S", false, display.StdoutStyles().FlagDesc.Render(help["show-last"]))
	flags.BoolVarP(&cfg.Quiet, "quiet", "q", cfg.Quiet, display.StdoutStyles().FlagDesc.Render(help["quiet"]))
	//flags.BoolVarP(&cfg.ShowHelp, "help", "h", false, display.StdoutStyles().FlagDesc.Render(help["help"]))
	//flags.BoolVarP(&cfg.Version, "version", "v", false, display.StdoutStyles().FlagDesc.Render(help["version"]))
	flags.IntVar(&cfg.MaxRetries, "max-retries", cfg.MaxRetries, display.StdoutStyles().FlagDesc.Render(help["max-retries"]))
	flags.BoolVar(&cfg.NoLimit, "no-limit", cfg.NoLimit, display.StdoutStyles().FlagDesc.Render(help["no-limit"]))
	flags.IntVar(&cfg.MaxTokens, "max-tokens", cfg.MaxTokens, display.StdoutStyles().FlagDesc.Render(help["max-tokens"]))
	flags.IntVar(&cfg.WordWrap, "word-wrap", cfg.WordWrap, display.StdoutStyles().FlagDesc.Render(help["word-wrap"]))
	flags.Float64Var(&cfg.Temperature, "temp", cfg.Temperature, display.StdoutStyles().FlagDesc.Render(help["temp"]))
	flags.StringArrayVar(&cfg.Stop, "stop", cfg.Stop, display.StdoutStyles().FlagDesc.Render(help["stop"]))
	flags.Float64Var(&cfg.TopP, "topp", cfg.TopP, display.StdoutStyles().FlagDesc.Render(help["topp"]))
	flags.IntVar(&cfg.TopK, "topk", cfg.TopK, display.StdoutStyles().FlagDesc.Render(help["topk"]))
	flags.UintVar(&cfg.Fanciness, "fanciness", cfg.Fanciness, display.StdoutStyles().FlagDesc.Render(help["fanciness"]))
	flags.StringVar(&cfg.LoadingText, "loading-text", cfg.LoadingText, display.StdoutStyles().FlagDesc.Render(help["status-text"]))
	flags.BoolVar(&cfg.NoCache, "no-cache", cfg.NoCache, display.StdoutStyles().FlagDesc.Render(help["no-cache"]))
	//flags.BoolVar(&cfg.ResetSettings, "reset-settings", cfg.ResetSettings, display.StdoutStyles().FlagDesc.Render(help["reset-settings"]))
	//flags.BoolVar(&cfg.Settings, "settings", false, display.StdoutStyles().FlagDesc.Render(help["settings"]))
	//flags.BoolVar(&cfg.Dirs, "dirs", false, display.StdoutStyles().FlagDesc.Render(help["dirs"]))
	//flags.StringVarP(&cfg.Role, "role", "R", cfg.Role, display.StdoutStyles().FlagDesc.Render(help["role"]))
	//flags.BoolVar(&cfg.ListRoles, "list-roles", cfg.ListRoles, display.StdoutStyles().FlagDesc.Render(help["list-roles"]))
	//flags.StringVar(&cfg.Theme, "theme", "charm", display.StdoutStyles().FlagDesc.Render(help["theme"]))
}
