package lambdaComm

import (
	"cloud-storage/core/lambdaClient"
)

func Info(c *lambdaClient.Client) {

	// Send request and wait
	c.Conn.W.WriteCmdString("info")

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.Warn("Failed to flush Info")
		return
	}

}

func KeepRunning(c *lambdaClient.Client) {

	// Send request and wait
	c.Conn.W.WriteCmdString("keepRunning")

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.WithError(err).Error("failed to flush keepRunning")
		return
	}

}

func Shutdown(c *lambdaClient.Client) {

	// Send request and wait
	c.Conn.W.WriteCmdString("shutdown")

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.WithError(err).Error("failed to flush shutdown")
		return
	}

}
