package server

import (
	"cloud-storage/core/client"
	"cloud-storage/core/utility/globals"
	"cloud-storage/orchestrator/gatewayComm"
	"cloud-storage/orchestrator/objectManager"
	"strings"
	"time"

	"github.com/mason-leap-lab/redeo/resp"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "gatewayComm/server")
)

func HandleGatewayStarted(w resp.ResponseWriter, c *resp.Command) {
	var addr string = c.Arg(0).String()         // address of Gateway Endpoint for Orchestrator to connect to
	globals.GatewayEndpoint = c.Arg(1).String() // address of Gateway Endpoint for Lambda-Runtime to connect to

	log.Info("GatewayStarted, initialize connection")

	// received GatewayStarted from Gateway, use addr to inititate a connection to the gateway
	gatewayComm.GatewayClient = client.NewClient()
	if !gatewayComm.GatewayClient.Dial(addr) {
		log.Fatal("Dial for gateway client failed, check network settings")
	}

	log.Info(strings.Repeat("=", 20))
	log.Info("System is ready.")
	log.Info(strings.Repeat("=", 20))

}

// handler blocking, so no concurrency unless goroutine used
func HandleReceivedGet(w resp.ResponseWriter, c *resp.Command) {
	var key string = c.Arg(0).String()
	var endpoint string = c.Arg(1).String()
	var timestampStr string = c.Arg(2).String()

	log.Debugf("|%s %3s %s| %13v |%s %-6s %s| %s",
		"\033[97;42m", "GET", "\033[0m",
		timestampStr,
		"\033[97;44m", endpoint, "\033[0m",
		key,
	)

	timestamp, err := time.Parse("2006-01-02 15:04:05.000", timestampStr)
	if err != nil {
		log.Panicf("Parsing timestamp failed: %s", timestampStr)
	}
	objectManager.ProcessGet(key, endpoint, timestamp)

}
