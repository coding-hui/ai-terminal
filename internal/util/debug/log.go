package debug

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	callDepth    = 2
	envEnableLog = "GO_PROMPT_ENABLE_LOG"
	logFileName  = "ai-terminal.log"
)

var (
	logfile *os.File
	logger  *log.Logger
)

func init() {
	if e := os.Getenv(envEnableLog); e == "true" || e == "1" {
		var err error
		logfile, err = os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			logger = log.New(logfile, "", log.Llongfile)
			return
		}
	}
	logger = log.New(io.Discard, "", log.Llongfile)
}

// Teardown to close logfile
func Teardown() {
	if logfile == nil {
		return
	}
	_ = logfile.Close()
}

func writeWithSync(msg string) {
	if logfile == nil {
		return
	}
	_ = logger.Output(callDepth, msg)
	_ = logfile.Sync() // immediately write msg
}

// Log to output message
func Log(msg string) {
	writeWithSync(msg)
}

func Trace(s string, args ...interface{}) string {
	writeWithSync(fmt.Sprintln("\n\n------------entering:", s, args))
	return s
}

func Un(s string) {
	writeWithSync(fmt.Sprintln("------------ leaving:", s))
}
