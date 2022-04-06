package client

import (
	"cloud-storage/core/client"
	"strconv"

	"github.com/mason-leap-lab/redeo/resp"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "gatewayComm/client")
)

// requests forwarding table from the Gateway
func LogGatewayState(c *client.Client) {
	c.Conn.W.WriteCmdString("logGatewayState")

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.Warn("Failed to send logGatewayState")
		return
	}
}

// notify Gateway that a given EC2 Redis instance is running
func RedisStart(c *client.Client, addr string) {

	// Send request and wait
	c.Conn.W.WriteCmdString("redisStart", addr)

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.Warn("Failed to send redisStart")
		return
	}
}

// notify Gateway that a given EC2 Redis instance is shut down
// waits for the ACK by the Gateway to ensure ongoing Redis Gets are still served.
func RedisStop(c *client.Client) {

	// Send request and wait
	c.Conn.W.WriteCmdString("redisStop")

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.Warn("Failed to send redisStop")
		return
	}

	// wait until gateway responds with ACK to process all inflight messages before returning
	type0, err := c.Conn.R.PeekType()
	if err != nil {
		log.Panic("RedisStop, PeekType error on receiving Gateway ACK: %v", err)
	}
	switch type0 {
	case resp.TypeError:
		strErr, err := c.Conn.R.ReadError()
		if err == nil {
			log.Panic(strErr)
		}
		log.Panic(err)
		return
	case resp.TypeInline:
		break
	default:
		log.Panicf("RedisStop, unhandled resp type: %s", type0.String())
	}

	msg, err := c.Conn.R.ReadInlineString()
	if err != nil {
		log.Panic(err)
	}
	if msg != "OK" {
		log.Panicf("RedisStop: wrong inline string? (msg: %s)", msg)
	}

}

// updates the Gateway about the Redis state for a given key
func RedisUpdate(c *client.Client, state string, keys []string) {

	len := len(keys)
	if len == 0 {
		log.Debug("RedisUpdate but no keys, don't send message")
		return
	}

	// cmd, state, length and each key
	c.Conn.W.WriteMultiBulkSize(len + 3)
	c.Conn.W.WriteBulkString("redisUpdate")
	c.Conn.W.WriteBulkString(state)
	c.Conn.W.WriteBulkString(strconv.Itoa(len))

	for _, key := range keys {
		c.Conn.W.WriteBulkString(key)
	}

	// Flush pipeline
	if err := c.Conn.W.Flush(); err != nil {
		log.Warn("Failed to send redisUpdate")
		return
	}
}
