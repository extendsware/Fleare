package commander

import (
	"fmt"
	"os"

	"github.com/parashmaity/fleare/server"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "Fleare",
	Short: "an in-memory database",
	Run: func(cmd *cobra.Command, args []string) {
		// config.Load(cmd.Flags())
		// slog.SetDefault(logger.New())
		server.Start()
	},
}

func Execute() {
	// Execute the root command
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
