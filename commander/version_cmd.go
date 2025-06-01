package commander

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(VersionCmd)
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Current version for Fleare",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("%s\nVersion: %s\nBuild Date: %s\n", ProjectName, Version, BuildDate)

	},
}
