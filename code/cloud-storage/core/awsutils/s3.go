package awsutils

import (
	"bytes"
	t "cloud-storage/lambda/gatewayComm/types"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
)

var (
	Downloader *s3manager.Downloader

	ErrKeyNotFound = errors.New("key for S3 does not exists")
)

// general downloader helper function for S3
func DownloadObject(key string) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(key),
	}

	object, err := Downloader.S3.GetObject(input)
	if err != nil {
		// 404 if key not available,  not really an Error for the Gateway
		if awsErr, ok := err.(awserr.Error); ok && (awsErr.Code() == "NoSuchKey") {
			log.Warnf("Specified Key (%s) for S3 does not exist", key)
			return nil, ErrKeyNotFound
		}
		// log error, should not happen
		log.WithError(err).Errorf("fetching key %s from bucket %s failed", key, Bucket)
		return nil, err
	}

	// If there is no content length, it is a directory
	if object.ContentLength == nil {
		log.Panic("directory?")
	}

	return object, nil
}

// used by gateway
// retrieves the object specified by the key from the S3 bucket and makes gin Context reply ready
func AccessStorage(c *gin.Context, key string) {

	object, err := DownloadObject(key)

	if err == ErrKeyNotFound {
		log.Info("ErrKeyNotFound case")
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.WithError(err).Error("AccessStorage retrieved unhandled error")
		c.Status(http.StatusInternalServerError)
		return
	}

	if object.Body != nil {
		defer object.Body.Close()
	}

	c.Header("Content-Type", *object.ContentType)
	c.Header("Content-Origin", "S3")
	c.Header("Content-Length", fmt.Sprintf("%d", *object.ContentLength))
	c.Status(http.StatusOK)
	io.Copy(c.Writer, object.Body)

}

// used by Orchestrator
func GetBytes(key string) *bytes.Buffer {

	object, err := DownloadObject(key)

	if err != nil {
		log.WithError(err).Error("GetBytes failed")
		return nil
	}
	if object.Body != nil {
		defer object.Body.Close()
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(object.Body)

	return buf
}

// used by Lambda-Runtime
func GetObject(key string) *t.Body {

	object, err := DownloadObject(key)

	if err != nil {
		log.WithError(err).Error("GetBytes failed")
		return nil
	}
	if object.Body != nil {
		defer object.Body.Close()
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(object.Body)

	// log.Info("downloaded object ", key, " successfully")

	return &t.Body{
		ContentType:   *object.ContentType,
		ContentLength: fmt.Sprintf("%d", *object.ContentLength),
		Msg:           buf.String(),
	}

}

// used for ElastiCache testing
// retrieves the object specified by the key from the S3 bucket and makes gin Context reply ready
func GetAndReturnObject(c *gin.Context, key string) *bytes.Buffer {

	object, err := DownloadObject(key)

	if err == ErrKeyNotFound {
		log.Info("ErrKeyNotFound case")
		c.Status(http.StatusNotFound)
		return nil
	} else if err != nil {
		log.WithError(err).Error("AccessStorage retrieved unhandled error")
		c.Status(http.StatusInternalServerError)
		return nil
	}

	if object.Body != nil {
		defer object.Body.Close()
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(object.Body)

	c.Header("Content-Type", *object.ContentType)
	c.Header("Content-Origin", "S3")
	c.Header("Content-Length", fmt.Sprintf("%d", *object.ContentLength))
	c.Status(http.StatusOK)
	c.Writer.WriteString(buf.String())

	return buf

}

// uploads object from GIN API request to the S3 bucket
func UploadStorage(c *gin.Context) {
	sess := c.MustGet("sess").(*session.Session)
	uploader := s3manager.NewUploader(sess)

	file, err := c.FormFile("file")
	if err != nil {
		log.WithError(err).Errorf("error retrieving object %s from HTTP request: ", file.Filename)
		c.JSON(http.StatusBadRequest, gin.H{
			"filename": file.Filename,
		})
		return
	}
	f, err := file.Open()

	if err != nil {
		log.WithError(err).Error("opening file failed: ")
	}
	defer f.Close()

	//upload to the s3 bucket
	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(file.Filename),
		Body:   f,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to upload file",
			"uploader": up,
		})
		log.WithError(err).Errorf("unable to upload %s to bucket %s: ", file.Filename, Bucket)
		return
	}

	filepath := "https://" + Bucket + "." + "s3-" + Region + ".amazonaws.com/" + file.Filename
	c.JSON(http.StatusOK, gin.H{
		"filepath": filepath,
	})
}
