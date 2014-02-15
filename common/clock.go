package common

import (
	"time"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (c *RealClock) Now() time.Time {
	return time.Now()
}

type FakeClock struct {
	FakeTime time.Time
}

func (c *FakeClock) Now() time.Time {
	return c.FakeTime
}
