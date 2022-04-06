package runtime

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "runtime")

	mu sync.RWMutex

	runtime *Runtime
)

type Runtime struct {
	Timeout *TimeoutState

	// channel to force shutdown
	done chan struct{}
}

// creates runtime if not already created
func CreateOrGetRuntime(t int) *Runtime {
	mu.Lock()
	defer mu.Unlock()

	// TICK set to time duration as specified by the invoke argument
	TICK = time.Second * time.Duration(t)

	if runtime == nil {
		log.Info("create new runtime")
		runtime = &Runtime{done: make(chan struct{})}

		runtime.Timeout = NewTimeout(runtime)
	} else {
		log.Info("runtime already created, warm startup (reset timeout)")
		// TODO: clean start
		runtime.Timeout.Restart()
	}

	return runtime
}

func GetRuntime() *Runtime {
	mu.RLock()
	defer mu.RUnlock()

	return runtime
}

func (r *Runtime) WaitDone() <-chan struct{} {
	return r.done
}

func (r *Runtime) Done() {
	mu.Lock()
	defer mu.Unlock()

	r.DoneLocked()
}

func (r *Runtime) DoneLocked() {
	r.done <- struct{}{}
}

func (r *Runtime) Lock() {
	mu.Lock()
}

func (r *Runtime) Unlock() {
	mu.Unlock()
}
