package orchComm

import (
	"cloud-storage/core/client"
)

func LambdaStart(c *client.Client, msg string) {

	// Send request
	request := &Request{
		Writer: c.Conn.W,
		Cmd:    "lambdaStart",
		Msg:    msg,
	}
	request.Send()
}
