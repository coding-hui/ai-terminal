package options

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	_ "embed"

	"github.com/adrg/xdg"
	"github.com/caarlos0/duration"
	"github.com/caarlos0/env/v9"
	"github.com/caarlos0/go-shellwords"
	str "github.com/charmbracelet/x/exp/strings"
	"gopkg.in/yaml.v3"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/system"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

//go:embed config_template.yml
var configTemplate string

const (
	// RecommendedEnvPrefix defines the ENV prefix used by all iam service.
	RecommendedEnvPrefix = "AI_"

	defaultMarkdownFormatText = "Format the response as markdown without enclosing backticks."
	defaultJSONFormatText     = "Format the response as json without enclosing backticks."
)

var Help = map[string]string{
	"api":                 "OpenAI compatible REST API (openai, localai, deepseek).",
	"apis":                "Aliases and endpoints for OpenAI compatible REST API.",
	"http-proxy":          "HTTP proxy to use for API requests.",
	"model":               "Default model (gpt-3.5-turbo, gpt-4, ggml-gpt4all-j...).",
	"ask-model":           "Ask which model to use with an interactive prompt.",
	"max-input-chars":     "Default character limit on input to model.",
	"format":              "Ask for the response to be formatted as markdown unless otherwise set.",
	"format-text":         "Text to append when using the -f flag.",
	"role":                "System role to use.",
	"roles":               "List of predefined system messages that can be used as roles.",
	"list-roles":          "List the roles defined in your configuration file",
	"prompt":              "Include the prompt from the arguments and stdin, truncate stdin to specified number of lines.",
	"prompt-args":         "Include the prompt from the arguments in the response.",
	"raw":                 "Render output as raw text when connected to a TTY.",
	"quiet":               "Quiet mode (hide the spinner while loading and stderr messages for success).",
	"help":                "Show help and exit.",
	"version":             "Show version and exit.",
	"max-retries":         "Maximum number of times to retry API calls.",
	"no-limit":            "Turn off the client-side limit on the size of the input into the model.",
	"word-wrap":           "Wrap formatted output at specific width (default is 80)",
	"max-tokens":          "Maximum number of tokens in response.",
	"temp":                "Temperature (randomness) of results, from 0.0 to 2.0.",
	"stop":                "Up to 4 sequences where the API will stop generating further tokens.",
	"topp":                "TopP, an alternative to temperature that narrows response, from 0.0 to 1.0.",
	"topk":                "TopK, only sample from the top K options for each subsequent token.",
	"fanciness":           "Your desired level of fanciness.",
	"loading-text":        "Text to show while generating.",
	"settings":            "Open settings in your $EDITOR.",
	"dirs":                "Print the directories in which mods store its data.",
	"reset-settings":      "Backup your old settings file and reset everything to the defaults.",
	"continue":            "Continue from the last response or a given save title.",
	"continue-last":       "Continue from the last response.",
	"no-cache":            "Disables caching of the prompt/response.",
	"title":               "Saves the current conversation with the given title.",
	"ls-convo":            "Lists saved conversations.",
	"rm-convo":            "Deletes a saved conversation with the given title or ID.",
	"rm-convo-older-than": "Deletes all saved conversations older than the specified duration. Valid units are: " + str.EnglishJoin(duration.ValidUnits(), true) + ".",
	"rm-all-convo":        "Deletes all saved conversations.",
	"show-convo":          "Show a saved conversation with the given title or ID.",
	"theme":               "Theme to use in the forms. Valid units are: 'charm', 'catppuccin', 'dracula', and 'base16'",
	"show-last":           "Show the last saved conversation.",
	"datastore":           "Configure the datastore to use.",
	"auto-coder":          "Configure the auto coder to use.",
	"auto-commit":         "Automatically commit code changes after generation.",
	"verbose":             "Verbose mode. 0: no verbose, 1: debug verbose",
}

