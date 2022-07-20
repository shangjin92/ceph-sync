package store

import (
	"bytes"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"github.com/wonderivan/logger"
)

type OssConfig struct {
	AccessID  string
	AccessKey string
	EndPoint  string
}

type OssClient struct {
	*oss.Client
}

func NewOssClient(cfg *OssConfig) (*OssClient, error) {
	client, err := oss.New(cfg.EndPoint, cfg.AccessID, cfg.AccessKey)
	if err != nil {
		return nil, err
	}
	return &OssClient{Client: client}, nil
}

func (ossClient *OssClient) ListBuckets() (*ListBucketsResult, error) {
	lbr, err := ossClient.Client.ListBuckets()
	if err != nil {
		return nil, err
	}

	var bucketNames []string
	for _, bucket := range lbr.Buckets {
		bucketNames = append(bucketNames, bucket.Name)
	}
	return &ListBucketsResult{
		BucketNames: bucketNames,
	}, nil
}

func (ossClient *OssClient) CheckBucketExist(bucketName string) (bool, error) {
	bucket, err := ossClient.Client.Bucket(bucketName)
	if err != nil {
		return false, err
	}
	return bucket != nil, nil
}

func (ossClient *OssClient) CreateBucket(bucketName string) error {
	err := ossClient.Client.CreateBucket(bucketName)

	return err
}

func (ossClient *OssClient) UploadFile(urlType UrlType, urlStr, dstBucketName, dstObjectName string) error {
	data, err := ReadUrlData(urlType, urlStr)
	if err != nil {
		logrus.Errorf("get object data failed, error: %v", err)
		return err
	}

	bucket, err := ossClient.Client.Bucket(dstBucketName)
	if err != nil {
		return err
	}

	return bucket.PutObject(dstObjectName, bytes.NewReader(data))
}

func (ossClient *OssClient) GetObjectUrl(bucketName, objectName string) (string, UrlType, error) {
	bucket, err := ossClient.Client.Bucket(bucketName)
	if err != nil {
		return "", HttpUrl, err
	}

	url, err := bucket.SignURL(objectName, oss.HTTPGet, 15*60)
	return url, HttpUrl, err
}

func (ossClient *OssClient) ListObjects(bucketName, marker, prefix string) (*ListObjectsResult, error) {
	bucket, err := ossClient.Client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	lor, err := bucket.ListObjects(oss.Marker(marker), oss.Prefix(prefix))
	if err != nil {
		logrus.Errorf("bucket: %s, list objects failed, error: %v", bucketName, err)
		return nil, err
	}

	var objectsName []string
	lastKey := ""
	for _, object := range lor.Objects {
		lastKey = object.Key
		objectsName = append(objectsName, object.Key)
	}

	suspendValue := true
	nonSuspendValue := false
	if !lor.IsTruncated {
		logrus.Infof("suspend listing objects in bucket: %s", bucketName)
		return &ListObjectsResult{
			ObjectsName: objectsName,
			Suspend:     &suspendValue,
		}, nil
	} else {
		prevMarker := marker
		if lor.NextMarker == "" {
			marker = lastKey
		} else {
			marker = lor.NextMarker
		}
		if marker == prevMarker {
			logger.Error("Unable to list all bucket objects.")
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
