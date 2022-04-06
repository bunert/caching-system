package awsutils

import (
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	ErrRetrieveIp             = errors.New("no IP addresses for EC2 instance available")
	ErrDescribeInstance       = errors.New("describe instance failed")
	ErrInstanceNotAvailable   = errors.New("given Instance does not exists or not running")
	ErrInstanceNotRunning     = errors.New("instance not running")
	ErrInstancePending        = errors.New("instance in pending state")
	ErrInstanceAlreadyRunning = errors.New("instance already running")
	ErrInstanceAlreadyStopped = errors.New("instance already stopped")
	ErrInstanceShuttingDown   = errors.New("instance is shutting down")

	ErrInstanceAlreadyStarting = errors.New("instance is already starting up")

	ErrWrongState = errors.New("EC2 instance not in expected state")

	ErrUnhandledState = errors.New("unhandled instance state")

	redisServerInstance *RedisServerInstance

	mu sync.Mutex
)

// Redis Server Instance to keep track internally.
type RedisServerInstance struct {
	instanceId string
	addr       string
	status     string
}

// Get RedisServerInstance
func getRedisServerInstance() *RedisServerInstance {
	if redisServerInstance == nil {
		log.Panic("redisServerInstance nil?")
	}
	return redisServerInstance
}

// Getter function to retrieve EC's instance public IP addr.
// return addr and nil error if successful
func GetRedisServerAddr() (string, error) {
	if GetRedisServerStatus() != "running" {
		return "", ErrInstanceNotRunning
	}
	return redisServerInstance.addr, nil
}

// Getter function to retrieve EC2's instance status.
// returns the status string
func GetRedisServerStatus() string {
	return getRedisServerInstance().status
}

// Getter function to retrieve EC2's instance ID.
// returns the ID string
func getRedisServerInstanceId() *string {
	if redisServerInstance == nil {
		log.Panic("redisServerInstance nil?")
	}
	return &redisServerInstance.instanceId
}

// Updates the status to the argument.
func updateStatus(status string) {
	if redisServerInstance == nil {
		log.Panic("redisServerInstance nil?")
	}
	redisServerInstance.status = status
}

// Updates the addr to the argument.
func updateAddr(addr string) {
	if redisServerInstance == nil {
		log.Panic("redisServerInstance nil?")
	}
	redisServerInstance.addr = addr
}

// Initial Setup for the EC2 instance.
// Using a hardocded instance ID from the AWS interface.
// Retrieves the status of the instance and updates the RedisServerInstance status.
//
// If instance is running, retrieves the public IP addr and sets it too.
// If instance is stopping, triggers goroutine which waits until instance is stopped and updates the status then.
func SetupEC2() {
	var instanceId string = "i-02c6faeff11583cd8"
	log.Info("Setup EC2 Redis Environment")

	redisServerInstance = &RedisServerInstance{
		instanceId: instanceId,
	}

	status, err := RetrieveStatus()
	if err != nil {
		log.Panic(err)
	}
	updateStatus(status)
	log.Infof("EC2 Redis status: %s", redisServerInstance.status)

	// if running, retrieve addr
	// orchestrator will now also initialize the redis client due to the running instance
	if redisServerInstance.status == "running" {
		log.Info("EC2 Redis running, retrieve addr")
		_, addr, err := RetrieveIP(redisServerInstance.instanceId)
		if err != nil {
			log.Panic(err)
		}
		updateAddr(addr)
	}

	if redisServerInstance.status == "stopping" {
		log.Info("EC2 Redis stopping, notify if stopped")
		go NotifyWhenInstanceStopped(nil)
	}

}

// Starts the EC2 instance.
// Starts the instance if the current status is stopped, otherwhise returns a corresponding error.
// Triggers a goroutine which waits until the instance is running and notifies the argument channel.
// Updates the status field to pending if successfully started.
//
// return an error:
// 	- nil, if successfully started (changes status to pending)
// 	- error otherwhise
func StartEC2Instance(c chan string) error {
	mu.Lock()
	defer mu.Unlock()

	switch GetRedisServerStatus() {
	case "stopped":
		// if stopped, you can start the instance as expected
		break
	case "pending":
		log.Warn("StartEC2Instance, but instance already pending state")
		return ErrInstanceAlreadyStarting
	case "running":
		log.Info("StartEC2Instance, but instance already running")
		return ErrInstanceAlreadyRunning
	case "stopping":
		log.Warn("StartEC2Instance, but instance in stopping state")
		return ErrInstanceShuttingDown
	default:
		log.Panic("StartEC2Instance, unhandled state: %s", redisServerInstance.status)
		return ErrUnhandledState
	}

	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{getRedisServerInstanceId()},
	}

	var status string
	result, err := Ec2Client.StartInstances(input)
	if err != nil {
		log.WithError(err).Error("starting EC2 instance failed")
		return err
	} else {
		status = *(result.StartingInstances[0].CurrentState.Name)
		log.Debugf("started EC2 instance successfully (CurrentState: %s)", status)
	}
	updateStatus(status)

	// goroutine waits asynch. until the instance is actually running to not block this call
	go NotifyWhenInstanceRunning(c)
	return nil
}

