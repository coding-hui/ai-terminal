package coders

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	StatusLoading StatusType = iota
	StatusInfo
	StatusSuccess
	StatusWarning
	StatusError
	StatusTrace

	StatusLoadingIcon = " ⏳ "
	StatusInfoIcon    = " ℹ️ "
	StatusSuccessIcon = " ✅ "
	StatusErrorIcon   = " ❌ "
	StatusWarningIcon = " ⚠️ "
)

type StatusType int

type Checkpoint struct {
	Done  bool
	Desc  string
	Error error
	Type  StatusType

	time time.Time
}

// Render prints the description of the checkpoint and blinks the text input.
func (s Checkpoint) Render() tea.Cmd {
	return tea.Sequence(
		tea.Println(s.Desc),
		textinput.Blink,
	)
}

// Loading sets the status to loading with the given description.
func (a *AutoCoder) Loading(desc string) {
	a.changeStatus(StatusLoading, desc)
}

// Info sets the status to info with the given description.
func (a *AutoCoder) Info(desc string) {
	a.changeStatus(StatusInfo, desc)
}

// Infof sets the status to info with the formatted description.
func (a *AutoCoder) Infof(format string, args ...interface{}) {
	a.changeStatus(StatusInfo, fmt.Sprintf(format, args...))
}

// Success sets the status to success with the given description.
func (a *AutoCoder) Success(desc string) {
	a.changeStatus(StatusSuccess, desc)
}

// Successf sets the status to success with the formatted description.
func (a *AutoCoder) Successf(format string, args ...interface{}) {
	a.changeStatus(StatusSuccess, fmt.Sprintf(format, args...))
}

// Warning sets the status to warning with the given description.
func (a *AutoCoder) Warning(desc string) {
	a.changeStatus(StatusWarning, desc)
}

// Warningf sets the status to warning with the formatted description.
func (a *AutoCoder) Warningf(format string, args ...interface{}) {
	a.changeStatus(StatusWarning, fmt.Sprintf(format, args...))
}

// Error sets the status to error with the given error message and returns the error.
func (a *AutoCoder) Error(args interface{}) error {
	err := fmt.Errorf("%s", args)
	a.state.querying = false
	a.checkpointChan <- Checkpoint{Type: StatusError, Error: err, Desc: err.Error(), time: time.Now()}
	return err
}

// Errorf sets the status to error with the formatted error message and returns the error.
func (a *AutoCoder) Errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	a.state.querying = false
	a.checkpointChan <- Checkpoint{Type: StatusError, Error: err, Desc: err.Error(), time: time.Now()}
	return err
}

// Trace sets the status to trace with the given description.
func (a *AutoCoder) Trace(desc string) {
	a.changeStatus(StatusTrace, desc)
}

// Tracef sets the status to trace with the formatted description.
func (a *AutoCoder) Tracef(format string, args ...interface{}) {
	a.changeStatus(StatusTrace, fmt.Sprintf(format, args...))
}

// WaitForUserConfirm waits for user confirmation with the formatted message and returns the user's choice.
func (a *AutoCoder) WaitForUserConfirm(format string, args ...interface{}) bool {
	a.state.confirming = true

	components.confirm = NewConfirmModel(fmt.Sprintf(strings.TrimSpace(format), args...))

	program.Send(components.confirm)

	defer func() {
		a.state.confirming = false
	}()

	return <-components.confirm.choice
}

// Done sets the checkpoint to done.
func (a *AutoCoder) Done() {
	a.state.querying = false
	a.checkpointChan <- Checkpoint{Done: true, time: time.Now()}
}

// changeStatus changes the status with the given status type and description.
func (a *AutoCoder) changeStatus(statusType StatusType, desc string) {
	a.state.querying = true
	a.checkpointChan <- Checkpoint{Type: statusType, Desc: desc, time: time.Now()}
}

// statusTickCmd creates a tick command to process checkpoints.
func (a *AutoCoder) statusTickCmd() tea.Cmd {
	return tea.Tick(time.Second/10, func(t time.Time) tea.Msg {
		return <-a.checkpointChan
	})
}

func checkpointIcon(checkpointType StatusType) string {
	switch checkpointType {
	case StatusLoading:
		return StatusLoadingIcon
	case StatusSuccess:
		return StatusSuccessIcon
	case StatusWarning:
		return StatusWarningIcon
	case StatusError:
		return StatusErrorIcon
	case StatusInfo:
		return StatusInfoIcon
	default:
		return " "
	}
}
