package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	apkg "github.com/diamnet/go/support/app"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print aurora version",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(apkg.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
