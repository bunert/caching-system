package main

/*
Running a Gateway Server acting as API endpoint for client requests.
Forwards the requests to either the storage or the lambda-runtime acting as cache.
Active feedback loop with the orchestrator to keep routing information up-to-date.
*/

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/utility/logger"
	"cloud-storage/lambda/gatewayComm"
	"cloud-storage/lambda/orchComm"
	r "cloud-storage/lambda/runtime"
	"cloud-storage/lambda/types"

	"github.com/google/uuid"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/mason-leap-lab/redeo"

	log "github.com/sirupsen/logrus"
)

var (
	gatewaySrv = redeo.NewServer(nil) // Serve requests from gateway
	orchSrv    = redeo.NewServer(nil) // Serve requests from orchestrator

	id string
)

func init() {
	// id logging to distinguish between lambda instances
	id = uuid.New().String()

	logger.SetupLambdaLogger(id)

}

func HandleLambdaEvent(request types.SpinupRequest) (types.SpinupResponse, error) {

	// SpinupRequest information:
	log.Infof("Invoke Arguments \t (tick: %d) \t (OrchEndpoint: %s) \t (GatewayEndpoint: %s) \t (key: %s)", request.Tick, request.OrchEndpoint, request.GatewayEndpoint, request.Key)

	// setup Runtime
	runtime := r.CreateOrGetRuntime(request.Tick)

	// setup AWS session
	awsutils.CreateLambdaSession()

	// fetch object (request.Key) if not already cached
	if _, found := gatewayComm.RespMap.Load(request.Key); !found {
		// object not cached, GET from S3
		body := awsutils.GetObject(request.Key)
		if body != nil {
			log.Infof("downloaded/added object (%s) to response mapping", request.Key)
			gatewayComm.RespMap.Store(request.Key, body)
			gatewayComm.NumberOfObjects++
		}
	}

	// check how many object are served (response mapping),
	// if 0, return immediately nothing to do for lambda-runtime
	if gatewayComm.NumberOfObjects == 0 {
		log.Warn("Lambda-Runtime has no object to serve, exit")
		return types.BuildResponse([]string{}, "no key to serve"), nil
	}

	// setup connections (dial...)
	gatewayComm.SetupConnections(request.GatewayEndpoint)
	orchComm.SetupConnections(request.OrchEndpoint)
	log.Infof("Connection to gateway (%v) and orchestrator (%v) established", gatewayComm.GatewayClient.GetRemoteAddr(), orchComm.OrchClient.GetRemoteAddr())

	// provide server functionality to serve requests for the object
	gatewayComm.ServeConnections(gatewaySrv)
	orchComm.ServeConnections(orchSrv)

	// inform gateway about the served objects
	keyList := make([]string, gatewayComm.NumberOfObjects)
	i := 0
	gatewayComm.RespMap.Range(func(key, value interface{}) bool {
		keyList[i] = key.(string)
		i++
		return true
	})

	// notify gateway and orchestrator that runtime is running and ready
	gatewayComm.LambdaStart(gatewayComm.GatewayClient, keyList)
	// orchComm.LambdaStart(orchComm.OrchClient, "Lambda->Orch Test")

	// remains until timer expires or an event is triggered to shut down the lambda-runtime
	log.Info("everything setup, wait and serve")
	return Wait(runtime, keyList)

}

func Wait(runtime *r.Runtime, keys []string) (types.SpinupResponse, error) {
	select {
	case <-runtime.WaitDone():
		// There's no turning back.
		log.Warn("manual shutdown requested")
		gatewayComm.LambdaBye(gatewayComm.GatewayClient, keys)
		log.Info("wait for GatewayByeAck")

		// wait for LambdaBye ACK by the Gateway (still serving in-flight Get requests)
		<-runtime.Timeout.S()

		log.Info("Done and GatewayByeAck received, shutdown lambda-runtime")
		return types.BuildResponse(keys, "manual shutdown"), nil
	case <-runtime.Timeout.C():
		// There's no turning back.
		log.Warn("timer expired")
		gatewayComm.LambdaBye(gatewayComm.GatewayClient, keys)
		log.Info("wait for GatewayByeAck")

		// wait for LambdaBye ACK by the Gateway (still serving in-flight Get requests)
		<-runtime.Timeout.S()

		log.Info("timer expired and GatewayByeAck received, shutdown lambda-runtime")
		return types.BuildResponse(keys, "timer expired"), nil
	}
}

func main() {
	log.Info("Cold Startup")

	// services for gateway:
	gatewaySrv.HandleFunc("get", gatewayComm.HandlerGet)
	gatewaySrv.HandleFunc("gateayByeAck", gatewayComm.HandlerGatewayByeAck)

	// services for orchestrator:
	orchSrv.HandleFunc("info", orchComm.HandlerInfo)
	orchSrv.HandleFunc("keepRunning", orchComm.HandlerKeepRunning)
	orchSrv.HandleFunc("shutdown", orchComm.HandlerShutdown)

	lambda.Start(HandleLambdaEvent)
}
