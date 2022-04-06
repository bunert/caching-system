package awsutils

import (
	g "cloud-storage/core/utility/globals"
	"cloud-storage/lambda/types"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var (
	tick int // TICK for lambda-runtime timeout functionality (in seconds)
)

// initialized in ObjectManager at the beginning
func LambdaSetTick(t int) {
	tick = t
}

// helper function for Orchestrator to spinup a lambda function
// blocks until lambda function returned
func OrchestratorStartLambda(key string) {
	orchEndpoint := fmt.Sprintf("%s:%v", g.HostName, g.LambdaPortLis)
	gatewayEndpoint := g.GatewayEndpoint

	SpinupLambda(key, orchEndpoint, gatewayEndpoint)
}

// invokes a lambda function synchronous, use this function inside a goroutine
func SpinupLambda(key string, orchEndpoint string, gatewayEndpoint string) {
	log.Debug("SpinupLambda called")

	functionName := "get-s3-object" // TODO: hardcoded function name, change later?

	request := types.SpinupRequest{
		Key:             key,
		Tick:            tick,
		OrchEndpoint:    orchEndpoint,
		GatewayEndpoint: gatewayEndpoint,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		log.Fatal("marshalling ", functionName, " request failed: ", err)
	}

	// function is invoked synchronous, waiting for response, therefore SpinupLambda is called as goroutine
	result, err := LambdaSession.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: payload})
	if err != nil {
		log.Fatal("calling ", functionName, " failed: ", err)
	}

	var resp types.SpinupResponse

	err = json.Unmarshal(result.Payload, &resp)
	if err != nil {
		log.Fatal("unmarshalling ", functionName, " reponse failed: ", err)
	}

	// If the status code is NOT 200, the call failed
	if resp.StatusCode != 200 {
		log.WithError(err).Error("invoking lambda function failed, Statuscode: ", strconv.Itoa(resp.StatusCode))
	}

	// TODO: response body includes all the keys the function served, not relevant for now...
	// for _, elem := range resp.Body.Keys {
	// }

	log.Debug("lamda-runtime shut-down: ", resp.Body.Message)
}

// helper function to list all available lambda functions
func ListAllFunctions() {
	result, err := LambdaSession.ListFunctions(nil)
	if err != nil {
		log.Fatal("cannot list functions: ", err)
	}

	log.Info("Functions:")

	for _, f := range result.Functions {
		log.Info("Name:        " + aws.StringValue(f.FunctionName))
		log.Info("Description: " + aws.StringValue(f.Description))
		log.Info("")
	}
}
