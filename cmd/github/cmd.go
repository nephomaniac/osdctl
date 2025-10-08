package github

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "github",
	Short: "Provides a set of commands for interacting with GitHub",
	Args:  cobra.NoArgs,
}

func init() {
	Cmd.AddCommand(reviewPRCmd)
}
