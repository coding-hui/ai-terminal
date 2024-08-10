package coders

import (
	"fmt"
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
)

type StatusType int

type Checkpoint struct {
	Done  bool
	Desc  string
	Error error
	Type  StatusType

	time time.Time
}

func (s Checkpoint) Render() tea.Cmd {
	return tea.Sequence(
		tea.Println(s.Desc),
		textinput.Blink,
	)
}

func (a *AutoCoder) Loading(desc string) {
	a.changeStatus(StatusLoading, desc)
}

func (a *AutoCoder) Info(desc string) {
	a.changeStatus(StatusInfo, desc)
}

func (a *AutoCoder) Infof(format string, args ...interface{}) {
	a.changeStatus(StatusInfo, fmt.Sprintf(format, args...))
}

func (a *AutoCoder) Success(desc string) {
	a.changeStatus(StatusSuccess, desc)
}

func (a *AutoCoder) Successf(format string, args ...interface{}) {
	a.changeStatus(StatusSuccess, fmt.Sprintf(format, args...))
}

func (a *AutoCoder) Warning(desc string) {
	a.changeStatus(StatusWarning, desc)
}

func (a *AutoCoder) Warningf(format string, args ...interface{}) {
	a.changeStatus(StatusWarning, fmt.Sprintf(format, args...))
}

func (a *AutoCoder) Error(args interface{}) error {
	err := fmt.Errorf("%s", args)
	a.checkpointChan <- Checkpoint{Type: StatusError, Error: err, Desc: err.Error(), time: time.Now()}
	return err
}

func (a *AutoCoder) Errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	a.checkpointChan <- Checkpoint{Type: StatusError, Error: err, Desc: err.Error(), time: time.Now()}
	return err
}

func (a *AutoCoder) Done() {
	a.checkpointChan <- Checkpoint{Done: true, time: time.Now()}
}

func (a *AutoCoder) changeStatus(statusType StatusType, desc string) {
	a.checkpointChan <- Checkpoint{Type: statusType, Desc: desc, time: time.Now()}
}

func (a *AutoCoder) statusTickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return <-a.checkpointChan
	})
}
