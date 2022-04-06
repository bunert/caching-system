package lambdaClient

/*
Client Library for client on the gateway/orchestrator side of the lambda Connection.

*/

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/mason-leap-lab/redeo/resp"

	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "lambdaClient")

	ErrCheckType = errors.New("type of peeked response of type Error")
)

type Connection struct {
	conn     net.Conn
	W        *resp.RequestWriter
	R        resp.ResponseReader
	mu       sync.Mutex
	NextType chan interface{}
	Closed   chan struct{}
}

type Client struct {
	Conn    *Connection
	Address string
}

// for lambda communication when the connection is obtained by an incoming connection
func NewClientWithConnection(cn net.Conn) *Client {
	log.Debugf("new Lambda Client (remoteAddr: %s), call ServeLambda", cn.RemoteAddr().String())
	return &Client{
		Conn: &Connection{
			conn:     cn,
			W:        NewRequestWriter(cn),
			R:        NewResponseReader(cn),
			NextType: make(chan interface{}),
			Closed:   make(chan struct{}),
		},
		Address: cn.RemoteAddr().String(),
	}
}

func NewRequestWriter(wr io.Writer) *resp.RequestWriter {
	return resp.NewRequestWriter(wr)
}
func NewResponseReader(rd io.Reader) resp.ResponseReader {
	return resp.NewResponseReader(rd)
}

func (c *Client) GetConn() net.Conn {
	return c.Conn.conn
}

func (c *Client) GetRemoteAddr() string {
	return c.Conn.conn.RemoteAddr().String()
}

func (conn *Connection) PeekResponse() {
	respType, err := conn.R.PeekType()
	if err != nil {
		conn.NextType <- err
	} else {
		conn.NextType <- respType
	}
}

func (conn *Connection) WaitForType() interface{} {
	var retPeek interface{}
	select {
	case <-conn.Closed:
		log.Info("ServeLambda closed")
		conn.Close()
		return nil
	case retPeek = <-conn.NextType:
		// Got Msg
		return retPeek
	}
}

func (conn *Connection) CheckType(retPeek interface{}) (resp.ResponseType, error) {
	var respType resp.ResponseType
	var err error = ErrCheckType

	switch retPeekType := retPeek.(type) {
	case error:
		if retPeekType == io.EOF {
			log.Debug("Lambda-Runtime disconnected, lambda function probably removed from memory.")
		} else {
			log.Warnf("Failed to peek response type: %v", retPeekType)
		}
		conn.Close()
	case resp.ResponseType:
		respType = retPeek.(resp.ResponseType)
		err = nil
	}
	return respType, err
}

func (conn *Connection) HandleErrorType() {
	strErr, err := conn.R.ReadError()
	if err != nil {
		err = fmt.Errorf("response error (Unknown): %v", err)
	} else {
		err = fmt.Errorf("response error: %s", strErr)
	}
	log.Warnf("%v", err)
	log.Fatal("resp Type Error case") //TODO: set a corresponding response (necessary? InfiniCache does)
}

func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.Closed:
		// already closed
		return
	default:
	}
	close(c.Closed)
	log.Debug("Close Lambda Connection")

	// Don't use c.Close(), it will stuck and wait for lambda.
	c.conn.(*net.TCPConn).SetLinger(0) // The operating system discards any unsent or unacknowledged data.
	c.conn.Close()
	log.Debug("Closed.")
}
