package gatewayComm

import (
	"cloud-storage/core/client"
	"io"

	"github.com/mason-leap-lab/redeo"
)

var (
	GatewayClient *client.Client
)

func SetupConnections(addr string) {
	if GatewayClient == nil {
		// setup gateway client
		GatewayClient = client.NewClient()
		log.Info("Create new GatewayClient")
		if !GatewayClient.Dial(addr) {
			log.Warn("No connection to gateway")
			return
		}
		return
	}

	log.Info("Gateway Client already present")
	if GatewayClient.Conn == nil {
		log.Info("GatewayClient no connection")
		if !GatewayClient.Dial(addr) {
			log.Warn("No connection to Gateway")
			return
		}
		return
	}

	log.Info("GatewayClient connection already present")

}

func ServeConnections(srv *redeo.Server) {
	// Serve gateway connection
	go func() {
		log.Debug("serve Foreign Client on Gateway Connection")
		err := srv.ServeForeignClient(GatewayClient.GetConn())
		if err != nil && err != io.EOF {
			log.Info("Connection closed: ", err)
		} else {
			log.Info("Connection closed.")
		}
		GatewayClient.Close()
	}()
}
