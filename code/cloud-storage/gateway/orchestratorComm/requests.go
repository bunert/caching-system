package orchestratorComm

import (
	"cloud-storage/core/client"
	"time"
)

var (
	OrchClient *client.Client
)

//
func GatewayStarted(c *client.Client, addrOrch string, addrLambda string) {

	// Send request and wait
	c.Conn.W.WriteCmdString("gatewayStarted", addrOrch, addrLambda)

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.Warn("Failed to send gatewayStarted to Orchestrator")
		return
	}
}

func ReceivedGet(c *client.Client, key string, endpoint string) {

	// Send request and wait
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05.000")
	c.Conn.W.WriteCmdString("receivedGet", key, endpoint, timestamp)

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.Warn("Failed to flush ReceivedGet")
		return
	}

}
