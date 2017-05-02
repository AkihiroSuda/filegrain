package commands

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/AkihiroSuda/filegrain/builder"
)

var (
	buildCmdConfig struct {
		refName string
		// TODO: add sourceType (rootfs, oci, ..)
	}

	BuildCmd = &cobra.Command{
		Use:   "build <source> <target>",
		Short: "Build a FILEgrain image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("must specify source and target")
			}
			source, target := args[0], args[1]
			b, err := builder.NewBuilderWithRootFS(source)
			if err != nil {
				return err
			}
			return b.Build(target, buildCmdConfig.refName)
		},
	}
)

func init() {
	BuildCmd.Flags().StringVar(&buildCmdConfig.refName, "tag", "latest", "tag (aka reference name)")
}
