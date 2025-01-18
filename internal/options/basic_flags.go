package options

import (
	"github.com/spf13/pflag"

	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

const (
	FlagLogFlushFrequency = "log-flush-frequency"
)

// AddBasicFlags binds client configuration flags to a given flagset.
func AddBasicFlags(flags *pflag.FlagSet, cfg *Config) {
	flags.StringVarP(&cfg.Model, "model", "m", cfg.Model, console.StdoutStyles().FlagDesc.Render(Help["model"]))
	flags.StringVarP(&cfg.API, "api", "a", cfg.API, console.StdoutStyles().FlagDesc.Render(Help["api"]))
	//flags.StringVarP(&cfg.HTTPProxy, "http-proxy", "x", cfg.HTTPProxy, console.StdoutStyles().FlagDesc.Render(Help["http-proxy"]))
	//flags.BoolVarP(&cfg.Format, "format", "f", cfg.Format, console.StdoutStyles().FlagDesc.Render(Help["format"]))
	flags.StringVar(&cfg.FormatAs, "format-as", cfg.FormatAs, console.StdoutStyles().FlagDesc.Render(Help["format-as"]))
	flags.BoolVarP(&cfg.Raw, "raw", "r", cfg.Raw, console.StdoutStyles().FlagDesc.Render(Help["raw"]))
	flags.BoolVarP(&cfg.Quiet, "quiet", "q", cfg.Quiet, console.StdoutStyles().FlagDesc.Render(Help["quiet"]))
	flags.IntVar(&cfg.MaxRetries, "max-retries", cfg.MaxRetries, console.StdoutStyles().FlagDesc.Render(Help["max-retries"]))
	flags.BoolVar(&cfg.NoLimit, "no-limit", cfg.NoLimit, console.StdoutStyles().FlagDesc.Render(Help["no-limit"]))
	flags.IntVar(&cfg.MaxTokens, "max-tokens", cfg.MaxTokens, console.StdoutStyles().FlagDesc.Render(Help["max-tokens"]))
	flags.IntVar(&cfg.WordWrap, "word-wrap", cfg.WordWrap, console.StdoutStyles().FlagDesc.Render(Help["word-wrap"]))
	flags.Float64Var(&cfg.Temperature, "temp", cfg.Temperature, console.StdoutStyles().FlagDesc.Render(Help["temp"]))
	flags.StringArrayVar(&cfg.Stop, "stop", cfg.Stop, console.StdoutStyles().FlagDesc.Render(Help["stop"]))
	flags.Float64Var(&cfg.TopP, "topp", cfg.TopP, console.StdoutStyles().FlagDesc.Render(Help["topp"]))
	flags.IntVar(&cfg.TopK, "topk", cfg.TopK, console.StdoutStyles().FlagDesc.Render(Help["topk"]))
	flags.UintVar(&cfg.Fanciness, "fanciness", cfg.Fanciness, console.StdoutStyles().FlagDesc.Render(Help["fanciness"]))
	flags.StringVar(&cfg.LoadingText, "loading-text", cfg.LoadingText, console.StdoutStyles().FlagDesc.Render(Help["status-text"]))
	flags.BoolVar(&cfg.NoCache, "no-cache", cfg.NoCache, console.StdoutStyles().FlagDesc.Render(Help["no-cache"]))
	flags.StringVarP(&cfg.Show, "show", "s", cfg.Show, console.StdoutStyles().FlagDesc.Render(Help["show"]))
	flags.BoolVarP(&cfg.ShowLast, "show-last", "S", false, console.StdoutStyles().FlagDesc.Render(Help["show-last"]))
	flags.StringVarP(&cfg.Continue, "continue", "c", "", console.StdoutStyles().FlagDesc.Render(Help["continue"]))
	flags.BoolVarP(&cfg.ContinueLast, "continue-last", "C", false, console.StdoutStyles().FlagDesc.Render(Help["continue-last"]))
	flags.StringVarP(&cfg.Title, "title", "T", cfg.Title, console.StdoutStyles().FlagDesc.Render(Help["title"]))
	flags.IntVarP(&cfg.Verbose, "verbose", "v", cfg.Verbose, console.StdoutStyles().FlagDesc.Render(Help["verbose"]))
	//flags.StringVarP(&cfg.Role, "role", "R", cfg.Role, console.StdoutStyles().FlagDesc.Render(Help["role"]))
	//flags.BoolVar(&cfg.ListRoles, "list-roles", cfg.ListRoles, console.StdoutStyles().FlagDesc.Render(Help["list-roles"]))
	//flags.StringVar(&cfg.Theme, "theme", "charm", console.StdoutStyles().FlagDesc.Render(Help["theme"]))
}
