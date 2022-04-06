package objectManager

import (
	"cloud-storage/core/awsutils"
	g "cloud-storage/core/utility/globals"
	"cloud-storage/orchestrator/control/redisControl"
	"cloud-storage/orchestrator/gatewayComm"
	"cloud-storage/orchestrator/gatewayComm/client"
	"cloud-storage/orchestrator/lambdaComm"
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "objectManager")

	// mapping for bookkeeping about requests per object
	getMap = sync.Map{}
)

type State int

const (
	Undefined State = iota
	S3
	Lambda
	RedisBootup
	RedisRunning
	RedisShutdown
	RestrictedLambda

	queueSize = 5
)

type ObjectState struct {
	mu          sync.RWMutex
	state       State
	Key         string
	lambdaCount int
	queue       *list.List
}

func GetOrCreateState(key string) *ObjectState {
	// create if not found
	state, ok := getMap.Load(key)
	if ok {
		// key present
		return state.(*ObjectState)
	} else {
		// key not present
		newState := &ObjectState{
			state:       S3,
			Key:         key,
			lambdaCount: 0,
			queue:       list.New(),
		}

		state, _ = getMap.LoadOrStore(key, newState)
		log.Debugf("no ObjectState, create new one (key: %s): %s", key, newState.state.String())

		return state.(*ObjectState)
	}
}

func (s State) String() string {
	return [...]string{"Undefined", "S3", "Lambda", "RedisBootup", "RedisRunning", "RedisShutdown", "RestrictedLambda"}[s]
}

func (s *ObjectState) GetState() State {
	return s.state
}

// locked
func (s *ObjectState) AddTimestamp(t time.Time) {

	if s.queue.Len() == queueSize {
		s.queue.Remove(s.queue.Front())
	}
	s.queue.PushBack(t)
}

// returns the difference of now and the timestamp of the 3rd last element
func (s *ObjectState) getLambdaWindow() time.Duration {
	// get lambdaWindowElements timestamp
	el := s.queue.Front()
	// safe, we know queue len >= lambdaWindowElements
	for i := 0; i < s.queue.Len()-conf.LambdaWindowElements; i++ {
		el = el.Next()
	}

	// convert list value to timestamp
	var timestamp time.Time
	switch el.Value.(type) {
	case time.Time:
		timestamp = el.Value.(time.Time)
	default:
		log.Panic("unknown List element type?")
	}

	now := time.Now().UTC()

	// log.Warnf("%v - %v = %v < %v", now, timestamp, window, time.Duration(time.Second*5))
	return now.Sub(timestamp)
}

// returns the difference between now and the first oldest timestamp
func (s *ObjectState) getRedisWindow() time.Duration {
	// get 3rd last element
	el := s.queue.Front()

	// convert list value to timestamp
	var timestamp time.Time
	switch el.Value.(type) {
	case time.Time:
		timestamp = el.Value.(time.Time)
	default:
		log.Panic("unknown List element type?")
	}

	now := time.Now().UTC()

	// log.Warnf("%v - %v = %v < %v", now, timestamp, window, time.Duration(time.Second*5))
	return now.Sub(timestamp)
}

func (s *ObjectState) LambdaReturnedLocked() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lambdaCount = 0

	if s.state == Lambda {
		log.Debug("LambdaReturnedLocked, state == Lambda, so reset to S3 state")
		s.ChangeState(S3)
	}

}

func (s *ObjectState) LambdaRestrictedReturnedLocked() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lambdaCount = 0

	if s.state == RestrictedLambda {
		log.Debug("LambdaRestrictedReturnedLocked, state == RestrictedLambda, so reset to S3 state")
		s.ChangeState(RedisShutdown)
	} else if s.state == Lambda {
		log.Debug("LambdaRestrictedReturnedLocked, state == Lambda, so reset to S3 state")
		s.ChangeState(S3)
	}

}

func (s *ObjectState) RedisStopped() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == RedisShutdown {
		log.Debug("RedisStopped, in RedisShutdown state")
		s.ChangeState(S3)
	} else if s.state == RestrictedLambda {
		log.Debug("RedisStopped, in RestrictedLambda state")
		s.ChangeState(Lambda)
	}

}

func (s *ObjectState) RedisRunningLocked() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Debug("RedisRunning, trigger manual lambda shutdown")

	if s.state != RedisBootup {
		log.Error("RedisRunningLocked but state not RedisBootup?")
	}

	// manual shutdown if state still RedisBootup
	lambdaComm.Shutdown(lambdaComm.LambdaClient)

	// change state
	s.ChangeState(RedisRunning)
}

func (s *ObjectState) PrepareRedis() {
	log.Debug("PrepareRedis, get object from S3 and push it to Redis, send update to Gateway")
	buf := awsutils.GetBytes(s.Key)

	err := redisControl.RedisSet(s.Key, buf.String())
	if err != nil {
		log.WithError(err).Error("PrepareRedis failed to write object to redis")
		return
	}

	// notify gateway about new key served by Redis
	client.RedisUpdate(gatewayComm.GatewayClient, "added", []string{s.Key})

}

