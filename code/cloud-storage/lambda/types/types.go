package types

import "net/http"

type SpinupRequest struct {
	Key             string `json:"key"`
	Tick            int    `json:"tick"`
	OrchEndpoint    string `json:"orchEndpoint"`
	GatewayEndpoint string `json:"gatewayEndpoint"`
}

type SpinupResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type SpinupResponseBody struct {
	Keys    []string `json:"keys"`
	Message string   `json:"message"`
}

type SpinupResponse struct {
	StatusCode int                   `json:"statusCode"`
	Headers    SpinupResponseHeaders `json:"headers"`
	Body       SpinupResponseBody    `json:"body"`
}

func BuildResponse(keys []string, msg string) (response SpinupResponse) {
	header := SpinupResponseHeaders{
		ContentType: "application/json",
	}

	body := SpinupResponseBody{
		Keys:    keys,
		Message: msg,
	}

	response = SpinupResponse{
		StatusCode: http.StatusOK,
		Headers:    header,
		Body:       body,
	}

	return
}
