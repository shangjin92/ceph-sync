package store

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type LocalClient struct {
}

func NewLocalClient() (*LocalClient, error) {
	return &LocalClient{}, nil
}

func (localClient *LocalClient) ListBuckets() (*ListBucketsResult, error) {
	return nil, nil
}

func (localClient *LocalClient) CheckBucketExist(dirName string) (bool, error) {
	return false, nil
}

func (localClient *LocalClient) CreateBucket(dirName string) error {
	return nil
}

func (localClient *LocalClient) UploadFile(urlType UrlType, urlStr, dstBucketName, dstObjectName string) error {
	return nil
}

func (localClient *LocalClient) GetObjectUrl(dirName, objectName string) (string, UrlType, error) {
	if filepath.IsAbs(objectName) {
		return objectName, LocalUrl, nil
	} else {
		objectName, _ = filepath.Abs(objectName)
		return objectName, LocalUrl, nil
	}
}

func (localClient *LocalClient) ListObjects(dirName, marker, prefix string) (*ListObjectsResult, error) {
	var objectsName []string
	err := filepath.Walk(dirName,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				objectsName = append(objectsName, path)
			}
			return nil
		})
	if err != nil {
		logrus.Errorf("recursive list files failed, dirname: %s, error: %s", dirName, err)
		return nil, err
	}
	suspendValue := true
	nextMarker := ""

	return &ListObjectsResult{
		ObjectsName: objectsName,
		Suspend:     &suspendValue,
		NextMarker:  &nextMarker,
	}, nil
}
