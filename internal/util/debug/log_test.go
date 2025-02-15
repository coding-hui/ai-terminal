//nolint:errcheck
package debug

import (
	"os"
	"testing"
)

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
