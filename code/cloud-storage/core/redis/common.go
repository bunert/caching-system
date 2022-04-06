package redis

import (
	"cloud-storage/gateway/global"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var (
	ErrRedisKeyNotFound = errors.New("EC2 Redis Key not found")
)

// Triggers a Redis Ping. (Redis Client must be initialized beforehand)
// returns an error if pong not received
func RedisPing() error {
	if !CheckRedisClient() {
		log.Warn("RedisPing, but no redis client initialized")
		return ErrRedisNoClient
	}
	r := GetRedisClient()

	pong, err := r.Ping().Result()
	if err != nil {
		log.WithError(err).Error("Redis Ping error")
		return ErrRedisPing
	}
	log.Info("Ping response: ", pong)
	return nil

}

// Triggers a Redis Get. (Redis Client must be initialized beforehand)
// returns the object as key with nil error if success
// returns an error if something went wrong
func RedisGet(key string) (string, error) {
	if !CheckRedisClient() {
		log.Warn("RedisGet, but no redis client initialized")
		return "", ErrRedisNoClient
	}
	r := GetRedisClient()

	val, err := r.Get(key).Result()

	if err == redis.Nil {
		log.Debugf("key (%s) does not exist", key)
		return "", ErrRedisKeyNotFound
	}

	if err != nil {
		log.WithError(err).Error("Redis Get failed")
		return "", ErrRedisGet
	}

	if val == "" {
		log.Warnf("value for key (%s) is empty", key)
		return "", ErrRedisValueEmpty
	}

	// log.Infof("value for %s: %s", key, val)
	return val, nil
}

// Triggers a Redis Set for the argument key and value. (Redis Client must be initialized beforehand)
// returns an error if not successfull
func RedisSet(key string, value string) error {
	if !CheckRedisClient() {
		log.Warn("RedisSet, but no redis client initialized")
		return ErrRedisNoClient
	}
	r := GetRedisClient()

	err := r.Set(key, value, 0).Err()
	if err != nil {
		log.WithError(err).Error("Redis Set failed")
		return ErrRedisSet
	}
	log.Debugf("added key (%s) to Redis", key)

	return nil

}

// Triggers a Redis Get taking a gin Context as additional argument, used by the Gateway (Redis Client must be initialized beforehand)
// writes the response to the gin context
// 404 if RedisGet failed
// 200 if RedisGet succeded
func Get(c *gin.Context, key string) {
	global.RedisLock.RLock()
	defer global.RedisLock.RLocker().Unlock()

	val, err := RedisGet(key)
	if err != nil {
		log.WithError(err).Errorf("fetching object %s from redis failed", key)
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Content-Origin", "redis")
	c.Status(http.StatusOK)
	c.Writer.WriteString(val)
}
