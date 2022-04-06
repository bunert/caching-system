package global

import (
	"sync"
)

var (
	// mapping if a lambda-runtime is running for a given key
	ReqMap sync.Map

	// mapping for reqId to a channel where the body is received
	RespChan sync.Map

	RedisLock  sync.RWMutex
	LambdaLock sync.RWMutex
)
