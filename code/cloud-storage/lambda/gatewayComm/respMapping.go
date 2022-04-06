package gatewayComm

import (
	"sync"
)

var (
	// mapping to object for a given key (downloaded from S3)
	RespMap sync.Map

	// number of objects currently in Respmap, used to send the keys which are served by the lambda-runtime to the gateway
	NumberOfObjects = 0
)
