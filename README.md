# ceph-sync
A simple tool to sync data from multi datasource to ceph cluster.

## Usage
To avoid AK information leakage, use local config to store AK information for source and target clusters.
```
source_cluster_access_key = ${AccessKey}
source_cluster_secret_key = ${SecretKey}
source_cluster_endpoint = http://192.0.0.1:7480

target_cluster_access_key = ${AccessKey}
target_cluster_secret_key = ${SecretKey}
target_cluster_endpoint = http://192.0.1.100:7480
```

### From Local File
* Write the AK information and Endpoint of the target Ceph cluster.
* Prepare the local file to be uploaded.
* Run the following command to synchronize data.

```bash
# source-dir-path: The file directory to be synchronized, all files in the current directory will be recursively uploaded.
# target-bucket:  The Bucket of the target Ceph cluster to be synchronized.
# target-object-prefix: This parameter is optional. The final file name is the relative path of the file in the directory on which the prefix is merged.
./ceph-sync bucket --config sync.properties --source-type local \
      --source-dir-path data-dir \
      --target-bucket bucket-name \
      --target-object-prefix file-prefix/
```

### From Aliyun OSS
* Write the AK information and Endpoint of the source OSS cluster and target Ceph cluster.
* Run the following command to synchronize data.

```bash
# source-bucket： The Bucket of source OSS.
# source-object-prefix: The prefix of the source file, it will filters out files that do not contain the prefix.
# target-bucket:  The Bucket of the target Ceph cluster to be synchronized. The final file name will be the same as the OSS file name.
./ceph-sync bucket --config sync.properties --source-type oss \
      --source-bucket bucket-name \
      --source-object-prefix file-prefix \
      --target-bucket bucket-name
```

### From Other Ceph Cluster
* Write the AK information and Endpoint of the source Ceph cluster and target Ceph cluster.
* Run the following command to synchronize data.

```bash
# source-bucket： The Bucket of source Ceph.
# source-object-prefix: The prefix of the source file, it will filters out files that do not contain the prefix.
# target-bucket: The Bucket of the target Ceph cluster to be synchronized. The final file name will be the same as the Ceph file name.
./ceph-sync bucket --config sync.properties --source-type ceph \
      --source-bucket bucket-name \
      --source-object-prefix file-prefix \
      --target-bucket bucket-name
```