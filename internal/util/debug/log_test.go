//nolint:errcheck
package debug

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	// Setup test environment
	os.Setenv(envEnableLog, "true")
	os.Setenv(envLogFormat, FormatJSON)
	defer os.Unsetenv(envEnableLog)
	defer os.Unsetenv(envLogFormat)

	// Clean up any existing counter file
	os.Remove(counterFileName)
	defer os.Remove(counterFileName)

	t.Run("basic logging", func(t *testing.T) {
		logger, err := NewLogger()
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		testData := map[string]interface{}{
			"test":  "value",
			"count": 42,
		}

		err = logger.Log(testData)
		if err != nil {
			t.Errorf("Log() failed: %v", err)
		}

		// Verify file was created
		filename := fmt.Sprintf("%s-%04d.%s", defaultFileName, 0, FormatJSON)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Log file was not created: %v", err)
		}
		os.Remove(filename)
	})

	t.Run("concurrent logging", func(t *testing.T) {
		logger, err := NewLogger()
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				err := logger.Log(map[string]interface{}{
					"goroutine": i,
				})
				if err != nil {
					t.Errorf("Log() failed in goroutine %d: %v", i, err)
				}
			}(i)
		}
		wg.Wait()
	})

	t.Run("buffer flushing", func(t *testing.T) {
		logger, err := NewLogger(WithBufferSize(100))
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Write enough data to trigger buffer flush
		for i := 0; i < 10; i++ {
			err := logger.Log(map[string]interface{}{
				"test": strings.Repeat("a", 20),
			})
			if err != nil {
				t.Errorf("Log() failed: %v", err)
			}
		}
	})

	t.Run("close with empty buffer", func(t *testing.T) {
		logger, err := NewLogger()
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		err = logger.Close()
		if err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})

	t.Run("counter file handling", func(t *testing.T) {
		// Test counter increment
		logger1, err := NewLogger()
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		logger1.Close()

		logger2, err := NewLogger()
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		logger2.Close()

		if logger2.counter != 1 {
			t.Errorf("Counter not incremented correctly, got %d want 1", logger2.counter)
		}

		// Test counter file cleanup
		err = os.WriteFile(counterFileName, []byte("0"), 0644)
		if err != nil {
			t.Fatalf("Failed to write counter file: %v", err)
		}

		logger3, err := NewLogger()
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		logger3.Close()

		if _, err := os.Stat(counterFileName); !os.IsNotExist(err) {
			t.Errorf("Counter file not cleaned up when reaching 0")
		}
	})

	t.Run("invalid buffer size", func(t *testing.T) {
		logger, err := NewLogger(WithBufferSize(-1))
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		if logger.bufferSize != defaultBufferSize {
			t.Errorf("Invalid buffer size not handled correctly, got %d want %d",
				logger.bufferSize, defaultBufferSize)
		}
	})
}

func TestTraceUn(t *testing.T) {
	// Setup test environment
	os.Setenv(envEnableLog, "true")
	os.Setenv(envLogFormat, FormatJSON)
	defer os.Unsetenv(envEnableLog)
	defer os.Unsetenv(envLogFormat)

	// Initialize logger
	Initialize()

	funcName := "TestFunction"
	args := []interface{}{"arg1", 123, true}

	// Test Trace
	traceName := Trace(funcName, args...)
	if traceName != funcName {
		t.Errorf("Trace() returned unexpected name: got %v want %v", traceName, funcName)
	}

	// Test Un
	Untrace(funcName)

	// Clean up
	Teardown()
}

func TestLogDisabled(t *testing.T) {
	// Ensure logging is disabled
	os.Unsetenv(envEnableLog)
	Initialize()

	testData := map[string]interface{}{
		"test": "should not be logged",
	}

	err := Log(testData)
	if err != nil {
		t.Errorf("Log() failed when disabled: %v", err)
	}
}

func TestLogFormats(t *testing.T) {
	testCases := []struct {
		format    string
		formatter Formatter
	}{
		{FormatJSON, JSONFormatter{}},
		{FormatYAML, YAMLFormatter{}},
	}

	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			// Setup test environment
			os.Setenv(envEnableLog, "true")
			defer os.Unsetenv(envEnableLog)

			// Create logger with specific formatter
			logger, err := NewLogger(WithFormatter(tc.formatter))
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer logger.Close()

			testData := map[string]interface{}{
				"time":   time.Now(),
				"format": tc.format,
			}

			err = logger.Log(testData)
			if err != nil {
				t.Errorf("Log() failed with format %s: %v", tc.format, err)
			}

			// Verify formatter implementation
			output, err := tc.formatter.Format(testData)
			if err != nil {
				t.Errorf("Formatter failed for %s: %v", tc.format, err)
			}

			if len(output) == 0 {
				t.Errorf("Formatter returned empty output for %s", tc.format)
			}

			// Verify file extension
			ext := tc.formatter.FileExtension()
			if ext != tc.format {
				t.Errorf("Invalid file extension, got %s want %s", ext, tc.format)
			}
		})
	}
}
