package core

import (
	"errors"
	"github.com/magiconair/properties"
	"github.com/shangjin92/ceph-sync/internal/store"
	"github.com/sirupsen/logrus"
	"github.com/wonderivan/logger"
	"strings"
	"sync"
)

const (
	SourceClusterAccessKey = "source_cluster_access_key"
	SourceClusterSecretKey = "source_cluster_secret_key"
	SourceClusterEndpoint  = "source_cluster_endpoint"
	TargetClusterAccessKey = "target_cluster_access_key"
	TargetClusterSecretKey = "target_cluster_secret_key"
	TargetClusterEndpoint  = "target_cluster_endpoint"
)

type SourceDataSourceConfig struct {
	dataSourceType   string
	clusterAccessKey string
	clusterSecretKey string
	clusterEndpoint  string
	clusterBucket    string
}

type TargetDataSourceConfig struct {
	clusterAccessKey string
	clusterSecretKey string
	clusterEndpoint  string
	clusterBucket    string
}

func loadSourceDataSourceConfig() *SourceDataSourceConfig {
	p := properties.MustLoadFile(SyncProperties, properties.UTF8)

	return &SourceDataSourceConfig{
		dataSourceType:   SourceType,
		clusterAccessKey: p.GetString(SourceClusterAccessKey, ""),
		clusterSecretKey: p.GetString(SourceClusterSecretKey, ""),
		clusterEndpoint:  p.GetString(SourceClusterEndpoint, ""),
	}
}

func loadTargetDataSourceConfig() *TargetDataSourceConfig {
	p := properties.MustLoadFile(SyncProperties, properties.UTF8)

	return &TargetDataSourceConfig{
		clusterSecretKey: p.MustGetString(TargetClusterSecretKey),
		clusterAccessKey: p.MustGetString(TargetClusterAccessKey),
		clusterEndpoint:  p.MustGetString(TargetClusterEndpoint),
	}
}

func newSourceStoreClient(config *SourceDataSourceConfig) (store.Store, error) {
	switch strings.ToLower(config.dataSourceType) {
	case "ceph":
		cephConfig := &store.CephConfig{
			AccessKey: config.clusterAccessKey,
			SecretKey: config.clusterSecretKey,
			EndPoint:  config.clusterEndpoint,
		}
		return store.NewCephClient(cephConfig)
	case "oss":
		ossConfig := &store.OssConfig{
			AccessID:  config.clusterAccessKey,
			AccessKey: config.clusterSecretKey,
			EndPoint:  config.clusterEndpoint,
		}
		return store.NewOssClient(ossConfig)
	case "local":
		return store.NewLocalClient()
	default:
		return nil, errors.New("don't support this client")
	}
}

func newTargetStoreClient(config *TargetDataSourceConfig) (store.Store, error) {
	cephConfig := &store.CephConfig{
		AccessKey: config.clusterAccessKey,
		SecretKey: config.clusterSecretKey,
		EndPoint:  config.clusterEndpoint,
	}

	return store.NewCephClient(cephConfig)
}

func SyncClusterBucketData() {
	logrus.Info("Begin sync data from source cluster bucket...")

	sourceCephClusterConfig := loadSourceDataSourceConfig()
	sourceStoreClient, err := newSourceStoreClient(sourceCephClusterConfig)
	if err != nil {
		logrus.Errorf("create source store client failed, error: %v", err)
		return
	}

	targetCephClusterConfig := loadTargetDataSourceConfig()
	targetStoreClient, err := newTargetStoreClient(targetCephClusterConfig)
	if err != nil {
		logrus.Errorf("create target store client failed, error: %v", err)
		return
	}

	syncBucketData(sourceStoreClient, targetStoreClient)

	logrus.Info("Finished sync data from source cluster bucket...")
}

func createBucketIfAbsent(bucketName string, targetClient store.Store) error {
	checkResult, _ := targetClient.CheckBucketExist(bucketName)
	if checkResult {
		return nil
	}
	err := targetClient.CreateBucket(bucketName)
	if err != nil {
		logrus.Errorf("create bucket failed, error: %v", err)
		return err
	}
	return nil
}

func syncBucketData(sourceClient, targetClient store.Store) {
	err := createBucketIfAbsent(TargetClusterBucket, targetClient)
	if err != nil {
		logrus.Errorf("Create bucket failed, bucket name: %s", TargetClusterBucket)
		return
	}

	marker := ""
	for {
		logrus.Infof("sync data to target cluster, bucket name: %s", TargetClusterBucket)

		var sourceBucket = SourceClusterBucket
		if SourceClusterBucket == "" && SourceLocalDirName != "" {
			sourceBucket = SourceLocalDirName
		}
		listObjectResult, err2 := sourceClient.ListObjects(sourceBucket, marker, SourceClusterObjectPrefix)
		if err2 != nil {
			logrus.Errorf("list objects failed, source type: %s, source cluster bucket: %s", SourceType, SourceClusterBucket)
			return
		}

		var wg sync.WaitGroup
		for _, key := range listObjectResult.ObjectsName {
			wg.Add(1)
			objectUrl, urlType, err3 := sourceClient.GetObjectUrl(sourceBucket, key)
			if err3 != nil {
				logrus.Errorf("get object url failed, object name: %s, error: %v", key, err3)
				return
			}
			var objectName = key
			if TargetClusterObjectPrefix != "" {
				objectName = TargetClusterObjectPrefix + objectName
			}
			go func(urlStr, dstObjectName string) {
				defer wg.Done()
				err = targetClient.UploadFile(urlType, urlStr, TargetClusterBucket, dstObjectName)
				if err != nil {
					logger.Error("upload object failed, bucket: %s, name: %s, error: %v", TargetClusterBucket, dstObjectName, err)
				}
			}(objectUrl, objectName)
		}
		wg.Wait()

		if *listObjectResult.Suspend {
			logrus.Info("sync process has finished.")
			return
		} else {
			marker = *listObjectResult.NextMarker
		}
	}
}