// Config is a structure used to configure a AI.
// Its members are sorted roughly in order of importance for composers.
type Config struct {
	Model         string     `yaml:"default-model" env:"MODEL"`
	API           string     `yaml:"default-api" env:"API"`
	Raw           bool       `yaml:"raw" env:"RAW"`
	Quiet         bool       `yaml:"quiet" env:"QUIET"`
	MaxTokens     int        `yaml:"max-tokens" env:"MAX_TOKENS"`
	MaxInputChars int        `yaml:"max-input-chars" env:"MAX_INPUT_CHARS"`
	Temperature   float64    `yaml:"temp" env:"TEMP"`
	Stop          []string   `yaml:"stop" env:"STOP"`
	TopP          float64    `yaml:"topp" env:"TOPP"`
	TopK          int        `yaml:"topk" env:"TOPK"`
	NoLimit       bool       `yaml:"no-limit" env:"NO_LIMIT"`
	NoCache       bool       `yaml:"no-cache" env:"NO_CACHE"`
	MaxRetries    int        `yaml:"max-retries" env:"MAX_RETRIES"`
	WordWrap      int        `yaml:"word-wrap" env:"WORD_WRAP"`
	Fanciness     uint       `yaml:"fanciness" env:"FANCINESS"`
	LoadingText   string     `yaml:"loading-text" env:"LOADING_TEXT"`
	FormatText    FormatText `yaml:"format-text"`
	FormatAs      string     `yaml:"format-as" env:"FORMAT_AS"`
	Verbose       int        `yaml:"verbose" env:"VERBOSE"`
	APIs          APIs       `yaml:"apis"`
	DataStore     DataStore  `yaml:"datastore"`
	AutoCoder     AutoCoder  `yaml:"auto-coder"`

	DefaultPromptMode string `yaml:"default-prompt-mode,omitempty"`
	ConversationID    string `yaml:"convo-id,omitempty"`
	Ai                Ai     `yaml:"ai"`

	CurrentModel Model
	CurrentAPI   API
	Models       map[string]Model
	SettingsPath string
	System       *system.Analysis
	Interactive  bool
	PromptFile   string
	ContinueLast bool
	Continue     string
	Title        string
	Show         string
	ShowLast     bool

	CacheReadFromID, CacheWriteToID, CacheWriteToTitle string
}

// AutoCoder is the configuration for the auto coder.
type AutoCoder struct {
	PromptPrefix string `yaml:"prompt-prefix" env:"PROMPT_PREFIX"`
	EditFormat   string `yaml:"edit-format" env:"EDIT_FORMAT"`
	CommitPrefix string `yaml:"commit-prefix" env:"COMMIT_PREFIX"`
	AutoCommit   bool   `yaml:"auto-commit" env:"AUTO_COMMIT" default:"true"`
	DesignModel  string `yaml:"design-model" env:"DESIGN_MODEL"`
	CodingModel  string `yaml:"coding-model" env:"CODING_MODEL"`
}

// Model represents the LLM model used in the API call.
type Model struct {
	Name     string
	API      string
	MaxChars int      `yaml:"max-input-chars"`
	Aliases  []string `yaml:"aliases"`
	Fallback string   `yaml:"fallback"`
}

// API represents an API endpoint and its models.
type API struct {
	Name       string
	APIKey     string           `yaml:"api-key"`
	APIKeyEnv  string           `yaml:"api-key-env"`
	APIKeyCmd  string           `yaml:"api-key-cmd"`
	Version    string           `yaml:"version"`
	BaseURL    string           `yaml:"base-url"`
	Region     string           `yaml:"region"`
	RetryTimes int              `yaml:"retry-times"`
	Timeout    time.Duration    `yaml:"timeout"`
	Models     map[string]Model `yaml:"models"`
	User       string           `yaml:"user"`
}

// APIs is a type alias to allow custom YAML decoding.
type APIs []API

