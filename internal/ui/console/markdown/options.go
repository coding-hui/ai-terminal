package markdown

import (
	"github.com/fatih/color"
)

// Config contains all configurable options for the renderer
type Config struct {
	Theme        string
	CodeTheme    string
	Background   color.Attribute
	HeaderStyle  *color.Color
	ListStyle    *color.Color
	TextEmphasis *color.Color
	RenderMode   string
}

// Option defines a function type for configuring the renderer
type Option func(*Config)

// WithTheme sets the main theme
func WithTheme(theme string) Option {
	return func(c *Config) {
		c.Theme = theme
	}
}

// WithCodeTheme sets the code highlighting theme
func WithCodeTheme(theme string) Option {
	return func(c *Config) {
		c.CodeTheme = theme
	}
}

// WithBackground sets the background color
func WithBackground(bg color.Attribute) Option {
	return func(c *Config) {
		c.Background = bg
	}
}

// WithHeaderStyle sets the header style
func WithHeaderStyle(style *color.Color) Option {
	return func(c *Config) {
		c.HeaderStyle = style
	}
}

// WithListStyle sets the list style
func WithListStyle(style *color.Color) Option {
	return func(c *Config) {
		c.ListStyle = style
	}
}

// WithTextEmphasis sets the text emphasis style
func WithTextEmphasis(style *color.Color) Option {
	return func(c *Config) {
		c.TextEmphasis = style
	}
}

// WithRenderMode sets the rendering mode
func WithRenderMode(mode string) Option {
	return func(c *Config) {
		c.RenderMode = mode
	}
}

// NewConfig creates a new Config with options
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		CodeTheme: "monokai",
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
