package ginAPI

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/redis"
	g "cloud-storage/core/utility/globals"
	"cloud-storage/core/utility/logger"
	"cloud-storage/orchestrator/control/redisControl"
	"cloud-storage/orchestrator/gatewayComm"
	"cloud-storage/orchestrator/gatewayComm/client"
	docs "cloud-storage/orchestrator/ginAPI/swaggerDocs"
	"cloud-storage/orchestrator/lambdaComm"
	"cloud-storage/orchestrator/objectManager"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	log = logrus.WithField("component", "ginAPI")
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

		if statusCode == 200 {
			// Log format
			log.Infof("|%s %3d %s| %13v | %15s |%s %-7s %s %#v",
				statusColor, statusCode, resetColor,
				latencyTime,
				clientIP,
				methodColor, reqMethod, resetColor,
				reqUri,
			)
		} else {
			// Log format
			log.Warnf("|%s %3d %s| %13v | %15s |%s %-7s %s %#v",
				statusColor, statusCode, resetColor,
				latencyTime,
				clientIP,
				methodColor, reqMethod, resetColor,
				reqUri,
			)
		}
	}
}

// @title Orchestrator Control API
// @version 1.0
// @description Functionality to test the control mechanism of the Orchestrator.
// @description * Start Lambda-Runtimes
// @description * Start/Stop EC2 Redis instance
// @termsOfService http://swagger.io/terms/

// @contact.name Tobias Buner
// @contact.email bunert@ethz.ch

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api/v1
// @schemes http
func RunControlAPI() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(LogrusLogger())

	r.Use(func(c *gin.Context) {
		c.Set("sess", awsutils.AWSsession)
		c.Next()
	})

	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", g.SwaggerIP, g.SwaggerPort)
	log.Infof("Swagger API Link: %s", fmt.Sprintf("%s:%d/swagger/index.html", g.SwaggerIP, g.SwaggerPort))

	// Routes
	userAPI := r.Group("/api/v1")
	{
		// GinAPI
		userAPI.GET("/", HealthCheck)

		// Gateway Info
		userAPI.GET("/gateway/state", GatewayState)

		// Lambda-Runtime
		userAPI.POST("/lambda/start", StartLambda)
		userAPI.POST("/lambda/stop", StopLambda)

		// EC2 Redis
		userAPI.POST("/ec2/start", StartEC2)
		userAPI.POST("/ec2/stop", StopEC2)

		userAPI.GET("/redis/ping", RedisPing)
		userAPI.GET("/redis/allKeys", RedisGetKeys)
		userAPI.GET("/redis/get/:key", RedisGet)
		userAPI.POST("/redis/set/file", RedisSetFile)
		userAPI.POST("/redis/set/inline", RedisSetInline)
		userAPI.POST("/redis/del", RedisDel)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := r.Run(":3000"); err != nil {
		log.Fatal(err)
	}

	log.Info("API server exited.")
}

// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func HealthCheck(c *gin.Context) {
	res := map[string]interface{}{
		"status": "Server is up and running",
	}

	c.JSON(http.StatusOK, res)
}

// Start EC2
// @Summary Start an EC2 Redis instance serving the given object.
// @Description Starts the corresponding EC2 Redis instance.
// @Tags EC2 Redis Management
// @Accept */*
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /ec2/start [post]
func StartEC2(c *gin.Context) {

	// ATTENTION: this is just for testing, ObjectState on Orchestrator ignored here
	err := redisControl.StartRedis(nil)

	if err != nil {
		res := map[string]interface{}{
			"status": err.Error(),
		}
		c.JSON(http.StatusOK, res)
		return
	}

	res := map[string]interface{}{
		"status": "started EC2 Redis instance",
	}

	c.JSON(http.StatusOK, res)
}

// Start EC2
// @Summary Stop the EC2 Redis instance serving the given object.
// @Description Stops the corresponding EC2 Redis instance.
// @Tags EC2 Redis Management
// @Accept */*
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /ec2/stop [post]
func StopEC2(c *gin.Context) {

	// ATTENTION: this is just for testing, ObjectState on Orchestrator ignored here
	err := redisControl.StopRedis(nil, "")

	if err != nil {
		res := map[string]interface{}{
			"status": err.Error(),
		}
		c.JSON(http.StatusOK, res)
		return
	}

	res := map[string]interface{}{
		"status": "EC2 instance stopped",
	}

	c.JSON(http.StatusOK, res)
}

// Redis Ping
// @Summary EC2 Redis Ping for a fixed InstandeId
// @Description Executes a Redis Ping request.
// @Tags EC2 Redis Operations
// @Accept */*
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /redis/ping [get]
func RedisPing(c *gin.Context) {

	err := redis.RedisPing()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	res := map[string]interface{}{
		"status": "EC2 Redis Ping successfull",
	}

	c.JSON(http.StatusOK, res)
}

// Redis Get all Keys
// @Summary EC2 Redis retrieve all current keys
// @Description Scans for all available keys currently stored on the given Redis instance.
// @Tags EC2 Redis Operations
// @Accept */*
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /redis/allKeys [get]
func RedisGetKeys(c *gin.Context) {

	keys, err := redisControl.RedisKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	res := map[string]interface{}{
		"keys": fmt.Sprintf("%v", keys),
	}

	c.JSON(http.StatusOK, res)

}

