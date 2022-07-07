package core

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/magiconair/properties"
	"github.com/wonderivan/logger"
	"sync"
	"time"
)

const (
	SourceClusterAccessKey = "source_cluster_access_key"
	SourceClusterSecretKey = "source_cluster_secret_key"
	SourceClusterEndpoint  = "source_cluster_endpoint"
	TargetClusterAccessKey = "target_cluster_access_key"
	TargetClusterSecretKey = "target_cluster_secret_key"
	TargetClusterEndpoint  = "target_cluster_endpoint"
	TargetClusterBucket    = "target_cluster_bucket"
)

func getTargetClusterBucket() string {
	p := properties.MustLoadFile(SyncProperties, properties.UTF8)
	return p.MustGetString(TargetClusterBucket)
}

func loadSourceClusterConfig() *CephConfig {
	p := properties.MustLoadFile(SyncProperties, properties.UTF8)
	accessKey := p.MustGetString(SourceClusterAccessKey)
	secretKey := p.MustGetString(SourceClusterSecretKey)
	endpoint := p.MustGetString(SourceClusterEndpoint)

	return &CephConfig{accessKey: accessKey,
		secretKey: secretKey,
		endPoint:  endpoint}
}

func loadTargetClusterConfig() *CephConfig {
	p := properties.MustLoadFile(SyncProperties, properties.UTF8)
	accessKey := p.MustGetString(TargetClusterAccessKey)
	secretKey := p.MustGetString(TargetClusterSecretKey)
	endpoint := p.MustGetString(TargetClusterEndpoint)

	return &CephConfig{accessKey: accessKey,
		secretKey: secretKey,
		endPoint:  endpoint}
}

func SyncClusterData() {
	logger.Info("Begin sync data from source cluster...")

	sourceCephClusterConfig := loadSourceClusterConfig()
	targetCephClusterConfig := loadTargetClusterConfig()

	var sourceCephClient = NewCephClient(sourceCephClusterConfig)
	var targetCephClient = NewCephClient(targetCephClusterConfig)

	result, err := sourceCephClient.client.ListBuckets(nil)
	if err != nil {
		logger.Error("list buckets failed, ", err)
		return
	}
	for _, b := range result.Buckets {
		logger.Info("----------------------------------------------------")
		bucket := aws.StringValue(b.Name)
		logger.Info("list bucket, * %s created on %s", bucket, aws.TimeValue(b.CreationDate))

		err = createBucketIfAbsent(bucket, targetCephClient)
		if err != nil {
			logger.Error("create bucket failed, error: %v", err)
			return
		}
		syncBucketData(bucket, sourceCephClient, targetCephClient)
		logger.Info("----------------------------------------------------")
	}
	logger.Info("Finished sync data from source cluster...")
}

func createBucketIfAbsent(bucketName string, targetCephClient *CephClient) error {
	checkResult := targetCephClient.CheckBucketExist(bucketName)
	if checkResult {
		return nil
	}
	err := targetCephClient.CreateBucket(bucketName)
	if err != nil {
		return err
	}
	return nil
}

func SyncClusterBucketData() {
	logger.Info("Begin sync data from source cluster bucket...")

	sourceCephClusterConfig := loadSourceClusterConfig()
	targetCephClusterConfig := loadTargetClusterConfig()

	var sourceCephClient = NewCephClient(sourceCephClusterConfig)
	var targetCephClient = NewCephClient(targetCephClusterConfig)

	syncBucketData(getTargetClusterBucket(), sourceCephClient, targetCephClient)

	logger.Info("Finished sync data from source cluster bucket...")
}

func syncBucketData(bucketName string, sourceCephClient, targetCephClient *CephClient) {
	marker := ""
	for {
		logger.Info("sync bucket: %s, list 1000 objects...", bucketName)
		listObjectsResponse, err := sourceCephClient.client.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(bucketName),
			Marker: aws.String(marker),
		})
		if err != nil {
			logger.Error("bucket: %s, list objects failed, error: %v", bucketName, err)
			return
		}

		lastKey := ""
		var wg sync.WaitGroup
		for _, key := range listObjectsResponse.Contents {
			//logger.Info("object Name: %s, object Last modified: %s, object size: %s, object storage class: %s",
			//	*key.Key, *key.LastModified, *key.Size, *key.StorageClass)
			lastKey = *key.Key

			//beforeTime := time.Now().AddDate(0, 0, -3)
			//if (*key.LastModified).Before(beforeTime) {
			//	continue
			//}

			req, _ := sourceCephClient.client.GetObjectRequest(&s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(*key.Key),
			})
			urlStr, _ := req.Presign(15 * time.Minute)
			//logger.Info("object name: %s, url: %s", *key.Key, urlStr)

			wg.Add(1)
			go func(urlStr, dstObjectName string) {
				defer wg.Done()
				err = targetCephClient.UploadFile(urlStr, bucketName, dstObjectName)
				if err != nil {
					logger.Error("upload object failed, bucket: %s, name: %s, error: %v", bucketName, dstObjectName, err)
				}
			}(urlStr, *key.Key)
		}
		wg.Wait()

		time.Sleep(3 * time.Second)
		logger.Info("sync bucket: %s, 1000 objects have been uploaded...", bucketName)

		if !*listObjectsResponse.IsTruncated {
			logger.Info("suspend listing objects in bucket: %s", bucketName)
			break
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
				logger.Error("Unable to list all bucket objects.")
				return
			}
		}
	}
}
