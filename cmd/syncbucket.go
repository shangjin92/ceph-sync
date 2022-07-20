package cmd

import (
	"github.com/shangjin92/ceph-sync/core"
	"github.com/spf13/cobra"
)

var syncBucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "sync ceph bucket data",
	Long: `ceph-sync bucket --config /root/sync.properties --source-type local --source-dir-path data-dir \
      --target-bucket bucket-name \
      --target-object-prefix file-prefix`,
	Run: func(cmd *cobra.Command, args []string) {
		core.SyncClusterBucketData()
	},
}

func init() {
	rootCmd.AddCommand(syncBucketCmd)

	syncBucketCmd.Flags().StringVar(&core.SyncProperties, "config", "/root/sync.properties", "ceph bucket sync config")
	syncBucketCmd.Flags().StringVar(&core.SourceType, "source-type", "", "source type, maybe: oss/ceph/local")
	syncBucketCmd.Flags().StringVar(&core.SourceLocalDirName, "source-dir-path", "", "local directory to be uploaded")
	syncBucketCmd.Flags().StringVar(&core.SourceClusterBucket, "source-bucket", "", "bucket name of source cluster")
	syncBucketCmd.Flags().StringVar(&core.SourceClusterObjectPrefix, "source-object-prefix", "", "object's prefix in source bucket")
	syncBucketCmd.Flags().StringVar(&core.TargetClusterBucket, "target-bucket", "", "bucket name of target cluster")
	syncBucketCmd.Flags().StringVar(&core.TargetClusterObjectPrefix, "target-object-prefix", "", "object's prefix in target bucket")
}
