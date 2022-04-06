package orchComm

import (
	"cloud-storage/lambda/runtime"
	"time"

	"github.com/mason-leap-lab/redeo/resp"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "orchComm")
)

func HandlerInfo(r resp.ResponseWriter, c *resp.Command) {

	log.Info("In Info handler")

	response := &Response{
		Writer: r,
		Cmd:    "infoResp",
		Msg:    "info response test",
	}
	response.Send()
}

func HandlerKeepRunning(r resp.ResponseWriter, c *resp.Command) {

	log.Info("in KeepRunning handler")
	runtime := runtime.GetRuntime()
	runtime.Timeout.ResetTimer(time.Second * 120)
	runtime.Timeout.Disable()

}

func HandlerShutdown(r resp.ResponseWriter, c *resp.Command) {

	log.Info("in Shutdown handler")
	runtime := runtime.GetRuntime()
	runtime.DoneLocked()

}