// UnmarshalYAML implements sorted API YAML decoding.
func (apis *APIs) UnmarshalYAML(node *yaml.Node) error {
	for i := 0; i < len(node.Content); i += 2 {
		var api API
		if err := node.Content[i+1].Decode(&api); err != nil {
			return fmt.Errorf("error decoding YAML file: %s", err)
		}
		api.Name = node.Content[i].Value
		*apis = append(*apis, api)
	}
	return nil
}

// FormatText is a map[format]formatting_text.
type FormatText map[string]string

// UnmarshalYAML conforms with yaml.Unmarshaler.
func (ft *FormatText) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var text string
	if err := unmarshal(&text); err != nil {
		var formats map[string]string
		if err := unmarshal(&formats); err != nil {
			return err
		}
		*ft = formats
		return nil
	}

	*ft = map[string]string{
		"markdown": text,
	}
	return nil
}

// Deprecated: Use Model instead.
type Ai struct {
	SystemPrompt        string       `yaml:"system-prompt,omitempty"`
	Token               string       `yaml:"token,omitempty"`
	Model               string       `yaml:"model,omitempty"`
	ApiBase             string       `yaml:"api-base,omitempty"`
	Temperature         float64      `yaml:"temperature,omitempty"`
	TopP                float64      `yaml:"top-p,omitempty"`
	MaxTokens           int          `yaml:"max-tokens,omitempty"`
	Proxy               string       `yaml:"proxy,omitempty"`
	OutputFormat        OutputFormat `yaml:"output-format,omitempty"`
	MultiContentEnabled bool         `yaml:"multi-content-enabled,omitempty"`
}

type DataStore struct {
	Type      string `yaml:"type,omitempty" env:"DATASTORE_TYPE"`
	CachePath string `yaml:"cache-path" env:"CACHE_PATH"`
	Url       string `yaml:"url,omitempty"`
	Username  string `yaml:"username,omitempty"`
	Password  string `yaml:"password,omitempty"`
}

type OutputFormat string

const (
	RawOutputFormat      OutputFormat = "raw"
	MarkdownOutputFormat OutputFormat = "markdown"
)

func (o OutputFormat) String() string {
	return string(o)
}

func (c *Config) GetModel(name string) (model Model, err error) {
	mod, ok := c.Models[name]
	if !ok {
		if c.API == "" {
			return model, errbook.Wrap(
				fmt.Sprintf(
					"model %s is not in the settings file.",
					console.StderrStyles().InlineCode.Render(c.Model),
				),
				errbook.NewUserErrorf(
					"Please specify an API endpoint with %s or configure the model in the settings: %s",
					console.StderrStyles().InlineCode.Render("--api"),
					console.StderrStyles().InlineCode.Render("ai -s"),
				),
			)
		}
		mod.Name = c.Model
		mod.API = c.API
		mod.MaxChars = c.MaxInputChars
	}

	if c.API != "" {
		mod.API = c.API
	}

	return mod, nil
}

func (c *Config) GetAPI(name string) (api API, err error) {
	for _, a := range c.APIs {
		if name == a.Name {
			api = a
			break
		}
	}

	if api.Name == "" {
		eps := make([]string, 0)
		for _, a := range c.APIs {
			eps = append(eps, console.StderrStyles().InlineCode.Render(a.Name))
		}
		return api, errbook.Wrap(
			fmt.Sprintf(
				"The API endpoint %s is not configured.",
				console.StderrStyles().InlineCode.Render(c.API),
			),
			errbook.NewUserErrorf(
				"Your configured API endpoints are: %s",
				eps,
			),
		)
	}

	api.APIKey, err = ensureApiKey(api)
	if err != nil {
		return api, err
	}

	return api, nil
}

