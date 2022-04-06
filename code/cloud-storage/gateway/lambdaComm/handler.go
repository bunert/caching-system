package lambdaComm

import (
	"cloud-storage/core/lambdaClient"
	"cloud-storage/gateway/global"
	t "cloud-storage/lambda/gatewayComm/types"
	"io"
	"strconv"

	"github.com/mason-leap-lab/redeo/resp"
	logrus "github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("component", "lambdaComm")
	// loc, _ = time.LoadLocation("Europe/Berlin")
)

// receives the object key the lambda-runtime is running for, sets the mapping to "lambda" for future requests
func HandleLambdaStart(r resp.ResponseReader) {

	lenString, err := r.ReadBulkString()
	if err != nil {
		log.Fatal("ReadBulkString should not fail")
	}

	len, err := strconv.Atoi(lenString)
	if err != nil {
		log.Fatalf("could not convert %s to int", lenString)
	}
	log.Debugf("lambdaStart, update ReqMap (number of keys: %d):", len)
	for i := 1; i <= len; i++ {
		key, err := r.ReadBulkString()
		if err != nil {
			log.Fatal("ReadBulkString should not fail")
		}
		global.ReqMap.Store(key, "lambda")
		// log.Warnf("Lambda %d", time.Now().In(loc).UnixMilli())

		log.Debugf("\t\t%-20s%10s", key, "lambda")
	}

}

// receives the object key the lambda-runtime was running for, removes the mapping
func HandleLambdaBye(r resp.ResponseReader) {

	lenString, err := r.ReadBulkString()
	if err != nil {
		log.Fatal("ReadBulkString should not fail")
	}

	len, err := strconv.Atoi(lenString)
	if err != nil {
		log.Fatalf("could not convert %s to int", lenString)
	}
	log.Debugf("lambdaBye, update ReqMap (number of keys: %d):", len)
	for i := 1; i <= len; i++ {
		key, err := r.ReadBulkString()
		if err != nil {
			log.Fatal("ReadBulkString should not fail")
		}
		val, found := global.ReqMap.Load(key)
		if found && val == "lambda" {
			global.ReqMap.Delete(key)
		}

		log.Debugf("\t\t%-20s%10s", key, "removed")
	}

	// if lambdaBye received, trigger goroutine claiming the Lambda Write Lock
	// if claimed, in-flight Lambda Get request -> send GatewayByeAck to Lambda-Runtime to continue shutdown
	go func() {
		// log.Warn("HandleLambdaBye, claiming write lock")
		global.LambdaLock.Lock()
		// log.Warn("Lambda write lock claimed")
		defer global.LambdaLock.Unlock()

		GatwayByeAck()
		// log.Warn("GatewayByeAck sent")
	}()

}

// receives the Get response from the lambda-runtime (content)
// pushes it on the specific channel specified by the reqId for the GinAPI routine to retrieve and forward to the client
func HandleGetResp(r resp.ResponseReader) error {

	// reqID required, to forward resp to the correct goroutine handling the specific Get request
	reqId, err := r.ReadBulkString()
	if err != nil {
		return err
	}
	contentType, err := r.ReadBulkString()
	if err != nil {
		return err
	}
	contentLength, err := r.ReadBulkString()
	if err != nil {
		return err
	}
	msg, err := r.ReadBulkString()
	if err != nil {
		return err
	}

	// log.Debugf("GetResp (reqId: %s)", reqId)
	val, found := global.RespChan.Load(reqId)
	channel := val.(chan t.Body)
	if found {
		channel <- t.Body{
			ContentType:   contentType,
			ContentLength: contentLength,
			Msg:           msg,
		}
	} else {
		// should never happen...
		log.Panic("response channel not found?")
	}

	return nil
}

// serves lambda requests/responses
// not possible as srv.HandleFunc, those connections are established by the lambda-runtime and reused for requests
// therefore for every incoming lambda-runtime connection ServeLambda is called
func ServeLambda(client *lambdaClient.Client) {
	for {
		go client.Conn.PeekResponse()

		retPeek := client.Conn.WaitForType()
		if retPeek == nil {
			return
		}

		var respType resp.ResponseType
		respType, err := client.Conn.CheckType(retPeek)
		if err != nil {
			return
		}

		switch respType {
		case resp.TypeError:
			client.Conn.HandleErrorType()

		default:
			cmd, err := client.Conn.R.ReadBulkString()
			if err != nil && err == io.EOF {
				log.Warn("Lambda disconnected")
				client.Conn.Close()
			} else if err != nil {
				log.Warn("Error on reading lambda command type: ", err)
				break
			}

			switch cmd {
			case "lambdaStart":
				// lambda-runtime started message (sets mapping to "lambda")
				HandleLambdaStart(client.Conn.R)
			case "lambdaBye":
				// lambda-runtime bye message (removes mapping if value is "lambda", if overwritten by Redis, ignore it)
				HandleLambdaBye(client.Conn.R)
			case "getResp":
				// handle response received by the lambda-runtime
				err := HandleGetResp(client.Conn.R)
				if err != nil {
					log.WithError(err).Error("HandleGetResp failed")
					log.Fatal("should not happen, exit")
				}
			default:
				log.Warnf("Unsupported response type: %s", cmd)
			}
		}
	}
}
