package types

import (
	"strconv"

	"github.com/mason-leap-lab/redeo/resp"
	logrus "github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "gatewayComm/types")
)

type Request struct {
	Writer *resp.RequestWriter

	Cmd  string
	Keys []string
}

func (r *Request) Send() {
	len := len(r.Keys)

	// cmd, length and each key
	// r.Writer.WriteMultiBulkSize(len + 2)
	r.Writer.WriteBulkString(r.Cmd)
	r.Writer.WriteBulkString(strconv.Itoa(len))

	for _, key := range r.Keys {
		r.Writer.WriteBulkString(key)
	}

	// Flush pipeline
	if err := r.Writer.Flush(); err != nil {
		log.Info("Error on flush: ", err)
		return
	}
	log.Debugf("Sent: %s", r.Cmd)

}

type Body struct {
	ContentType   string
	ContentLength string
	Msg           string
}

type Response struct {
	Writer resp.ResponseWriter

	Cmd   string
	ReqId string
	Body  Body
}

func (r *Response) Send() {

	r.Writer.AppendBulkString(r.Cmd)
	r.Writer.AppendBulkString(r.ReqId)
	r.Writer.AppendBulkString(r.Body.ContentType)
	r.Writer.AppendBulkString(r.Body.ContentLength)
	r.Writer.AppendBulkString(r.Body.Msg)

	// Flush pipeline
	if err := r.Writer.Flush(); err != nil {
		log.Info("Error on flush: ", err)
		return
	}
	log.Debugf("Sent: %s", r.Cmd)

}
