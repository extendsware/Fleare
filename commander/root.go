package commander

import (
	"fmt"
	"os"

	_ "github.com/parashmaity/fleare/commander/list_cmd"
	_ "github.com/parashmaity/fleare/commander/map_cmd"
	_ "github.com/parashmaity/fleare/commander/num_cmd"
	_ "github.com/parashmaity/fleare/commander/string_cmd"
	"github.com/parashmaity/fleare/server"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "Fleare",
	Short: "an in-memory database",
	Run: func(cmd *cobra.Command, args []string) {
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
