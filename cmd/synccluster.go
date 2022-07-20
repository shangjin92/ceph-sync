package cmd

import (
	"fmt"
	"github.com/shangjin92/ceph-sync/core"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "cluster",
	Short: "sync ceph cluster data",
	Long:  `ceph-sync cluster --config /root/sync.properties`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("To be realized")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVar(&core.SyncProperties, "config", "/root/sync.properties", "ceph cluster sync config")
}
