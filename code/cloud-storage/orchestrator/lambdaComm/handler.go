package lambdaComm

import (
	"cloud-storage/core/lambdaClient"
	"io"

	"github.com/mason-leap-lab/redeo/resp"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "lambdaComm")
)

// used for connection establishment if lambda-runtime is started
// sends Info request to lambda-runtime (dummy content)
// TODO: change Info request or delete if not required
func HandleLambdaStart(r resp.ResponseReader) {

	msg, err := r.ReadBulkString()
	if err != nil {
		log.Fatal("ReadBulkString should not fail")
	}

	log.Debugf("lambdaStart: %s", msg)

	// request Info from Lambda-runtime
	// Info(LambdaClient)

}

// simple dummy info response
// TODO: change or delete if not actually required
func HandleInfoResp(r resp.ResponseReader) {

	msg, err := r.ReadBulkString()
	if err != nil {
		log.Fatal("ReadBulkString should not fail")
	}

	log.Debug("InfoResp: ", msg)
}

// serves lambda requests/responses
// not possible as srv.HandleFunc, those connections are established by the lambda-runtime and reused for requests
// therefore for every incoming lambda-runtime connection ServeLambda is called
func ServeLambda(client *lambdaClient.Client) {
	for {
		go client.Conn.PeekResponse()

		retPeek := client.Conn.WaitForType()
		if retPeek == nil {
			return
		}

		var respType resp.ResponseType
		respType, err := client.Conn.CheckType(retPeek)
		if err != nil {
			return
		}

		switch respType {
		case resp.TypeError:
			client.Conn.HandleErrorType()

		default:
			cmd, err := client.Conn.R.ReadBulkString()
			if err != nil && err == io.EOF {
				log.Warn("Lambda disconnected")
				client.Conn.Close()
			} else if err != nil {
				log.Warn("Error on reading lambda command type: ", err)
				break
			}

			switch cmd {
			case "lambdaStart":
				// simple startup notify message received from lambda-runtime at startup for connection establishment
				HandleLambdaStart(client.Conn.R)
			case "infoResp":
				// InfoResp request triggered when LambdaStart message received, simple dummy response so far
				HandleInfoResp(client.Conn.R)
			default:
				log.Warnf("Unsupported response type: %s", cmd)
			}
		}
	}
}
