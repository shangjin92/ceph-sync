package store

import (
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type ListBucketsResult struct {
	BucketNames []string
}

type ListObjectsResult struct {
	ObjectsName []string
	Suspend     *bool
	NextMarker  *string
}

type UrlType string

const (
	HttpUrl  UrlType = "http"
	LocalUrl UrlType = "file"
)

type Store interface {
	ListBuckets() (*ListBucketsResult, error)
	CheckBucketExist(bucketName string) (bool, error)
	CreateBucket(bucketName string) error
	UploadFile(urlType UrlType, urlStr, dstBucketName, dstObjectName string) error
	GetObjectUrl(bucketName, objectName string) (string, UrlType, error)
	ListObjects(bucketName, marker, prefix string) (*ListObjectsResult, error)
}

func ReadUrlData(urlType UrlType, urlStr string) ([]byte, error) {
	if urlType == HttpUrl {
		return readHttpUrl(urlStr)
	} else {
		return readLocalUrl(urlStr)
	}
}

func readHttpUrl(urlStr string) ([]byte, error) {
	resp, err := http.Get(strings.TrimSpace(urlStr))
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logrus.Errorf("close http body error: %v", err)
		}
	}(resp.Body)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func readLocalUrl(urlStr string) ([]byte, error) {
	return ioutil.ReadFile(urlStr)
}
