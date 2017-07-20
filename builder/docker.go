package builder

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type fromDockerImageBuilder struct {
	source string
}

func NewBuilderWithDockerImage(source string) (Builder, error) {
	return &fromDockerImageBuilder{
		source: source,
	}, nil
}

// Build builds the FILEgrain image from b.source.
// Current implementation converts the source to raw rootfs,
// and internally uses fromRootFSBuilder.
// TODO: use fromOCIBuilder when `docker save` supports OCI.
func (b *fromDockerImageBuilder) Build(img, refName string) error {
	tmpDir, err := ioutil.TempDir("", "filegrain-fromDockerImageBuilder")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	rootfs := filepath.Join(tmpDir, "rootfs")
	if err := os.Mkdir(rootfs, 0755); err != nil {
		return err
	}
	logrus.Infof("unpacking docker image %s", b.source)
	if err = convertDockerImageToRootFS(rootfs, b.source); err != nil {
		return err
	}
	rb, err := NewBuilderWithRootFS(rootfs)
	if err != nil {
		return err
	}
	return rb.Build(img, refName)
}

// convertDockerImageToRootFS converts a Docker image to a raw rootfs.
// current implementation uses os/exec rather than client pkg,
// so as to reduce go dependencies.
func convertDockerImageToRootFS(dir, source string) error {
	var createOut bytes.Buffer
	create := exec.Command("docker", "container", "create", source)
	create.Env = os.Environ()
	create.Stdout = &createOut
	create.Stderr = os.Stderr // prints progress of `docker pull`
	if err := create.Run(); err != nil {
		return errors.Wrapf(err, "running cmd=%s %v, stdout=%q", create.Path, create.Args, createOut.String())
	}
	tmpContainerID := strings.TrimSpace(createOut.String())
	defer exec.Command("docker", "container", "rm", "-f", tmpContainerID).Run()

	export := exec.Command("docker", "container", "export", tmpContainerID)
	export.Env = os.Environ()
	tarCxf := exec.Command("tar", "Cxf", dir, "-")
	tarCxf.Env = os.Environ()
	pr, pw := io.Pipe()
	export.Stdout = pw
	export.Stderr = os.Stderr
	tarCxf.Stdin = pr
	tarCxf.Stdout = os.Stdout
	tarCxf.Stderr = os.Stderr
	if err := export.Start(); err != nil {
		return errors.Wrapf(err, "running cmd=%s %v", export.Path, export.Args)
	}
	if err := tarCxf.Start(); err != nil {
		return errors.Wrapf(err, "running cmd=%s %v", tarCxf.Path, tarCxf.Args)
	}
	if err := export.Wait(); err != nil {
		return errors.Wrapf(err, "waiting for cmd=%s %v", export.Path, export.Args)
	}
	if err := pw.Close(); err != nil {
		return err
	}
	if err := tarCxf.Wait(); err != nil {
		return errors.Wrapf(err, "waiting for cmd=%s %v", tarCxf.Path, tarCxf.Args)
	}
	return nil
}
