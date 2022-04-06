package redis

/*
go-redis documentation: https://redis.uptrace.dev/guide/server.html#connecting-to-redis-server

this package provides helper function for the actual go-redis calls
*/

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "redis")

	// errors
	ErrRedisNoClient       = errors.New("no redis client initialized")
	ErrRedisPing           = errors.New("EC2 Redis Ping failed")
	ErrRedisAllKeys        = errors.New("EC2 Redis Scan all keys failed")
	ErrRedisGet            = errors.New("EC2 Redis Get failed")
	ErrRedisValueEmpty     = errors.New("EC2 Redis Value empty")
	ErrRedisSet            = errors.New("EC2 Redis Set failed")
	ErrRedisDel            = errors.New("EC2 Redis Del failed")
	ErrRedisDelKeyNotFound = errors.New("EC2 Redis Del key not found")

	password = "cloud-testing" // hardcoded auth password for redis

	client *redis.Client
)

// creates a new Redis Client if not already initialized with the same addr
func NewRedisClient(addr string) {
	if client != nil && client.Options().Addr == fmt.Sprintf("%s:%v", addr, 6379) {
		log.Info("NewRedisClient, but client with same address already exists")
		return
	}

	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%v", addr, 6379),
		Password: password,
		DB:       0,
	})
	log.Debug("new Redis client initialized")
}

// creates a new Redis Client if not already initialized with the same addr
func NewECRedisClient(addr string) {
	if client != nil && client.Options().Addr == fmt.Sprintf("%s:%v", addr, 6379) {
		log.Info("NewRedisClient, but client with same address already exists")
		return
	}

	client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%v", addr, 6379),
		DB:   0,
	})
	log.Debug("new Redis client initialized")
}

// Check if Redis Client is initialized.
// return true
func CheckRedisClient() bool {
	return client != nil
}

// Return the redis client and panic if not initialized.
func GetRedisClient() *redis.Client {
	if client == nil {
		log.Panic("redisClient not initialized")
	}
	return client
}

// Set the redis client to nil. Used to prevent usage when instance is stopped.
func RemoveRedisClient() {
	client = nil
}
