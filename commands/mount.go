package commands

import (
	"errors"

	"github.com/AkihiroSuda/filegrain/lazyfs"
	"github.com/spf13/cobra"
)

var MountCmd = &cobra.Command{
	Use:   "mount <mountpoint> <manifest>",
	Short: "Mount with lazy fs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("must specify mountpoint and manifest")
		}
		opts := &lazyfs.ServerOptions{
			Mountpoint: args[0],
		}
		sv, err := lazyfs.NewServer(opts)
		if err != nil {
			return err
		}
		sv.Serve()
		return nil
	},
}
