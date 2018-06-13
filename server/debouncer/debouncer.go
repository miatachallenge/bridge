package debouncer

import (
	"sync"
	"time"
)

// Debouncer is a channel debouncer that triggers its output after the passed
// duration passes without any incoming events.
type Debouncer struct {
	sync.Mutex
	duration time.Duration
	Output   chan struct{}
	stop     chan struct{}
	timer    *time.Timer
}

// New creates a new debouncer
func New(duration time.Duration) *Debouncer {
	if duration == 0 {
		duration = time.Millisecond * 125
	}

	d := &Debouncer{
		duration: duration,
		Output:   make(chan struct{}),
		stop:     make(chan struct{}),
	}

	return d
}

// Trigger puts an event into the debouncer's pipeline.
func (d *Debouncer) Trigger() {
	d.Lock()
	defer d.Unlock()

	if d.timer != nil {
		d.timer.Reset(d.duration)
		return
	}

	d.timer = time.NewTimer(d.duration)

	go func() {
		select {
		case <-d.timer.C:
			d.Lock()
			d.Output <- struct{}{}
			d.timer = nil
			d.Unlock()
		case <-d.stop:
			return
		}
	}()
}

// Close closes and cleans after the debouncer.
func (d *Debouncer) Close() {
	d.Lock()
	defer d.Unlock()

	if d.timer != nil {
		d.stop <- struct{}{}
		d.timer.Stop()
	}

	close(d.stop)
	close(d.Output)
}
