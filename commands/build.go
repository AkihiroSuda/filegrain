package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/AkihiroSuda/filegrain/builder"
)

var (
	buildCmdConfig struct {
		refName    string
		sourceType string
		target     string
	}

	BuildCmd = &cobra.Command{
		Use:   "build -o  <target> <source>",
		Short: "Build a FILEgrain image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if buildCmdConfig.target == "" {
				return errors.New("must specify target output (-o)")
			}
			if len(args) != 1 {
				return errors.New("must specify source")
			}
			source := args[0]
			b, err := newBuilder(buildCmdConfig.sourceType, source)
			if err != nil {
				return err
			}
			return b.Build(buildCmdConfig.target, buildCmdConfig.refName)
		},
	}
)

func init() {
	BuildCmd.Flags().StringVarP(&buildCmdConfig.target, "output", "o", "", "target output path")
	BuildCmd.Flags().StringVar(&buildCmdConfig.refName, "tag", "latest", "tag (aka reference name)")
	BuildCmd.Flags().StringVar(&buildCmdConfig.sourceType, "source-type", "auto", "source type (auto, oci-image, docker-image, rootfs)")
}

func newBuilder(sourceType, source string) (builder.Builder, error) {
	if sourceType == "auto" || sourceType == "" {
		sourceType = guessSourceType(source)
		if sourceType != "" {
			logrus.Infof("detected source type %q for %s", sourceType, source)
		} else {
			return nil, fmt.Errorf("could not detect source type for %s", source)
		}
	}
	switch sourceType {
	case "oci-image":
		return builder.NewBuilderWithOCIImage(source)
	case "docker-image":
		return builder.NewBuilderWithDockerImage(source)
	case "rootfs":
		return builder.NewBuilderWithRootFS(source)
	}
	return nil, fmt.Errorf("unknown source type: %s", sourceType)
}

func guessSourceType(source string) string {
	fi, err := os.Stat(source)
	if err == nil && fi.IsDir() {
		_, err := os.Stat(filepath.Join(source, "oci-layout"))
		if err == nil {
			return "oci-image"
		}
		return "rootfs"
	}
	// FIXME: not accurate
	return "docker-image"
}
