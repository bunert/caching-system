package awsutils

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	log           = logrus.WithField("component", "awsutils")
	AWSsession    *session.Session
	LambdaSession *lambda.Lambda
	Ec2Client     *ec2.EC2

	AccessKeyID     string
	SecretAccessKey string
	Region          string = "eu-central-1"
	Bucket          string = "bunert-testbucket"
)

//GetEnvWithKey : get env value
func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}

// load env variabled from .env file
// TODO: not working for lambda (.env file not available, therefore hardcoded) -> better way to do it?
func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.WithError(err).Error("error loading .env file: ")
		os.Exit(1)
	}
	Bucket = GetEnvWithKey("BUCKET_NAME")
	AccessKeyID = GetEnvWithKey("AWS_ACCESS_KEY_ID")
	SecretAccessKey = GetEnvWithKey("AWS_SECRET_ACCESS_KEY")
	Region = GetEnvWithKey("AWS_REGION")
}

// aws session setup for lambda-runtime
func CreateLambdaSession() {
	if AWSsession == nil {
		AWSsession = session.Must(session.NewSession(&aws.Config{Region: aws.String(Region)}))

		log.Debug("AWS session established")
	} else {
		log.Debug("AWS session already established")
	}
	// setup downloader for S3
	Downloader = s3manager.NewDownloader(AWSsession)

}

// aws session setup for Gateway
func SetupGatewaySession() {
	AWSsession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// setup downloader for S3
	Downloader = s3manager.NewDownloader(AWSsession)

	Ec2Client = ec2.New(AWSsession, &aws.Config{
		Region: aws.String(Region),
	})

}

// aws session setup for Orchestrator
func SetupOrchestratorSession() {
	AWSsession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// setup downloader for S3
	Downloader = s3manager.NewDownloader(AWSsession)

	LambdaSession = lambda.New(AWSsession, &aws.Config{
		Region: aws.String(Region),
	})

	Ec2Client = ec2.New(AWSsession, &aws.Config{
		Region: aws.String(Region),
	})

}

// aws session setup for S3 proxy server
func SetupS3ProxySession() {
	AWSsession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// setup downloader for S3
	Downloader = s3manager.NewDownloader(AWSsession)
}

// aws session setup for ElastiCache proxy server
func SetupECProxySession() {
	AWSsession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// setup downloader for S3
	Downloader = s3manager.NewDownloader(AWSsession)

	// setup ElastiCache session
	ElastiCache = elasticache.New(AWSsession)

}
