package gatewayComm

import (
	"cloud-storage/core/client"
	t "cloud-storage/lambda/gatewayComm/types"
)

func LambdaStart(c *client.Client, keys []string) {

	// Send request
	request := &t.Request{
		Writer: c.Conn.W,
		Cmd:    "lambdaStart",
		Keys:   keys,
	}
	request.Send()
}

func LambdaBye(c *client.Client, keys []string) {

	// Send request
	request := &t.Request{
		Writer: c.Conn.W,
		Cmd:    "lambdaBye",
		Keys:   keys,
	}
	request.Send()
}
