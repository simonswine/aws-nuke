package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/simonswine/aws-nuke/pkg/aws_nuke"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "aws-nuke",
	Short: "Remove all resources from an AWS account",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().Bool(aws_nuke.FlagForceDestroy, false, "Enable this to actually destory resources")
}
