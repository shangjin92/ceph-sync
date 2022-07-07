package core

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/wonderivan/logger"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const DefaultS3Region string = "us-east-1"

type CephConfig struct {
	accessKey string
	secretKey string
	endPoint  string
}

type CephClient struct {
	client  *s3.S3
	session *session.Session
}

func NewCephClient(cfg *CephConfig) *CephClient {
	cephClient := &CephClient{}

	var credential = credentials.NewStaticCredentials(cfg.accessKey, cfg.secretKey, "")
	var awsConfig = aws.NewConfig().
		WithRegion(DefaultS3Region).
		WithEndpoint(cfg.endPoint).
		WithDisableSSL(false).
		WithLogLevel(3).
		WithS3ForcePathStyle(true).
		WithCredentials(credential)

	cephClient.session = session.Must(session.NewSession())
	cephClient.client = s3.New(cephClient.session, awsConfig)

	return cephClient
}

func (cephClient *CephClient) CheckBucketExist(bucketName string) bool {
	headBucketInput := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := cephClient.client.HeadBucket(headBucketInput)
	if err != nil {
		logger.Error("check bucket failed, error: %v", err)
		return false
	}
	logger.Info("check bucket existence successful")
	return true
}

func (cephClient *CephClient) CreateBucket(bucketName string) error {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := cephClient.client.CreateBucket(params)
	if err != nil {
		logger.Error("unable to create bucket: %s, %v", bucketName, err)
		return err
	}
	// Wait until bucket is created before finishing
	logger.Info("waiting for bucket %q to be created...", bucketName)
	err = cephClient.client.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		logger.Error("error occurred while waiting for bucket: %s to be created, error: %v", bucketName, err)
		return err
	}
	logger.Info("bucket: %q successfully created...", bucketName)
	return nil
}

func (cephClient *CephClient) UploadFile(urlStr, dstBucketName, dstObjectName string) error {
	resp, err := http.Get(strings.TrimSpace(urlStr))
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("close http body error: %v", err)
		}
	}(resp.Body)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	_, err = cephClient.client.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(data),
		Bucket: &dstBucketName,
		Key:    &dstObjectName,
	})
	if err != nil {
		return err
	}
	//logger.Info("upload object successful, bucket: %s, name: %s, result: %s", dstBucketName, dstObjectName, result)
	return nil
}
