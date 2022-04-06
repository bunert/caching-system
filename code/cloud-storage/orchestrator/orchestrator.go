package orchestrator

import (
	"cloud-storage/core/awsutils"
	g "cloud-storage/core/utility/globals"
	"cloud-storage/orchestrator/gatewayComm/server"
	"cloud-storage/orchestrator/ginAPI"
	"cloud-storage/orchestrator/lambdaComm"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mason-leap-lab/redeo"

	log "github.com/sirupsen/logrus"
)

func createLogFile() {
	// open output file
	fo, err := os.Create("ec2-redis.txt")
	if err != nil {
		log.Panic(err)
	}
	g.LogFile = fo

}

func closeLogFile() {
	if err := g.LogFile.Close(); err != nil {
		log.Panic(err)
	}

}

func Run() {
	done := make(chan struct{}, 1)

	// Register signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

	// setup GIN API
	go ginAPI.RunControlAPI()

	// Gateway Listener
	gatewayLis, err := net.Listen("tcp", fmt.Sprintf(":%v", g.GatewayPortLis))
	if err != nil {
		log.Errorf("Failed to listen Gateway: %v", err)
		os.Exit(1)
		return
	}

	// Lambda Listener
	lambdaLis, err := net.Listen("tcp", fmt.Sprintf(":%v", g.LambdaPortLis))
	if err != nil {
		log.Errorf("Failed to listen Lambda: %v", err)
		os.Exit(1)
		return
	}

	createLogFile()

	log.Info("Start listening to gateway (port ", g.GatewayPortLis, ") and lamda-runtime (port ", g.LambdaPortLis, ")")
	gatewaySrv := redeo.NewServer(nil)
	lambdaSrv := lambdaComm.NewServer()

	gatewaySrv.HandleFunc("gatewayStarted", server.HandleGatewayStarted)
	gatewaySrv.HandleFunc("receivedGet", server.HandleReceivedGet)

	log.Info("Serve Lambda Listener (port ", g.LambdaPortLis, ")")
	go lambdaSrv.Serve(lambdaLis)
	<-lambdaSrv.Ready()

	go func() {
		<-sig
		log.Info("Receive signal, killing Orchestrator...")
		close(sig)

		// Close server
		log.Info("Closing server...")
		gatewaySrv.Close(gatewayLis)

		// Close Lambda server
		lambdaSrv.Close(lambdaLis)

		log.Info("Closing EC2-Redis Log File...")
		closeLogFile()

		// shutdown Redis if running
		log.Info("Stopping Redis instance...")
		err = awsutils.StopEC2Instance(nil)

		close(done)
	}()

	log.Info("Serve Gateway Listener (port ", g.GatewayPortLis, ")")
	err = gatewaySrv.Serve(gatewayLis)
	if err != nil {
		select {
		case <-sig:
			// Normal close
		default:
			log.WithError(err).Error("Error on serve gateway:")
		}
		gatewaySrv.Release()
	}
	log.Info("Orchestrator closed.")

	<-done
	os.Exit(0)

}