// Waits until instance is running and retrieves the public IP of the instance which is sent to the specified channel.
// Updates instance addr and status fields.
func NotifyWhenInstanceRunning(c chan string) {
	describeInstancesInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{getRedisServerInstanceId()},
	}

	if err := Ec2Client.WaitUntilInstanceRunning(describeInstancesInput); err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotReady" {
			log.Warn("WaitUntilInstance failed, probably received a StopEC2 while in pending state")
			return
		}
		log.WithError(err).Error("waiting until Instance is running failed")
		return
	}

	_, addr, err := RetrieveIP(*getRedisServerInstanceId())
	if err != nil {
		log.Panic(err)
	}
	log.Debugf("EC2 instance running (addr: %s)", addr)
	updateAddr(addr)
	updateStatus("running")

	// send addr to channel
	c <- addr
}

// Stops the EC2 instance.
// Stops the instance if the current status is running or pending, otherwhise returns a corresponding error.
// Triggers a goroutine which waits until the instance is stopped which triggers the state update.
// Updates the status field to stopping if successfully stopped which also removes the associated public IP address.
//
// return an error:
// 	- nil, if successfully stopped (changes status to stopping)
// 	- error otherwhise
func StopEC2Instance(c chan struct{}) error {
	mu.Lock()
	defer mu.Unlock()

	switch redisServerInstance.status {
	case "running":
		// if running, you can stop the instance as expected
		break
	case "pending":
		// if starting up, instance gets stopped anyway
		log.Warn("instance is starting, but shutdown requested")
	case "stopped":
		log.Info("StopEC2Instance, but instance already stopped")
		return ErrInstanceAlreadyStopped
	case "stopping":
		log.Info("StopEC2Instance, but instance is already stopping")
		return ErrInstanceShuttingDown
	default:
		log.Error("StopEC2Instance, unhandled state: %s", redisServerInstance.status)
		return ErrUnhandledState
	}

	input := &ec2.StopInstancesInput{
		InstanceIds: []*string{getRedisServerInstanceId()},
	}

	var status string
	result, err := Ec2Client.StopInstances(input)
	if err != nil {
		log.WithError(err).Error("stopping EC2 instance failed")
		return err
	} else {
		status = *(result.StoppingInstances[0].CurrentState.Name)
		log.Debugf("stopped EC2 instance successfully (CurrentState: %s)", status)
	}
	updateStatus(status)
	updateAddr("")

	// goroutine waits asynch. until the instance is actually stopped to not block this call
	go NotifyWhenInstanceStopped(c)
	return nil
}

// Waits until instance is stopped.
// Updates instance status fields (set to stopped).
func NotifyWhenInstanceStopped(c chan struct{}) {
	describeInstancesInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{getRedisServerInstanceId()},
	}

	if err := Ec2Client.WaitUntilInstanceStopped(describeInstancesInput); err != nil {
		log.WithError(err).Error("waiting until Instance stopped failed")
	}
	if c != nil {
		c <- struct{}{}
	}
	updateStatus("stopped")

	log.Debug("EC2 instance stopped")

}

// Get the public IP address of the specified EC2 instance.
// Attention: This function is also used for the Gateway and Orchestrator EC2 instances.
//
// return addr and error
// 	- error nil if addr retrieved successfully
// 	- empty addr if error
func RetrieveIP(instanceId string) (publicIP string, privateIP string, err error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceId},
	}

	result, err := Ec2Client.DescribeInstances(input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			log.WithError(awsErr).Error("DescribeInstances for EC2 instance failed with AWS error")
		} else {
			log.WithError(err).Error("DescribeInstances for EC2 instance failed")
		}
		return "", "", ErrDescribeInstance
	}

	// check if instance listed
	if len(result.Reservations[0].Instances) == 0 {
		return "", "", ErrInstanceNotAvailable
	}
	// check if running
	if *result.Reservations[0].Instances[0].State.Code != 16 {
		return "", "", ErrRetrieveIp
	}

	return *(result.Reservations[0].Instances[0].PublicIpAddress), *(result.Reservations[0].Instances[0].PrivateIpAddress), nil
}

// Get the Status of the EC2 instance.
// return status and error:
// 	- error nil if status retrieved successfully
// 	- empty status if error
func RetrieveStatus() (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{getRedisServerInstanceId()},
	}

	result, err := Ec2Client.DescribeInstances(input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			log.WithError(awsErr).Error("DescribeInstances for EC2 instance failed with AWS error")
		} else {
			log.WithError(err).Error("DescribeInstances for EC2 instance failed")
		}
		return "", ErrDescribeInstance
	}

	// check if instance listed and running
	if len(result.Reservations[0].Instances) == 0 {
		return "", ErrInstanceNotAvailable
	}

	return *(result.Reservations[0].Instances[0].State.Name), nil

}
