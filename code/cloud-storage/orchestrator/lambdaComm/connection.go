package lambdaComm

import (
	"cloud-storage/core/lambdaClient"
	"net"
)

var (
	LambdaClient *lambdaClient.Client
)

type Server struct {
	ready chan struct{}
}

// setup server and use ready channel to notify if ready
func NewServer() *Server {
	s := &Server{
		ready: make(chan struct{}),
	}

	close(s.ready)

	return s
}

func (s *Server) Serve(lis net.Listener) {
	for {
		cn, err := lis.Accept()
		if err != nil {
			return
		}

		log.Debug("new Lambda connection accepted, call ServeLambda")
		LambdaClient = lambdaClient.NewClientWithConnection(cn)
		go ServeLambda(LambdaClient)
	}
}

func (s *Server) Ready() chan struct{} {
	return s.ready
}

func (s *Server) Close(lis net.Listener) {
	lis.Close()
}
