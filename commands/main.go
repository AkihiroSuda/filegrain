package commands

import (
	"github.com/spf13/cobra"
)

var (
	MainCmd = &cobra.Command{
		Use:   "filegrain <command>",
		Short: "transport-agnostic, fine-grained content-addressable container image layout",
	}
)

func init() {
	MainCmd.AddCommand(MountCmd)
}
