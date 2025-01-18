package flag

import (
	"time"

	"github.com/caarlos0/duration"
)

func NewDurationFlag(val time.Duration, p *time.Duration) *DurationFlag {
	*p = val
	return (*DurationFlag)(p)
}

type DurationFlag time.Duration

func (d *DurationFlag) Set(s string) error {
	v, err := duration.Parse(s)
	*d = DurationFlag(v)
	//nolint: wrapcheck
	return err
}

func (d *DurationFlag) String() string {
	return time.Duration(*d).String()
}

func (*DurationFlag) Type() string {
	return "duration"
}
