package gatewayComm

import (
	t "cloud-storage/lambda/gatewayComm/types"
	"cloud-storage/lambda/runtime"

	"github.com/mason-leap-lab/redeo/resp"
	logrus "github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "gatewayComm")
)

func HandlerGet(r resp.ResponseWriter, c *resp.Command) {
	// TODO: extend timer if not in shutting down mode
	runtime := runtime.GetRuntime()
	defer runtime.Timeout.ExtendTimer()

	reqID := c.Arg(0).String()
	key := c.Arg(1).String()

	log.Debugf("in GET handler for key: %s (reqID: %s)", key, reqID)

	val, found := RespMap.Load(key)
	if !found {
		log.Panic("key not present in response map?")
	}
	object := val.(*t.Body)

	response := &t.Response{
		Writer: r,
		Cmd:    "getResp",
		ReqId:  reqID,
		Body:   *object,
	}
	response.Send()
}

func HandlerGatewayByeAck(r resp.ResponseWriter, c *resp.Command) {
	log.Infof("in GatewayByeAck handler")
	runtime := runtime.GetRuntime()
	runtime.Timeout.GatewayAckedShutdown()

}
