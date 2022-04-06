package orchestratorComm

import (
	"cloud-storage/core/redis"
	"cloud-storage/gateway/global"
	"strings"

	"github.com/mason-leap-lab/redeo"
	"github.com/mason-leap-lab/redeo/resp"
	logrus "github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "orchComm")
	// loc, _ = time.LoadLocation("Europe/Berlin")
)

func HandleLogGatewayState(w resp.ResponseWriter, c *resp.Command) {
	client := redeo.GetClient(c.Context())
	connId := int(client.ID())

	log.Infof("LogGatewayState \t\t [connId: %d]", connId)

	log.Infof("%s", strings.Repeat("-", 30))
	log.Infof("---%s%16s%s---", strings.Repeat(" ", 4), "Forwarding Table", strings.Repeat(" ", 4))
	log.Infof("%s", strings.Repeat("-", 30))
	log.Infof("%-20s%10s", "key", "endpoint")
	log.Infof("%s", strings.Repeat("-", 30))
	global.ReqMap.Range(func(key interface{}, value interface{}) bool {
		log.Infof("%-20s%10s", key, value)
		return true
	})

}

func HandleRedisStart(w resp.ResponseWriter, c *resp.Command) {
	client := redeo.GetClient(c.Context())
	connId := int(client.ID())
	addr := c.Arg(0).String()

	log.Debugf("RedisStart: %s \t\t [connId: %d]", addr, connId)

	// initialize Redis Client to use
	redis.NewRedisClient(addr)

	// TODO: remove PING, just to make sure for now
	// err := redis.RedisPing()
	// if err != nil {
	// 	log.Warn("Redis Ping failed")
	// 	return
	// }

}

// Handler for RedisStop message from the Orchestrator.
// Wait for the Writer redis lock before removing redis client and sending ACK to orchestrator.
// Redis read lock is aquired by each RedisGet request to make sure inflight requests are served before shutting down,
func HandleRedisStop(w resp.ResponseWriter, c *resp.Command) {
	client := redeo.GetClient(c.Context())
	connId := int(client.ID())

	log.Debugf("RedisStop \t\t [connId: %d]", connId)

	// waits until no more Redis Gets in-flight
	global.RedisLock.Lock()
	defer global.RedisLock.Unlock()

	redis.RemoveRedisClient()
	// send response OK, so Orchestrator knows it can shut down the instance
	w.AppendOK()
	if err := w.Flush(); err != nil {
		log.Panic("Failed to ACK RedisStop")
	}
}

func HandleRedisUpdate(w resp.ResponseWriter, c *resp.Command) {
	client := redeo.GetClient(c.Context())
	connId := int(client.ID())
	state := c.Arg(0).String()
	len64, _ := c.Arg(1).Int()
	len := int(len64) // convert int64 to int (c.Arg requires int)

	log.Debugf("RedisUpdate (number of keys: %d): %s \t\t [connId: %d]", len, state, connId)

	switch state {
	case "added":
		for i := 1; i <= len; i++ {
			key := c.Arg(1 + i).String()
			global.ReqMap.Store(key, "redis")
			// log.Warnf("Redis %d", time.Now().In(loc).UnixMilli())
			log.Debugf("\t\t%-20s%10s", key, "redis")
		}

		return
	case "removed":
		for i := 1; i <= len; i++ {
			key := c.Arg(1 + i).String()
			global.ReqMap.Delete(key)
			log.Debugf("\t\t%-20s%10s", key, "removed")
		}

		return
	default:
		log.Panic("unknown RedisUpdate state")
	}

}
