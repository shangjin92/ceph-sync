package cmd

import (
	"github.com/shangjin92/ceph-sync/core"
	"github.com/spf13/cobra"
)

var syncBucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "sync ceph bucket data",
	Long:  `ceph-sync bucket --config /root/sync.properties`,
	Run: func(cmd *cobra.Command, args []string) {
		core.SyncClusterBucketData()
	},
}

func init() {
	rootCmd.AddCommand(syncBucketCmd)

	syncBucketCmd.Flags().StringVar(&core.SyncProperties, "config", "/root/sync.properties",
		"ceph bucket sync config")
}