func EnsureConfig() (Config, error) {
	var c Config
	sp, err := xdg.ConfigFile(filepath.Join("ai-terminal", "config.yml"))
	if err != nil {
		return c, errbook.Wrap("Could not find settings path.", err)
	}
	c.SettingsPath = sp

	dir := filepath.Dir(sp)
	if dirErr := os.MkdirAll(dir, 0o700); dirErr != nil {
		return c, errbook.Wrap("Could not create cache directory.", err)
	}

	if dirErr := WriteConfigFile(sp); dirErr != nil {
		return c, dirErr
	}
	content, err := os.ReadFile(sp)
	if err != nil {
		return c, errbook.Wrap("Could not read settings file.", err)
	}
	if err := yaml.Unmarshal(content, &c); err != nil {
		return c, errbook.Wrap("Could not parse settings file.", err)
	}
	ms := make(map[string]Model)
	for _, api := range c.APIs {
		for mk, mv := range api.Models {
			mv.Name = mk
			mv.API = api.Name
			// only set the model key and aliases if they haven't already been used
			_, ok := ms[mk]
			if !ok {
				ms[mk] = mv
			}
			for _, a := range mv.Aliases {
				_, ok := ms[a]
				if !ok {
					ms[a] = mv
				}
			}
		}
	}
	c.Models = ms

	if err := env.ParseWithOptions(&c, env.Options{Prefix: RecommendedEnvPrefix}); err != nil {
		return c, errbook.Wrap("Could not parse environment into settings file.", err)
	}

	if c.DataStore.CachePath == "" {
		c.DataStore.CachePath = filepath.Join(xdg.DataHome, "ai-terminal", "cache")
	}

	if err := os.MkdirAll(c.DataStore.CachePath, 0o700); err != nil { //nolint:mnd
		return c, errbook.Wrap("Could not create cache directory.", err)
	}

	if c.WordWrap == 0 {
		c.WordWrap = 80
	}

	c.CurrentModel, err = c.GetModel(c.Model)
	if err != nil {
		return c, err
	}

	c.CurrentAPI, err = c.GetAPI(c.API)
	if err != nil {
		return c, err
	}

	return c, nil
}

// DefaultConfig returns a Config struct with the default values.
func DefaultConfig() Config {
	return Config{
		FormatAs: "markdown",
		FormatText: FormatText{
			"markdown": defaultMarkdownFormatText,
			"json":     defaultJSONFormatText,
		},
	}
}

func WriteConfigFile(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return createConfigFile(path)
	} else if err != nil {
		return errbook.Wrap("Could not stat path.", err)
	}
	return nil
}

func createConfigFile(path string) error {
	tmpl := template.Must(template.New("config").Parse(configTemplate))

	f, err := os.Create(path)
	if err != nil {
		return errbook.Wrap("Could not create configuration file.", err)
	}
	defer func() { _ = f.Close() }()

	m := struct {
		Config Config
		Help   map[string]string
	}{
		Config: DefaultConfig(),
		Help:   Help,
	}
	if err := tmpl.Execute(f, m); err != nil {
		return errbook.Wrap("Could not render template.", err)
	}
	return nil
}

func ensureApiKey(api API) (string, error) {
	key := api.APIKey
	if key == "" && api.APIKeyEnv != "" && api.APIKeyCmd == "" {
		key = os.Getenv(api.APIKeyEnv)
	}
	if key == "" && api.APIKeyCmd != "" {
		args, err := shellwords.Parse(api.APIKeyCmd)
		if err != nil {
			return "", errbook.Wrap("Failed to parse api-key-cmd", err)
		}
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput() //nolint:gosec
		if err != nil {
			return "", errbook.Wrap("Cannot exec api-key-cmd", err)
		}
		key = strings.TrimSpace(string(out))
	}
	if key != "" {
		return key, nil
	}
	return "", errbook.Wrap(
		fmt.Sprintf(
			"%[1]s required; set the environment variable %[1]s or update %[2]s through %[3]s.",
			console.StderrStyles().InlineCode.Render(api.APIKeyEnv),
			console.StderrStyles().InlineCode.Render("Config.yaml"),
			console.StderrStyles().InlineCode.Render("ai Config"),
		),
		errbook.NewUserErrorf(
			"You can grab one at %s.",
			console.StderrStyles().Link.Render(api.BaseURL),
		),
	)
}
