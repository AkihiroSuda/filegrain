package cp

import (
	"os/exec"

	"github.com/Sirupsen/logrus"
)

func CopyFile(dst, src string) error {
	// terrible implementation, but it does not hurt for POC...
	out, err := exec.Command("cp", "-a", src, dst).CombinedOutput()
	if len(out) != 0 {
		logrus.Warnf("output from `cp -a %s %s`: %s", src, dst, out)
	}
	return err
}

func CopyDirectory(dst, src string) error {
	// terrible implementation, but it does not hurt for POC...
	out, err := exec.Command("rm", "-rf", dst).CombinedOutput()
	if len(out) != 0 {
		logrus.Warnf("output from `rm -rf %s`: %s", dst, out)
	}
	if err != nil {
		return err
	}
	out, err = exec.Command("cp", "-a", src, dst).CombinedOutput()
	if len(out) != 0 {
		logrus.Warnf("output from `cp -a %s %s`: %s", src, dst, out)
	}
	return err
}
