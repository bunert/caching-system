package main

/*
Running a Gateway Server acting as API endpoint for client requests.
Forwards the requests to either the storage or the lambda-runtime acting as cache.
Active feedback loop with the orchestrator to keep routing information up-to-date.

- IP is retrieved by hardcoded instance ID for the corresponding AWS EC2 instance
*/

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/utility/logger"
	"cloud-storage/core/utility/network"
	"cloud-storage/gateway"
	"flag"

	log "github.com/sirupsen/logrus"
)

var (
	// cmd-line flag
	isDebug = flag.Bool("debug", false, "specifiy log level (debug/production), default production")
)

func main() {
	flag.Parse()

	logger.SetupLogger(*isDebug)
	awsutils.LoadEnv()

	// initiate AWS session
	awsutils.SetupGatewaySession()

	network.InitNetworkingGateway()

	gateway.Run()

	log.Error("Gateway exited")

}
