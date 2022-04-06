package ginAPI

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/redis"
	"cloud-storage/core/utility/globals"
	"cloud-storage/core/utility/logger"
	docs "cloud-storage/gateway/ginAPI/swaggerDocs"
	"fmt"
	"net/http"
	"time"

	g "cloud-storage/gateway/global"
	"cloud-storage/gateway/lambdaComm"
	"cloud-storage/gateway/orchestratorComm"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	log = logrus.WithField("component", "ginAPI")

	// loc, _ = time.LoadLocation("Europe/Berlin")
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

// @title Gateway Server API
// @version 1.0
// @description Gateway (Proxy) Server acting as a middle man for API requests for an actual Cloud Storage Service.
// @description * GET requests for S3 working
// @description * POST requests for S3 working (not used atm)
// @termsOfService http://swagger.io/terms/

// @contact.name Tobias Buner
// @contact.email bunert@ethz.ch

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api/v1
// @schemes http
func RunAPI() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(LogrusLogger())

	r.Use(func(c *gin.Context) {
		c.Set("sess", awsutils.AWSsession)
		c.Next()
	})

	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", globals.SwaggerIP, globals.SwaggerPort)
	log.Infof("Swagger API Link: %s", fmt.Sprintf("%s:%d/swagger/index.html", globals.SwaggerIP, globals.SwaggerPort))

	// Routes
	userAPI := r.Group("/api/v1")
	{
		userAPI.GET("/", HealthCheck)
		userAPI.GET("/objects/:key", GetRequest)
		userAPI.POST("/objects/upload", UploadImage)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := r.Run(":4000"); err != nil {
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

// DownloadImageByKey godoc
// @Summary Get the image referenced by the Key from the S3 bunert-testbucket.
// @Description get the object from S3 referenced by the Key.
// @Tags Testing
// @Accept */*
// @Param key path string true "S3 Key (e.g. index.html)"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} object
// @Router /objects/{key} [get]
func GetRequest(c *gin.Context) {
	key := c.Param("key")

	// lookup for forwarding
	val, found := g.ReqMap.Load(key)
	// log.Warnf("Get %d", time.Now().In(loc).UnixMilli())

	// goroutine sends information to orchestrator
	if found {
		go orchestratorComm.ReceivedGet(orchestratorComm.OrchClient, key, val.(string))
	} else {
		go orchestratorComm.ReceivedGet(orchestratorComm.OrchClient, key, "s3")
	}

	if found {
		switch val {
		case "lambda":
			log.Debug("request forwarded: \t Lambda-Runtime")
			lambdaComm.Get(c, key)
			return
		case "redis":
			log.Debug("request forwarded: \t EC2 Redis")
			redis.Get(c, key)
			return
		default:
			log.Panic("ReqMap value unknown")
		}
	} else {
		log.Debug("request forwarded: \t S3")
		awsutils.AccessStorage(c, key)
		return
	}
}

// UploadImage godoc
// @Summary Upload a given image to the S3 bucket bunert-testbucket
// @Description uploads the image from the body to the S3 bucket.
// @Tags Testing
// @Accept */*
// @Param file formData file true "image"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} object
// @Router /objects/upload [post]
func UploadImage(c *gin.Context) {

	awsutils.UploadStorage(c)

}
