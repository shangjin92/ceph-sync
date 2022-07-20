package cmd

import (
	"fmt"
	"github.com/shangjin92/ceph-sync/internal/utils/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	DebugAble bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ceph-sync",
	Short: "ceph sync tool",
	Long:  "A simple tool to sync data to ceph.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initialization)

	rootCmd.PersistentFlags().BoolVar(&DebugAble, "debug", false, "logger ture for Debug, false for Info")
}

func initialization() {
	initLogger()
}

func initLogger() {
	// basic setting
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&logger.LogFormatter{
		WithColor: true,
	})

	// set debugAble
	logger.SetLogStdHook(DebugAble)
}
