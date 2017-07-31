package commands

import (
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/spf13/cobra"

	"github.com/AkihiroSuda/filegrain/lazyfs"
	"github.com/AkihiroSuda/filegrain/puller"
)

var (
	mountCmdConfig struct {
		fuseDebug        bool
		refName          string
		fuseCacheTimeout time.Duration
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
			nodefsOpts := nodefs.NewOptions()
			nodefsOpts.EntryTimeout = mountCmdConfig.fuseCacheTimeout
			nodefsOpts.AttrTimeout = mountCmdConfig.fuseCacheTimeout
			nodefsOpts.NegativeTimeout = mountCmdConfig.fuseCacheTimeout
			return serve(opts, nodefsOpts)
		},
	}
)

func init() {
	MountCmd.Flags().StringVar(&mountCmdConfig.refName, "tag", "latest", "tag (aka reference name)")
	MountCmd.Flags().BoolVar(&mountCmdConfig.fuseDebug, "fuse-debug", false, "debug FUSE")
	MountCmd.Flags().DurationVar(&mountCmdConfig.fuseCacheTimeout, "fuse-cache-timeout", 3*time.Minute, "To be documented")
}

func serve(opts lazyfs.Options, nodefsOpts *nodefs.Options) error {
	fs, err := lazyfs.NewFS(opts)
	if err != nil {
		return err
	}
	sv, err := lazyfs.NewServer(fs, nodefsOpts)
	if err != nil {
		return err
	}
	if mountCmdConfig.fuseDebug {
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
