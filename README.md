# ceph-sync
A simple tool to sync ceph data from different cluster or different bucket.

```bash
./ceph-cli sync --config sync.properties
```

Sample sync.properties
```
source_cluster_access_key = ${AccessKey}
source_cluster_secret_key = ${SecretKey}
source_cluster_endpoint = http://192.0.0.1:7480

target_cluster_access_key = ${AccessKey}
target_cluster_secret_key = ${SecretKey}
target_cluster_endpoint = http://192.0.1.100:7480
target_cluster_bucket = test-bucket
```