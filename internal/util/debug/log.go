package debug

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	FormatJSON = "json"
	FormatYAML = "yaml"
)

const (
	defaultFileName   = "ai-terminal"
	counterFileName   = "log_counter.txt"
	envEnableLog      = "AI_TERMINAL_ENABLE_LOG"
	envLogFormat      = "AI_TERMINAL_LOG_FORMAT"
	envLogBufferSize  = "AI_TERMINAL_LOG_BUFFER_SIZE"
	defaultBufferSize = 4096
)

var (
	// Protects global logger instance
	globalLogger     *Logger
	globalLoggerOnce sync.Once
)

// Formatter defines the log formatter interface
type Formatter interface {
	Format(interface{}) ([]byte, error)
	FileExtension() string
}

// JSONFormatter implements JSON format
type JSONFormatter struct{}

func (f JSONFormatter) Format(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func (f JSONFormatter) FileExtension() string {
	return FormatJSON
}

// YAMLFormatter implements YAML format
type YAMLFormatter struct{}

func (f YAMLFormatter) Format(data interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}

func (f YAMLFormatter) FileExtension() string {
	return FormatYAML
}

// Logger core logging structure
type Logger struct {
	mu         sync.Mutex
	formatter  Formatter
	writer     *bufio.Writer
	file       *os.File
	counter    int
	bufferSize int
}

// NewLogger creates a new Logger instance
func NewLogger(options ...Option) (*Logger, error) {
	logger := &Logger{}

	// Apply options
	for _, option := range options {
		if err := option(logger); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Initialize logger
	if err := logger.initialize(); err != nil {
		if errors.Is(err, errors.New("logging disabled")) {
			return logger, nil
		}
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger, nil
}

// Teardown to close logfile and clean counter if empty

// initialize actual initialization logic
func (l *Logger) initialize() error {
	if os.Getenv(envEnableLog) != "true" && os.Getenv(envEnableLog) != "1" {
		return errors.New("logging disabled")
	}

	// Initialize formatter
	if l.formatter == nil {
		switch os.Getenv(envLogFormat) {
		case FormatYAML:
			l.formatter = YAMLFormatter{}
		default:
			l.formatter = JSONFormatter{}
		}
	}

	// Read/initialize counter
	if data, err := os.ReadFile(counterFileName); err == nil {
		_, _ = fmt.Sscanf(string(data), "%d", &l.counter)
	}

	// Create log file
	filename := fmt.Sprintf("%s-%04d.%s", defaultFileName, l.counter, l.formatter.FileExtension())

	var err error
	l.file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return fmt.Errorf("error creating log file: %w", err)
	}

	// Initialize buffered writer
	if l.bufferSize <= 0 {
		l.bufferSize = defaultBufferSize
	}
	l.writer = bufio.NewWriterSize(l.file, l.bufferSize)

	// Increment and save counter
	l.counter++
	if err := os.WriteFile(counterFileName, []byte(fmt.Sprintf("%d", l.counter)), 0644); err != nil {
		return fmt.Errorf("error saving counter: %w", err)
	}

	return nil
}

// Close closes the log file and cleans up resources
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return nil
	}

	// Flush buffer
	if err := l.writer.Flush(); err != nil {
		return err
	}

	// Close file
	if err := l.file.Close(); err != nil {
		return err
	}

	// Clean up counter file
	if data, err := os.ReadFile(counterFileName); err == nil && string(data) == "0" {
		_ = os.Remove(counterFileName)
	}

	return nil
}

// Log writes structured log
func (l *Logger) Log(data interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer == nil {
		return errors.New("logger not initialized")
	}

	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"data":      data,
	}

	output, err := l.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("format error: %w", err)
	}

	if _, err := l.writer.Write(output); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	// Add format separator
	if _, ok := l.formatter.(YAMLFormatter); ok {
		if _, err := l.writer.WriteString("\n---\n"); err != nil {
			return err
		}
	}

	// Flush if buffer is half full
	if l.writer.Buffered() >= l.bufferSize/2 {
		return l.writer.Flush()
	}

	return nil
}

// Trace logs function entry
func (l *Logger) Trace(funcName string, args ...interface{}) string {
	_ = l.Log(map[string]interface{}{
		"type":      "function_entry",
		"function":  funcName,
		"arguments": args,
	})
	return funcName
}

// Untrace logs function exit
func (l *Logger) Untrace(funcName string) {
	_ = l.Log(map[string]interface{}{
		"type":     "function_exit",
		"function": funcName,
	})
}

/************************ Global Functions ************************/

// Initialize global logger
func Initialize() {
	globalLoggerOnce.Do(func() {
		l, err := NewLogger()
		if err == nil {
			globalLogger = l
		}
	})
}

// Teardown global logger
func Teardown() {
	if globalLogger != nil {
		_ = globalLogger.Close()
		globalLogger = nil
	}
}

// Log global logging function
func Log(data interface{}) error {
	if globalLogger == nil {
		return errors.New("logger not initialized")
	}
	return globalLogger.Log(data)
}

// Trace global trace function
func Trace(funcName string, args ...interface{}) string {
	if globalLogger != nil {
		return globalLogger.Trace(funcName, args...)
	}
	return funcName
}

// Untrace global untrace function
func Untrace(funcName string) {
	if globalLogger != nil {
		globalLogger.Untrace(funcName)
	}
}
