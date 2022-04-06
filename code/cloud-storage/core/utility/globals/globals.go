package globals

import "os"

var (
	// hardcoded for initial gateway->orchestrator communication
	OrchFixPort uint
	OrchFixHost string

	// Listener ports
	OrchPortLis    uint // gateway is listening on this port for orchestrator
	LambdaPortLis  uint // used by both gateway and orchestrator for lambdaComm
	GatewayPortLis uint // orchestrator is listening on this port for gateway

	// External addresses for communication
	HostName string // private EC2 instance IP address

	// specific addresses for specific usage
	SwaggerIP       string // swapper API interface IP, public IPv4 address of the EC2 instance
	SwaggerPort     uint   // swapper API interface Port
	GatewayEndpoint string // used by orchestrator to safe gateway address and hand it over to the lambda-runtime at startup for initial lambda->gateway communication

	LogFile *os.File
)
