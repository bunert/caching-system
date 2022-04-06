package objectManager

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/config"
	"time"
)

var (
	conf config.Conf

	loc, _ = time.LoadLocation("Europe/Berlin")
)

func LoadConfiguration() {
	log.Info("Load ObjectManager config file (conf.yaml)")

	conf.ReadConf()
	awsutils.LambdaSetTick(conf.Tick)
	conf.PrintConfig()
}

func ProcessGet(key string, endpoint string, t time.Time) {
	// get ObjectState or create new one if not present
	state := GetOrCreateState(key)

	// process synchronized
	state.mu.Lock()
	defer state.mu.Unlock()

	state.AddTimestamp(t)

	// TODO: clean processing
	switch state.state {
	case S3:
		// lambda not running, check if it should be started
		state.checkLambdaThreshold()
	case Lambda:
		// increment counter
		state.incLambdaCount()
		// check counter
		state.checkLambdaCount()
	case RedisBootup:
		break
	case RedisRunning:
		break
	case RedisShutdown:
		state.checkLambdaThresholdRestricted()
	case RestrictedLambda:
		state.incLambdaCount()
	default:
		log.Warn("unknown object state?")
		return
	}

}
