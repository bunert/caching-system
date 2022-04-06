package main

/*
Running the Orchestrator Server,  which forwards routing decisions to the gateway servers.
Invokes lambda-runtimes when needed and keeps track of those sessions.
Receives feedback from the gateway servers fo the decision making process.

- IP is retrieved by hardcoded instance ID for the corresponding AWS EC2 instance
*/

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/utility/logger"
	"cloud-storage/core/utility/network"
	"cloud-storage/orchestrator"
	"cloud-storage/orchestrator/control/redisControl"
	"cloud-storage/orchestrator/objectManager"
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

	// set thresholds for ObjectManager
	objectManager.LoadConfiguration()

	// setup AWS session
	awsutils.SetupOrchestratorSession()

	network.InitNetworkingOrchestrator()

	// EC2 Redis instance setup
	awsutils.SetupEC2()
	// setup Redis client if instance running
	redisControl.SetupRedis()

	orchestrator.Run()
	log.Error("Orchestrator exited")

}