func (s *ObjectState) RedisTicker() {

	log.Debug("Start RedisTicker")
	ticker := time.NewTicker(time.Minute)

	// checks RedisUtilization periodically in locked mode
	go func() {
		for range ticker.C {
			s.mu.Lock()

			log.Debug("RedisTicker, check RedisUtilization")
			// if true, shuts down redis and trigger state transition to RedisShutdown
			if s.checkRedisUtilization() {
				ticker.Stop()
			}

			s.mu.Unlock()
		}
	}()

}

func (s *ObjectState) ChangeState(state State) {
	log.Infof("change state: %s", state.String())
	s.state = state
}

func (s *ObjectState) ChangeStateLocked(state State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ChangeState(state)
}

func (s *ObjectState) incLambdaCount() {
	s.lambdaCount++
}

func (s *ObjectState) getLambdaCount() int {
	return s.lambdaCount
}

func (s *ObjectState) checkLambdaThreshold() {
	// if more than lambdaWindowElements in queue, check lambdaWindow
	if s.queue.Len() >= conf.LambdaWindowElements {

		// get lambdaWindow
		var window time.Duration = s.getLambdaWindow()

		// if 3 requests in 20 seconds, spinup lambda if not already running
		if window < conf.LambdaThreshold {
			log.Debug("LambdaThreshold passed, start Lambda-Runtime")
			// check if lambda already running
			s.ChangeState(Lambda)
			go func() {
				awsutils.OrchestratorStartLambda(s.Key)

				// reset state to S3 if still in Lambda state (locked)
				s.LambdaReturnedLocked()
			}()
		}
	}
}

func (s *ObjectState) checkLambdaThresholdRestricted() {
	// if more than lambdaWindowElements in queue, check lambdaWindow
	if s.queue.Len() >= conf.LambdaWindowElements {

		// get lambdaWindow
		var window time.Duration = s.getLambdaWindow()

		// if 3 requests in 20 seconds, spinup lambda if not already running
		if window < conf.LambdaThreshold {
			log.Debug("LambdaThreshold passed, start Lambda-Runtime")
			// check if lambda already running
			s.ChangeState(RestrictedLambda)
			go func() {
				awsutils.OrchestratorStartLambda(s.Key)

				// reset state to RedisShutdown if still in Lambda state (locked)
				s.LambdaRestrictedReturnedLocked()
			}()
		}
	}
}

func (s *ObjectState) checkLambdaCount() {
	// check if redis-server should be started
	if s.getLambdaCount() > conf.RedisThreshold {
		log.Debug("LambdaCount passed, start Redis instance and send KeepRunning to Lambda-Runtime")

		// send keep running to lambda runtime
		log.Info("send KeepRunning to lambda-runtime")
		lambdaComm.KeepRunning(lambdaComm.LambdaClient)

		// start redis, goroutine waits until redis instance is running to trigger state change
		s.StartRedis()

	}
}

func (s *ObjectState) checkRedisUtilization() bool {
	var window time.Duration = s.getRedisWindow()

	// window > 1 min (last 5 requests not within one minute)
	if window > conf.RedisUtilization {
		log.Debug("RedisUtilization not sufficient, shutdown Redis")
		notifyChan := make(chan struct{})

		log.Info("stop redis")
		// shutdown redis
		err := redisControl.StopRedis(notifyChan, s.Key)
		if err != nil {
			log.WithError(err).Error("StopRedis in checkRedis failed")
		}
		g.LogFile.WriteString(fmt.Sprintf("%d\n", time.Now().In(loc).Unix()))

		s.ChangeState(RedisShutdown)

		go func() {
			<-notifyChan
			defer close(notifyChan)
			log.Info("redis stopped")
			s.RedisStopped()
		}()
		return true
	}
	return false
}

func (s *ObjectState) StartRedis() {
	s.ChangeState(RedisBootup)

	log.Info("startup Redis")

	// goroutine waiting until redis is running, notified by goroutine in StartRedis
	go func() {
		notifyChan := make(chan struct{})
		// start redis-server
		err := redisControl.StartRedis(notifyChan)
		if err != nil {
			log.WithError(err).Error("StartRedis in objectManager failed")
			return
		}

		// block until redis running
		<-notifyChan
		defer close(notifyChan)
		g.LogFile.WriteString(fmt.Sprintf("%d\t", time.Now().In(loc).Unix()))

		log.Info("redis running")
		// get object from S3 and put it on redis instance
		s.PrepareRedis()

		// send manual shutdown for lambda-runtime and change state to RedisRunning
		s.RedisRunningLocked()

		// create ticker and check every 30 seconds
		s.RedisTicker()
	}()
}

// func (s *ObjectState) printQueue() {
// 	log.Info("queue values: ")
// 	for e := s.queue.Front(); e != nil; e = e.Next() {
// 		log.Infof("\t %v", e.Value)
// 	}
// }
