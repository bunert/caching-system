package runtime

import (
	"sync/atomic"
	"time"
)

var (
	TICK time.Duration
)

type TimeoutState struct {
	runtime *Runtime

	start time.Time

	// channel to reset/extend timer, time.Duration is used to actually extend the timeout for the given duration
	reset chan time.Duration

	// channel used to notify main function that the timer expired
	c chan time.Time

	// channel used to notify main function when GatewayByeAck was received and shutdown can continue safely
	s chan struct{}

	// actual timer
	timer *time.Timer

	// initial duration which was used to configure timer
	due time.Duration

	// boolean if timeout reset
	hasReset bool

	// boolean if timeout expired
	timeout bool

	// flag if timer disabled
	disabled int32
}

// initialize new timeout and start validateTimeout go routine
func NewTimeout(r *Runtime) *TimeoutState {
	t := &TimeoutState{
		runtime: r,
		start:   time.Now(),
		reset:   make(chan time.Duration, 1),
		c:       make(chan time.Time, 1),
		s:       make(chan struct{}),
	}
	// get duration and time when timer should expire given the duration
	timeout, due := t.getTimeout(TICK)
	t.due = due

	// initialize new timer
	t.timer = time.NewTimer(timeout)

	log.Infof("Timer initialized (expected: %s) -> timeout in %v", t.start.Add(t.due).Format("2006-01-02 15:04:05.000"), timeout)

	// goroutine checking if reset channel received information about extending the timer or if the timer expires
	go t.checkTimeout(r.done)
	return t
}

func (t *TimeoutState) Restart() {
	t.start = time.Now()
	t.timeout = false
	if !t.Enable() {
		log.Info("already enabled")
	}
	// runtime already locked
	t.ResetTimerLocked(TICK)
}

func (t *TimeoutState) getTimeout(duration time.Duration) (timeout, due time.Duration) {
	// duration should always be positive
	if duration < 0 {
		log.Fatal("negative duration, exit")
	}

	now := time.Since(t.start) // time elapsed since start of the Timeout

	// computes the duration when the timer will expire, rounding up using the granularity of billing cycle duration of lambda functions (TICK)
	due = time.Duration(float64(now + duration))

	// computes the remaining time untile the duration expires
	timeout = due - now
	return
}

func (t *TimeoutState) extendTimeout(duration time.Duration) (timeout, due time.Duration) {
	// duration should always be positive
	if duration < 0 {
		log.Fatal("negative duration, exit")
	}

	now := time.Since(t.start) // time elapsed since start of the Timeout

	// computes the duration when the timer will expire, rounding up using the granularity of billing cycle duration of lambda functions (TICK)
	// TODO: due computation buggy, too high
	due = time.Duration(float64(t.due + duration))

	// computes the remaining time untile the duration expires
	timeout = due - now
	return
}

func (t *TimeoutState) checkTimeout(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case extension := <-t.reset: // case if timeout extension is waiting on reset channel
			t.Stop() // stops timer or consumes event if not stoppable

			timeout, due := t.getTimeout(extension)
			t.due = due
			t.timer.Reset(timeout)
			log.Infof("Timer updated (expected: %s) -> timeout in %v", t.start.Add(t.due).Format("2006-01-02 15:04:05.000"), timeout)
			t.hasReset = false

		case ti := <-t.timer.C: // case if timeout expired
			// main timeout channel should be empty, or we clear it
			select {
			case <-t.c:
			default:
				// Nothing
			}
			t.runtime.Lock()

			if t.hasReset || t.IsDisabled() {
				// pass
			} else {
				log.Warn("timer expired (checkTimeout), notify main function")
				t.c <- ti
				t.timeout = true
			}

			t.runtime.Unlock()
		}
	}
}

// compute new timeout due and push it on the reset channel
func (t *TimeoutState) ResetTimer(ext time.Duration) {
	t.runtime.Lock()
	defer t.runtime.Unlock()

	t.ResetTimerLocked(ext)
}

func (t *TimeoutState) ResetTimerLocked(ext time.Duration) {
	if t.timeout || t.IsDisabled() {
		log.Debug("ResetTimer but timeout expired or disabled")
	} else {
		t.reset <- ext

		t.hasReset = true
	}
}

// compute new timeout due and push it on the reset channel
func (t *TimeoutState) ExtendTimer() {
	t.runtime.Lock()
	defer t.runtime.Unlock()

	t.ExtendTimerLocked()
}

func (t *TimeoutState) ExtendTimerLocked() {
	if t.timeout || t.IsDisabled() {
		log.Debug("ExtendTimer but timeout expired or disabled")
	} else {
		timeout, _ := t.extendTimeout(TICK)

		t.reset <- timeout

		t.hasReset = true
	}
}

func (t *TimeoutState) C() <-chan time.Time {
	return t.c
}

func (t *TimeoutState) GatewayAckedShutdown() {
	t.s <- struct{}{}
}

func (t *TimeoutState) S() <-chan struct{} {
	return t.s
}

// Drain the timer to be accurate and safe to reset.
func (t *TimeoutState) Stop() {
	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
		}
	}
}

// Disable and returns false if state has been disabled already
func (t *TimeoutState) Disable() bool {
	return atomic.CompareAndSwapInt32(&t.disabled, 0, 1)
}

// Enable and returns false if state has been enabled already
func (t *TimeoutState) Enable() bool {
	return atomic.CompareAndSwapInt32(&t.disabled, 1, 0)
}

func (t *TimeoutState) IsDisabled() bool {
	return atomic.LoadInt32(&t.disabled) > 0
}
