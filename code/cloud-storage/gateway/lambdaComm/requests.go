package lambdaComm

import (
	"cloud-storage/gateway/global"
	t "cloud-storage/lambda/gatewayComm/types"
	"net/http"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func Get(c *gin.Context, key string) {
	global.LambdaLock.RLock()
	defer global.LambdaLock.RLocker().Unlock()

	// send request to lambda
	reqId := uuid.New().String()

	channel := make(chan t.Body)
	global.RespChan.Store(reqId, channel)

	LambdaClient.Conn.W.WriteCmdString("get", reqId, key)

	// Flush pipeline
	if err := LambdaClient.Conn.W.Flush(); err != nil {
		log.Warn("Failed to flush get")
		return
	}

	// close channel at the end and remove the mapping
	defer close(channel)
	defer global.RespChan.Delete(reqId)

	// waits on the channel with the correct requestId for the response (added by the handler)
	body := <-channel

	c.Header("Content-Type", body.ContentType)
	c.Header("Content-Origin", "lambda")
	c.Header("Content-Length", body.ContentLength)
	c.Status(http.StatusOK)
	c.Writer.WriteString(body.Msg)

}

func GatwayByeAck() {
	LambdaClient.Conn.W.WriteCmdString("gateayByeAck")

	// Flush pipeline
	if err := LambdaClient.Conn.W.Flush(); err != nil {
		log.Warn("Failed to flush ByeAck")
		return
	}

}
