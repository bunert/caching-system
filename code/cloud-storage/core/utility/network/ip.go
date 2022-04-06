package network

/*
ip.go is responsible to initialize the connection addresses correctly.

hardcoded port assignment:
- 3000 Swagger Port for Orchestrator
- 4000 Swagger Port for Gateway
- 5000 Listener Port on Gateway for communication with Orchestrator
- 6000 Listener Port on Orchestrator for communication with Gateway
- 7000 Listener Port on Orchestrator for communication with Lambda-Runtime
- 8000 Listener Port on Gateway for communication with Lambda-Runtime

special variables:
- OrchFixPort: Gateway requires the listener port of the Orchestrator for the initial communication
- OrchFixHost: Gateway requires the IP addr of the Orchestrator for the initial communication (depends on deployment location)
- Instance IDs: the EC2 instance IDs of the orchestrator and gateway EC2 isntances is hardcoded and used to retrieve the public IP addresses (changes each startup)
- GatewayEndpoint: retrieved in the initial message by the orchestrator and used later to inform the Lambda-Runtime about the address of the Gateway

*/

import (
	"cloud-storage/core/awsutils"
	"cloud-storage/core/utility/globals"
	"errors"
	"net"

	logrus "github.com/sirupsen/logrus"
)

var (
	ErrPrivateIPNotFound = errors.New("no private IP found")
	log                  = logrus.WithField("component", "IP")

	// EC2 instance IDs
	OrchInstandeId    = "i-01f8420ccf3e6927e"
	GatewayInstanceId = "i-02ef4b6ac93e84613"
)

func InitNetworkingGateway() {
	globals.OrchFixPort = uint(6000) // fixed port where Orchestrator is listening for gateway communication

	globals.SwaggerPort = uint(4000)

	// listener ports
	globals.OrchPortLis = uint(5000)
	globals.LambdaPortLis = uint(8000)

	ip, err := GetPrivateIp()
	if err != nil {
		log.Fatal(err)
	}
	globals.HostName = ip

	_, orchAddr_private, err := awsutils.RetrieveIP(OrchInstandeId)
	if err != nil {
		log.Panic("could not retrieve Public IP Orchestrator instance: %s", OrchInstandeId)
		return
	}
	globals.OrchFixHost = orchAddr_private

	gatewayAddr, _, err := awsutils.RetrieveIP(GatewayInstanceId)
	if err != nil {
		log.Panic("could not retrieve Public IP for Gateway instance: %s", GatewayInstanceId)
		return
	}

	globals.SwaggerIP = gatewayAddr // swagger API reachable from outside the network

	log.Infof("Starting Gateway on EC2 instance (private: %s, public: %s)", globals.HostName, gatewayAddr)

}

func InitNetworkingOrchestrator() {

	globals.SwaggerPort = uint(3000)

	// listener ports
	globals.GatewayPortLis = uint(6000)
	globals.LambdaPortLis = uint(7000)

	ip, err := GetPrivateIp()
	if err != nil {
		log.Fatal(err)
	}
	globals.HostName = ip

	orchAddr, _, err := awsutils.RetrieveIP(OrchInstandeId)
	if err != nil {
		log.Panic("could not retrieve Public IP for Orchestrator instance: %s", OrchInstandeId)
		return
	}

	globals.SwaggerIP = orchAddr // swagger API reachable from outside the network

	log.Infof("Starting Orchestrator on EC2 instance (private: %s, public: %s)", globals.HostName, orchAddr)

}

func GetPrivateIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && IsPrivateIp(ipnet.IP) {
			return ipnet.IP.String(), nil
		}
	}

	return "", ErrPrivateIPNotFound
}

func IsPrivateIp(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsMulticast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 {
			return true
		} else if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		} else if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
	}
	return false
}
