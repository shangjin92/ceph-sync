package store

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
	"time"
)

const DefaultS3Region string = "us-east-1"

type CephConfig struct {
	AccessKey string
	SecretKey string
	EndPoint  string
}

type CephClient struct {
	*s3.S3
	session *session.Session
}

func NewCephClient(cfg *CephConfig) (*CephClient, error) {
	cephClient := &CephClient{}

	var credential = credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, "")
	var awsConfig = aws.NewConfig().
		WithRegion(DefaultS3Region).
		WithEndpoint(cfg.EndPoint).
		WithDisableSSL(false).
		WithLogLevel(3).
		WithS3ForcePathStyle(true).
		WithCredentials(credential)

	cephClient.session = session.Must(session.NewSession())
	cephClient.S3 = s3.New(cephClient.session, awsConfig)

	return cephClient, nil
}

func (cephClient *CephClient) ListBuckets() (*ListBucketsResult, error) {
	result, err := cephClient.S3.ListBuckets(nil)
	if err != nil {
		logrus.Errorf("list buckets failed, ", err)
		return nil, err
	}

	var bucketNames []string
	for _, b := range result.Buckets {
		bucket := aws.StringValue(b.Name)
		logrus.Infof("list bucket, * %s created on %s", bucket, aws.TimeValue(b.CreationDate))

		bucketNames = append(bucketNames, bucket)
	}
	return &ListBucketsResult{BucketNames: bucketNames}, nil
}

func (cephClient *CephClient) CheckBucketExist(bucketName string) (bool, error) {
	headBucketInput := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := cephClient.S3.HeadBucket(headBucketInput)
	if err != nil {
		logrus.Error("check bucket failed, error: %v", err)
		return false, nil
	}
	logrus.Info("check bucket existence successful")
	return true, nil
}

func (cephClient *CephClient) CreateBucket(bucketName string) error {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := cephClient.S3.CreateBucket(params)
	if err != nil {
		logrus.Errorf("unable to create bucket: %s, %v", bucketName, err)
		return err
	}
	// Wait until bucket is created before finishing
	logrus.Infof("waiting for bucket %q to be created...", bucketName)
	err = cephClient.S3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		logrus.Errorf("error occurred while waiting for bucket: %s to be created, error: %v", bucketName, err)
		return err
	}
	logrus.Infof("bucket: %q successfully created...", bucketName)
	return nil
}

func (cephClient *CephClient) UploadFile(urlType UrlType, urlStr, dstBucketName, dstObjectName string) error {
	data, err := ReadUrlData(urlType, urlStr)
	if err != nil {
		logrus.Errorf("get object data failed, error: %v", err)
		return err
	}

	_, err = cephClient.S3.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(data),
		Bucket: &dstBucketName,
		Key:    &dstObjectName,
	})
	if err != nil {
		logrus.Errorf("upload object failed, bucket: %s, object name: %s", dstBucketName, dstObjectName)
		return err
	}
	logrus.Infof("upload object successful, bucket: %s, object name: %s", dstBucketName, dstObjectName)
	return nil
}

func (cephClient *CephClient) GetObjectUrl(bucketName, objectName string) (string, UrlType, error) {
	req, _ := cephClient.S3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})

	url, err := req.Presign(15 * time.Minute)
	return url, HttpUrl, err
}

func (cephClient *CephClient) ListObjects(bucketName, marker, prefix string) (*ListObjectsResult, error) {
	logrus.Infof("sync bucket: %s, list 1000 objects...", bucketName)
	listObjectsResponse, err := cephClient.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
		Marker: aws.String(marker),
		Prefix: &prefix,
	})
	if err != nil {
		logrus.Errorf("bucket: %s, list objects failed, error: %v", bucketName, err)
		return nil, err
	}

	var objectsName []string
	lastKey := ""
	for _, key := range listObjectsResponse.Contents {
		lastKey = *key.Key
		objectsName = append(objectsName, *key.Key)
	}

	suspendValue := true
	nonSuspendValue := false
	if !*listObjectsResponse.IsTruncated {
		logrus.Infof("suspend listing objects in bucket: %s", bucketName)
		return &ListObjectsResult{
			ObjectsName: objectsName,
			Suspend:     &suspendValue,
		}, nil
	} else {
		prevMarker := marker
		if listObjectsResponse.NextMarker == nil {
			// From the s3 docs: If response does not include the
			// NextMarker and it is truncated, you can use the value of the
			// last Key in the response as the marker in the subsequent
			// request to get the next set of object keys.
			marker = lastKey
		} else {
			marker = *listObjectsResponse.NextMarker
		}
		if marker == prevMarker {
			logrus.Error("Unable to list all bucket objects.")
			return &ListObjectsResult{
				ObjectsName: objectsName,
				Suspend:     &suspendValue,
			}, nil
		} else {
			return &ListObjectsResult{
				ObjectsName: objectsName,
				Suspend:     &nonSuspendValue,
				NextMarker:  &marker,
			}, nil
		}
	}
}
