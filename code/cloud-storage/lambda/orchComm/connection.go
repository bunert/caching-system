package orchComm

import (
	"cloud-storage/core/client"
	"io"

	"github.com/mason-leap-lab/redeo"
)

var (
	OrchClient *client.Client
)

func SetupConnections(addr string) {
	if OrchClient == nil {
		// setup gateway client
		OrchClient = client.NewClient()
		log.Info("Create new OrchClient")
		if !OrchClient.Dial(addr) {
			log.Fatal("No connection to Orchestrator")
			return
		}
		return
	}

	log.Info("Orchestrator Client already present")
	if OrchClient.Conn == nil {
		log.Info("OrchClient no connection")
		if !OrchClient.Dial(addr) {
			log.Fatal("No connection to Orchestrator possible")
			return
		}
		return
	}

	log.Info("OrchClient connection already present")
}

func ServeConnections(srv *redeo.Server) {
	// Serve orchestrator connection
	go func() {
		log.Debug("serve Foreign Client on Orchestrator Connection")
		err := srv.ServeForeignClient(OrchClient.GetConn())
		if err != nil && err != io.EOF {
			log.Info("Connection closed: ", err)
		} else {
			log.Info("Connection closed.")
		}
		OrchClient.Close()
	}()
}
