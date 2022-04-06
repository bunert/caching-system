package orchComm

import (
	"github.com/mason-leap-lab/redeo/resp"
)

type Request struct {
	Writer *resp.RequestWriter

	Cmd string
	Msg string
}

func (r *Request) Send() {
	r.Writer.WriteBulkString(r.Cmd)
	r.Writer.WriteBulkString(r.Msg)

	// Flush pipeline
	if err := r.Writer.Flush(); err != nil {
		log.Info("Error on flush: ", err)
		return
	}

}

type Response struct {
	Writer resp.ResponseWriter

	Cmd string
	Msg string
}

func (r *Response) Send() {
	r.Writer.AppendBulkString(r.Cmd)
	r.Writer.AppendBulkString(r.Msg)

	// Flush pipeline
	if err := r.Writer.Flush(); err != nil {
		log.Info("Error on flush: ", err)
		return
	}

}
