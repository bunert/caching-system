package client

import (
	"io"
	"net"

	"github.com/mason-leap-lab/redeo/resp"

	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "client")
)

type Conn struct {
	conn net.Conn
	W    *resp.RequestWriter
	R    resp.ResponseReader
}

func (c *Conn) Close() {
	c.conn.Close()
}

type Client struct {
	Conn    *Conn
	Address string
}

// new client when dialing is explicitly called afterwards
func NewClient() *Client {
	return &Client{
		Conn: nil,
	}
}

func (c *Client) Dial(addr string) bool {
	c.Address = addr

	log.Info("Dialing ", addr)
	if err := c.connect(); err != nil {
		log.WithError(err).Error("Failed to dial ", addr)
		c.Close()
		return false
	}

	return true
}

func (c *Client) connect() error {
	cn, err := net.Dial("tcp", c.Address)
	if err != nil {
		return err
	}
	c.Conn = &Conn{
		conn: cn,
		W:    NewRequestWriter(cn),
		R:    NewResponseReader(cn),
	}
	return nil
}

func (c *Client) Close() {
	log.Info("Cleaning up...")
	if c.Conn != nil {
		c.Conn.Close()
		c.Conn = nil
	}
	log.Info("Client closed.")
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
