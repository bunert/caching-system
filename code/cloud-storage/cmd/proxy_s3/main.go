package main

/*
WARNING: This is not part of the actual cloud-storage system, just a proxy for testing when using S3 only.

Running a Proxy Server acting as API endpoint for client requests.
Forwards the requests S3.
*/

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/utility/logger"
	"flag"
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
	log.Info("S3 only Ready, start simulation")
	if err := r.Run(":4000"); err != nil {
		log.Fatal(err)
	}

}

func GetRequest(c *gin.Context) {
	key := c.Param("key")

	awsutils.AccessStorage(c, key)
}

func main() {
	flag.Parse()
	logger.SetupLogger(*isDebug)

	awsutils.LoadEnv()
	log.Info("Starting S3 Proxy")

	// initiate AWS session
	awsutils.SetupS3ProxySession()

	done := make(chan struct{}, 1)

	// Register signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

	go RunAPI()

	// setup GIN API
	go func() {
		<-sig
		log.Info("Receive signal, killing S3 Proxy...")
		close(sig)

		close(done)
	}()

	<-done
	os.Exit(0)
}
