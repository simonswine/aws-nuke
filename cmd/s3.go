package cmd

import (
	"github.com/spf13/cobra"

	"github.com/simonswine/aws-nuke/pkg/aws_nuke"
)

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Delete all S3 buckets and their contents",
	Run: func(cmd *cobra.Command, args []string) {
		a := aws_nuke.New()
		a.Must(a.CmdS3(cmd, args))
	},
}

func init() {
	RootCmd.AddCommand(s3Cmd)
}
