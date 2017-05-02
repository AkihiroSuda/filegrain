package commands

import (
	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	mainCmdConfig struct {
		debug bool
	}

	MainCmd = &cobra.Command{
		Use:   "filegrain <command>",
		Short: "FILEgrain: transport-agnostic, fine-grained content-addressable container image layout",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if mainCmdConfig.debug {
				logrus.SetLevel(logrus.DebugLevel)
				logrus.Debug("running in debug mode")
			}
			return nil
		},
	}
)

func init() {
	MainCmd.PersistentFlags().BoolVar(&mainCmdConfig.debug, "debug", false, "debug")
	MainCmd.AddCommand(MountCmd)
	MainCmd.AddCommand(BuildCmd)
}
