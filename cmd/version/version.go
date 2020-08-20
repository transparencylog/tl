package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.transparencylog.com/tl/config"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and build information",

	Args: cobra.ExactArgs(0),

	Run: version,
}

func version(cmd *cobra.Command, args []string) {
	fmt.Printf("version: %s\n", config.Version)
	fmt.Printf("commit: %s\n", config.Commit)
	fmt.Printf("date: %s\n", config.Date)
}
