package gateway

import (
	"cloud-storage/core/client"
	g "cloud-storage/core/utility/globals"
	"cloud-storage/gateway/lambdaComm"
	"cloud-storage/gateway/orchestratorComm"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"cloud-storage/gateway/ginAPI"

	"github.com/mason-leap-lab/redeo"
)

var ()

func Run() {
	done := make(chan struct{}, 1)

	// Register signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

	// // initiate AWS session
	// awsutils.ConnectAws()

	// Orchestrator Listener
	orchLis, err := net.Listen("tcp", fmt.Sprintf(":%v", g.OrchPortLis))
	if err != nil {
		log.Error("Failed to listen Orchestrator: %v", err)
		os.Exit(1)
		return
	}

	// Lambda Listener
	lambdaLis, err := net.Listen("tcp", fmt.Sprintf(":%v", g.LambdaPortLis))
	if err != nil {
		log.Error("Failed to listen Lambda: %v", err)
		os.Exit(1)
		return
	}

	log.Info("Start listening to orchestrator(port ", g.OrchPortLis, ") and lamda-runtime (port ", g.LambdaPortLis, ")")

	// setup servers
	orchSrv := redeo.NewServer(nil)
	lambdaSrv := lambdaComm.NewServer()

	// setup Orchestrator server function handlers
	// Redis related stuff:
	orchSrv.HandleFunc("redisStart", orchestratorComm.HandleRedisStart)
	orchSrv.HandleFunc("redisStop", orchestratorComm.HandleRedisStop)
	orchSrv.HandleFunc("redisUpdate", orchestratorComm.HandleRedisUpdate)
	// specific handlers (additional handlers for debugging):
	orchSrv.HandleFunc("logGatewayState", orchestratorComm.HandleLogGatewayState)

	// serve Lambda Server
	log.Info("Serve Lambda Listener (port ", g.LambdaPortLis, ")")
	go lambdaSrv.Serve(lambdaLis)
	<-lambdaSrv.Ready()

	// setup Client for Orchestrator using a fixed address
	orchestratorComm.OrchClient = client.NewClient()
	if !orchestratorComm.OrchClient.Dial(fmt.Sprintf("%s:%v", g.OrchFixHost, g.OrchFixPort)) {
		log.Fatal("initial connection to Orchestrator failed, Orchestrator required to be running before Gateway is started")
	}
	// inform Orchestrator that Gateway is running
	// sends Gateway addr, such that Orchestrator can establish a second connection for Orch->Gateway requests
	var orchEndpoint string = fmt.Sprintf("%s:%v", g.HostName, g.OrchPortLis)
	var lambdaEndpoint string = fmt.Sprintf("%s:%v", g.HostName, g.LambdaPortLis)
	go orchestratorComm.GatewayStarted(orchestratorComm.OrchClient, orchEndpoint, lambdaEndpoint)

	go func() {
		<-sig
		log.Info("Receive signal, killing Gateway...")
		close(sig)

		// Close Orchestrator server if running
		log.Info("Closing server...")
		orchSrv.Close(orchLis)

		// Close Lambda server
		lambdaSrv.Close(lambdaLis)

		close(done)
	}()

	// setup GIN API
	go ginAPI.RunAPI()

	log.Info("Serve Orchestrator Listener (port ", g.OrchPortLis, ")")
	err = orchSrv.Serve(orchLis)
	if err != nil {
		select {
		case <-sig:
			// Normal close
		default:
			log.WithError(err).Error("Error on serve orchestrator:")
		}
		orchSrv.Release()
	}

	log.Info("Gateway closed.")

	<-done
	os.Exit(0)

}
