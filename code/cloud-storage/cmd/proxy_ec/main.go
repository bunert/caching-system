package main

/*
WARNING: This is not part of the actual cloud-storage system, just a proxy for testing when using ElastiCache only.

Running a Proxy Server acting as API endpoint for client requests.
Forwards the requests ElastiCache.
*/

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/redis"
	"cloud-storage/core/utility/logger"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	// cmd-line flag
	isDebug = flag.Bool("debug", false, "specifiy log level (debug/production), default production")
)

func LogrusLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()               // Starting time
		c.Next()                              // Processing request
		endTime := time.Now()                 // End Time
		latencyTime := endTime.Sub(startTime) // execution time
		reqMethod := c.Request.Method         // Request method
		reqUri := c.Request.RequestURI        // Request route
		statusCode := c.Writer.Status()       // status code
		clientIP := c.ClientIP()              // Request IP
		statusColor := logger.StatusCodeColor(statusCode)
		methodColor := logger.MethodColor(reqMethod)
		resetColor := logger.ResetColor()
		routing := c.Writer.Header().Get("Content-Origin")

		if statusCode == 200 {
			// Log format
			log.Infof("|%s %3d %s| %13v | %15s | %7s |%s %-7s %s %#v",
				statusColor, statusCode, resetColor,
				latencyTime,
				clientIP,
				routing,
				methodColor, reqMethod, resetColor,
				reqUri,
			)
		} else {
			// Log format
			log.Warnf("|%s %3d %s| %13v | %15s | %7s |%s %-7s %s %#v",
				statusColor, statusCode, resetColor,
				latencyTime,
				clientIP,
				routing,
				methodColor, reqMethod, resetColor,
				reqUri,
			)
		}
	}
}

func RunAPI() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(LogrusLogger())

	r.Use(func(c *gin.Context) {
		c.Set("sess", awsutils.AWSsession)
		c.Next()
	})

	// Routes
	userAPI := r.Group("/api/v1")
	{
		userAPI.GET("/objects/:key", GetRequest)
	}
	log.Info("Running Gin API")
	if err := r.Run(":4000"); err != nil {
		log.Fatal(err)
	}

}

func GetRequest(c *gin.Context) {
	key := c.Param("key")

	// TODO: adapt this part and use ElastiCache if key available, otherwhise get from S3 and push to ElastiCache

	val, err := redis.RedisGet(key)
	if err != nil {
		if err == redis.ErrRedisKeyNotFound {
			// key not in redis, get from S3 and put it to redis
			log.Debugf("key (%s) does not exists in ElastiCache, access S3 and push it to redis", key)

			buf := awsutils.GetAndReturnObject(c, key)
			if buf == nil {
				log.Warn("GetAndReturnObject did not retrieve an object?")
				return
			}

			err := redis.RedisSet(key, buf.String())
			if err != nil {
				log.WithError(err).Error("PrepareRedis failed to write object to redis")
				return
			}

			return
		} else {
			log.WithError(err).Errorf("fetching object %s from redis failed", key)
			c.Status(http.StatusNotFound)
			return
		}
	}

	c.Header("Content-Origin", "EC")
	c.Status(http.StatusOK)
	c.Writer.WriteString(val)

	// awsutils.AccessStorage(c, key)
}

func main() {
	flag.Parse()
	logger.SetupLogger(*isDebug)

	done := make(chan struct{}, 1)

	// Register signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

	awsutils.LoadEnv()
	log.Info("Starting ElastiCache Proxy")

	go func() {
		// initiate AWS session
		awsutils.SetupECProxySession()
		awsutils.CreateCluster()
		log.Info("ElastiCache Cluster created, wait until available")
		awsutils.WaitForCluster()
		log.Info("ElastiCache Cluster available")

		addr, err := awsutils.GetClusterAddr()
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("ElastiCache addr: %s", addr)

		// initialize Redis Client to use
		redis.NewECRedisClient(addr)

		// TODO: remove PING, just to make sure for now
		err = redis.RedisPing()
		if err != nil {
			log.Warn("Redis Ping failed")
			return
		}
		log.Info("ElastiCache Ready, start simulation")
	}()

	go RunAPI()

	// setup GIN API
	go func() {
		<-sig
		log.Info("Receive signal, killing S3 Proxy...")
		close(sig)

		log.Info("Deleting ElastiCache Cluster")
		awsutils.DeleteCluster()
		log.Info("ElastiCache Cluster deleted")

		close(done)
	}()

	<-done
	os.Exit(0)
}
