package lazyfs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/AkihiroSuda/filegrain/lazyfs/dummycontent"
	"github.com/AkihiroSuda/filegrain/puller"
)

// FS is a read-only filesystem with lazy-pull feature.
//
// FS implements github.com/hanwen/go-fuse/fuse/pathfs.FileSystem
//
// Supported objects:
//  - regular files
//  - symbolic links
type FS struct {
	pathfs.FileSystem
	opts      Options
	underlier string
}

func (fs *FS) openDummyContent(name string) (*dummycontent.DummyRegularFileContent, error) {
	b, err := ioutil.ReadFile(filepath.Join(fs.underlier, name))
	if err != nil {
		return nil, err
	}
	var c dummycontent.DummyRegularFileContent
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if len(c.Digests) < 1 {
		return nil, fmt.Errorf("expected len >= 1, got %d", len(c.Digests))
	}
	return &c, nil
}

func (fs *FS) GetAttr(name string, fc *fuse.Context) (*fuse.Attr, fuse.Status) {
	attr, ok := fs.FileSystem.GetAttr(name, fc)
	if ok != fuse.OK {
		return attr, ok
	}
	c, _ := fs.openDummyContent(name)
	if c != nil {
		attr.Size = uint64(c.Size)
	}
	return attr, fuse.OK
}

func (fs *FS) Open(name string, flags uint32, fc *fuse.Context) (nodefs.File, fuse.Status) {
	f, ok := fs.FileSystem.Open(name, flags, fc)
	if ok != fuse.OK {
		return f, ok
	}
	c, _ := fs.openDummyContent(name)
	if c == nil {
		return f, ok
	}
	reader, err := fs.opts.Puller.PullBlob(fs.opts.Image, c.Digests[0])
	if err != nil {
		return nil, fuse.EIO
	}
	newf, err := newFile(name, reader, fs)
	if err != nil {
		return nil, fuse.EIO
	}
	return newf, fuse.OK
}

type file struct {
	nodefs.File
	name string
	fs   *FS
}

func (f *file) GetAttr(out *fuse.Attr) fuse.Status {
	attr, ok := f.fs.GetAttr(f.name, nil)
	if attr != nil {
		*out = *attr
	}
	return ok
}

func newFile(name string, reader io.Reader, fs *FS) (*file, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	f := nodefs.NewReadOnlyFile(nodefs.NewDataFile(b))
	return &file{
		File: f,
		name: name,
		fs:   fs,
	}, nil
}

type Options struct {
	Mountpoint string
	Puller     puller.Puller
	Image      string
	RefName    string
}

func NewFS(opts Options) (*FS, error) {
	underlier, err := ioutil.TempDir("", "filegrain-underlier")
	if err != nil {
		return nil, err
	}
	logrus.Infof("underlier: %s", underlier)
	if err := loadUnderlier(opts, underlier); err != nil {
		return nil, err
	}
	fs := &FS{
		FileSystem: pathfs.NewReadonlyFileSystem(pathfs.NewLoopbackFileSystem(underlier)),
		opts:       opts,
		underlier:  underlier,
	}
	return fs, nil
}

func CleanupWithFS(fs *FS) {
	logrus.Infof("removing %s", fs.underlier)
	os.RemoveAll(fs.underlier)
}

func NewServer(fs *FS) (*fuse.Server, error) {
	nfs := pathfs.NewPathNodeFs(pathfs.NewReadonlyFileSystem(fs), nil)
	conn := nodefs.NewFileSystemConnector(nfs.Root(), nil)
	return fuse.NewServer(conn.RawFS(), fs.opts.Mountpoint,
		&fuse.MountOptions{
			FsName: fs.opts.Mountpoint + ":" + fs.opts.RefName,
			Name:   "filegrain.lazyfs",
		})
}