// Redis Get
// @Summary Get an object from EC2 Redis instance.
// @Description Uses Redis Client to retrieve the specified object.
// @Tags EC2 Redis Operations
// @Accept */*
// @Param key path string true "Object Key (e.g. index.html)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /redis/get/{key} [get]
func RedisGet(c *gin.Context) {
	key := c.Param("key")

	val, err := redis.RedisGet(key)
	switch {
	case err == redis.ErrRedisKeyNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	default:

	}

	res := map[string]interface{}{
		"val": val,
	}

	c.JSON(http.StatusOK, res)
}

// Redis SetFile
// @Summary Pushes an object to the EC2 Redis instance
// @Description Uses Redis Client to push the specified object.
// @Tags EC2 Redis Operations
// @Accept */*
// @Param file formData file true "value"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /redis/set/file/ [post]
func RedisSetFile(c *gin.Context) {
	file, err := c.FormFile("file")

	if err != nil {
		log.WithError(err).Errorf("error retrieving object %s from HTTP request: ", file.Filename)
		c.JSON(http.StatusBadRequest, gin.H{
			"filename": file.Filename,
		})
		return
	}
	f, err := file.Open()
	if err != nil {
		log.WithError(err).Error("opening file failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read from file",
		})
		return
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.WithError(err).Error("reading file failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read from file",
		})
		return
	}

	err = redisControl.RedisSet(file.Filename, string(bytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// notify gateway about new key served by Redis
	client.RedisUpdate(gatewayComm.GatewayClient, "added", []string{file.Filename})

	res := map[string]interface{}{
		"status": "RedisSet success",
	}

	c.JSON(http.StatusOK, res)
}

// Redis SetInline
// @Summary Pushes an object to the EC2 Redis instance
// @Description Uses Redis Client to push the specified object.
// @Tags EC2 Redis Operations
// @Accept */*
// @Param key query string true "object key name"
// @Param value query string true "object value"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /redis/set/inline [post]
func RedisSetInline(c *gin.Context) {
	key := c.Query("key")
	val := c.Query("value")

	err := redisControl.RedisSet(key, val)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// notify gateway about new key served by Redis
	client.RedisUpdate(gatewayComm.GatewayClient, "added", []string{key})

	res := map[string]interface{}{
		"status": "RedisSet success",
	}

	c.JSON(http.StatusOK, res)
}

// Redis Del
// @Summary Removes an objectfrom the EC2 Redis instance
// @Description Uses Redis Client to remove the specified object.
// @Tags EC2 Redis Operations
// @Accept */*
// @Param key query string true "Object Key (e.g. index.html)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /redis/del/ [post]
func RedisDel(c *gin.Context) {
	key := c.Query("key")
	log.Info("Del EC2 Redis: ", key)

	// notify gateway about removed Redis keys
	client.RedisUpdate(gatewayComm.GatewayClient, "removed", []string{key})

	err := redisControl.RedisDel(key)
	// TODO: if delete failed, gateway should be notified that it is still available on Redis
	if err != nil {
		switch {
		case err == redis.ErrRedisDelKeyNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "RedisDel success",
	})

}

// Start Lambda-Runtime
// @Summary Start a Lambda-Runtime serving the given object.
// @Description Starts a corresponding Lambda-Runtime.
// @Description Keep in mind that the Lambda-Runtime currently works with a timer with shuts down after 5 seconds when no request was received.
// @Description After a request was received, the timer extends for 1 second.
// @Tags Lambda-Runtime Management
// @Accept */*
// @Param key query string true "Object Key (e.g. index.html)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /lambda/start [post]
func StartLambda(c *gin.Context) {
	key := c.Query("key")
	log.Info("Start Lambda-Runtime: ", key)

	state := objectManager.GetOrCreateState(key)

	if state.GetState() == objectManager.S3 {
		state.ChangeStateLocked(objectManager.Lambda)
		go func() {
			awsutils.OrchestratorStartLambda(state.Key)

			// get lock when lambda-runtime returned to change state in lock safe manner
			state.LambdaReturnedLocked()
		}()

		c.JSON(http.StatusOK, gin.H{
			"status": "Lambda-Runtime started",
		})
		return
	}

	log.Infof("Lambda-Runtime for key (%s) already running", key)
	c.JSON(http.StatusOK, gin.H{
		"status": "Lambda-Runtime already running",
	})
}

// Stop Lambda-Runtime
// @Summary Stop the running Lambda-Runtime.
// @Description Stops the corresponding Lambda-Runtime.
// @Tags Lambda-Runtime Management
// @Accept */*
// @Param key query string true "Object Key (e.g. index.html)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /lambda/stop [post]
func StopLambda(c *gin.Context) {
	key := c.Query("key")
	log.Info("Stop Lambda-Runtime: ", key)

	state := objectManager.GetOrCreateState(key)

	if state.GetState() == objectManager.Lambda {
		lambdaComm.Shutdown(lambdaComm.LambdaClient)

		c.JSON(http.StatusOK, gin.H{
			"status": "Lambda-Runtime stopped",
		})
		return
	}

	log.Infof("Lambda-Runtime for key (%s) not running", key)
	c.JSON(http.StatusOK, gin.H{
		"status": "Lambda-Runtime not running or in wrong state for manual shutdown?",
	})
}

// Logs the forwarding state at the Gateway Log
// @Summary Show the status of the Gateway forwarding rules in the Gateway Log.
// @Description logs the forwarding status.
// @Tags Gateway Info
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /gateway/state [get]
func GatewayState(c *gin.Context) {

	// TODO: get info from gateway
	client.LogGatewayState(gatewayComm.GatewayClient)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
