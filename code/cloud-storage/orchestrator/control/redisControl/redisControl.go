package redisControl

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/redis"
	"cloud-storage/orchestrator/gatewayComm"
	"cloud-storage/orchestrator/gatewayComm/client"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "redisControl")
)

// setup Redis client for orchestrator if Redis instance is running when orchestrator is starting
func SetupRedis() {
	addr, err := awsutils.GetRedisServerAddr()
	if err != nil {
		// no addr, can't setup redis client
		return
	}
	// if addr retrieved, setup redis client
	redis.NewRedisClient(addr)
}

// Start the EC2 Redis instance.
// If instance is starting, triggers a goroutine which receives the addr through a channel as soon as the instance is running.
// As soon as the instance in running, triggers RedisStart and RedisUpdate message for the gateway.
// The RedisUpdate message includes all keys the Redis Server is currently serving which is queried as soon as the instance is running.
//
// returns an error:
// - nil if the instance status "stopped" and start successfully triggered
// - erros which should be handled
//		- awsutils.ErrInstanceAlreadyRunning: instance already running, this triggers RedisStart & RedisUpdate message for the Gateway
//		- awsutils.ErrInstanceAlreadyStarting: instance starting
//		- awsutils.ErrInstanceShuttingDown: instance shutting down
//		- awsutils.ErrUnhandledState: should not occur
//		- other errors should not happend
func StartRedis(c chan struct{}) error {
	notifyChan := make(chan string)

	err := awsutils.StartEC2Instance(notifyChan)

	if err != nil {
		// if EC2 instance already running
		if err == awsutils.ErrInstanceAlreadyRunning {
			log.Warn("Redis EC2 isntance already running when StartRedis called")

			log.Debug("notify Gateway about running EC2 Redis instance")
			addr, err := awsutils.GetRedisServerAddr()
			if err != nil {
				log.Panic("RedisServer addr not available?")
			}

			redis.NewRedisClient(addr)

			waitRedis()
			// notify channel if running
			if c != nil {
				c <- struct{}{}
			}

			// notify gateway
			client.RedisStart(gatewayComm.GatewayClient, addr)

			keys, err := RedisKeys()

			if err != nil {
				log.Panic("retrieve all redis keys failed")
			}

			// notify gateway about the available keys
			client.RedisUpdate(gatewayComm.GatewayClient, "added", keys)

		}
		return err
	}

	go func() {
		addr := <-notifyChan
		defer close(notifyChan)

		redis.NewRedisClient(addr)

		waitRedis()
		// notify channel if running
		if c != nil {
			c <- struct{}{}
		}

		log.Debug("notify Gateway about running EC2 Redis instance")

		// notify gateway
		client.RedisStart(gatewayComm.GatewayClient, addr)

		keys, err := RedisKeys()

		if err != nil {
			log.Panic("retrieve all redis keys failed")
		}

		// notify gateway about the available keys
		client.RedisUpdate(gatewayComm.GatewayClient, "added", keys)

	}()

	return nil
}

// Helper function for RedisStart.
// If EC2 instance running, redis-server possibly still not reachable until ready so ping every 300ms until rechable.
// Maximum of 10 retries (~3 s).
func waitRedis() {
	r := redis.GetRedisClient()
	// max of 10 retries
	for i := 0; i < 10; i++ {
		_, err := r.Ping().Result()
		if err != nil {
			// TODO better option than sleep?
			time.Sleep(time.Millisecond * 300)
			continue
		}
		return
	}
	log.Error("waitRedis reached maximum retries, redis not rechable")
}

// Stops the EC2 Redis instance.
// If a Redis client exists, sets the client to nil to prevent further usage.
// Afterwards it sends a RedisUpdate and RedisStop to the Gateway, the update queries all keys first which are served by the Redis Server.
// Finnaly stops the EC2 Instance, also works if the instance is starting up (pending state).
// awsutils.StopEC2Instance triggers a goroutine which waits until the instance is stopped which triggers the internal instance state change.
//
// returns an error:
// - nil if the instance status "running" or "pending" and stops the instance successfully
// - erros which should be handled
//		- awsutils.ErrInstanceAlreadyStopped: instance not running
//		- awsutils.ErrInstanceShuttingDown: already shutting down
//		- awsutils.ErrUnhandledState: should not occur
//		- other errors should not happend
func StopRedis(c chan struct{}, key string) error {
	if !redis.CheckRedisClient() {
		return awsutils.StopEC2Instance(c)
	}

	keys, err := RedisKeys()

	if err != nil {
		log.Panic("retrieve all redis keys failed")
	}

	// notify gateway about the removed keys
	client.RedisUpdate(gatewayComm.GatewayClient, "removed", keys)

	// notify gateway, RedisStop waits for ACK by gateway
	client.RedisStop(gatewayComm.GatewayClient)
	log.Debug("Gateway responded to RedisStop, shutdown instance")

	// remove key from redis if specified
	if key != "" {
		err := RedisDel(key)
		// TODO: if delete failed, gateway should be notified that it is still available on Redis
		if err != nil {
			log.WithError(err).Error("StopRedis with key but delete the key failed")
		}
	}

	// set redisConn to nil, so connection is not used anymore while stopping the EC2 instance
	redis.RemoveRedisClient()

	return awsutils.StopEC2Instance(c)
}

// Triggers a Redis scan to retrieve all keys currently served by the EC2 Redis instance.
// returns the list of served keys if error is nil
func RedisKeys() ([]string, error) {
	if !redis.CheckRedisClient() {
		log.Warn("RedisKeys, but no redis client initialized")
		return nil, redis.ErrRedisNoClient
	}
	r := redis.GetRedisClient()

	keys, _, err := r.Scan(0, "*", 0).Result()

	if err != nil {
		log.WithError(err).Error("Redis Scan failed")
		return nil, redis.ErrRedisAllKeys
	}

	log.Debugf("current Redis Keys: %v", keys)

	return keys, nil

}

// Triggers a Redis Set for the argument key and value. (Redis Client must be initialized beforehand)
// returns an error if not successfull
func RedisSet(key string, value string) error {
	if !redis.CheckRedisClient() {
		log.Warn("RedisSet, but no redis client initialized")
		return redis.ErrRedisNoClient
	}
	r := redis.GetRedisClient()

	err := r.Set(key, value, 0).Err()
	if err != nil {
		log.WithError(err).Error("Redis Set failed")
		return redis.ErrRedisSet
	}
	log.Debugf("added key (%s) to Redis", key)

	return nil

}

// Triggers a Redis Del for the argument key. (Redis Client must be initialized beforehand)
// returns an error if not successfull
func RedisDel(key string) error {
	if !redis.CheckRedisClient() {
		log.Warn("RedisDel, but no redis client initialized")
		return redis.ErrRedisNoClient
	}
	r := redis.GetRedisClient()

	cmd := r.Del(key)

	if err := cmd.Err(); err != nil {
		log.WithError(err).Error("Redis Del failed")
		return redis.ErrRedisDel
	}

	code, err := cmd.Result()
	if err != nil {
		log.Panic(err)
	}
	// code is the number of keys removed, if not 0 key was not present beforehand
	if code != 1 {
		log.Info("RedisDel but key not present")
		return redis.ErrRedisDelKeyNotFound
	}

	log.Debugf("removed key (%s) from Redis", key)
	return nil

}
