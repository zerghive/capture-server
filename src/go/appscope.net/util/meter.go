package util

import (
	"time"
)

type Meter struct {
	name     string
	start    time.Time
	duration time.Duration
}

func NewMeter(name string) *Meter {
	return &Meter{
		name:  name,
		start: time.Now(),
	}
}

func (m *Meter) Stop() {
	m.duration = time.Since(m.start)
}

func (m *Meter) Elapsed() string {
	return m.duration.String()
}
