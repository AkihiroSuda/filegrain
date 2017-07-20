package commands

import (
	"errors"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/AkihiroSuda/filegrain/lazyfs"
	"github.com/AkihiroSuda/filegrain/puller"
)

var (
	mountCmdConfig struct {
		debugFUSE bool
		refName   string
	}

	MountCmd = &cobra.Command{
		Use:   "mount <image> <mountpoint>",
		Short: "Mount with lazy fs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("must specify image and mountpoint")
			}
			img, mountpoint := args[0], args[1]
			cachePath, err := ioutil.TempDir("", "filegrain-blobcache")
			if err != nil {
				return err
			}
			logrus.Infof("Blob cache (ephemeral): %s", cachePath)
			defer os.RemoveAll(cachePath) // FIXME
			pvller, err := puller.NewBlobCacher(cachePath,
				puller.NewLocalPuller())
			if err != nil {
				return err
			}
			opts := lazyfs.Options{
				Mountpoint: mountpoint,
				Puller:     pvller,
				Image:      img,
				RefName:    mountCmdConfig.refName,
			}
			return serve(opts)
		},
	}
)

func init() {
	MountCmd.Flags().StringVar(&mountCmdConfig.refName, "tag", "latest", "tag (aka reference name)")
	MountCmd.Flags().BoolVar(&mountCmdConfig.debugFUSE, "debug-fuse", false, "debug FUSE")
}

func serve(opts lazyfs.Options) error {
	fs, err := lazyfs.NewFS(opts)
	if err != nil {
		return err
	}
	sv, err := lazyfs.NewServer(fs)
	if err != nil {
		return err
	}
	if mountCmdConfig.debugFUSE {
		fs.SetDebug(true)
		sv.SetDebug(true)
	}
	go sv.Serve()
	logrus.Infof("Mounting on %s", opts.Mountpoint)
	if err := sv.WaitMount(); err != nil {
		return err
	}
	logrus.Infof("Mounted on %s", opts.Mountpoint)
	defer func() {
		logrus.Infof("Unmounting %s", opts.Mountpoint)
		if err := sv.Unmount(); err != nil {
			panic(err) // FIXME
		}
		logrus.Infof("Unmounted %s", opts.Mountpoint)
	}()
	waitForSIGINT()
	return nil
}

func waitForSIGINT() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
