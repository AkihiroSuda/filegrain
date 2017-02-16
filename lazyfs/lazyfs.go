package lazyfs

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// FS is a read-only filesystem with lazy-pull feature.
// Supported digests:
//  - SHA256
//
// Supported objects:
//  - regular files
//  - symbolic links
type FS struct {
	pathfs.FileSystem
	manifest interface{}
	puller   interface{}
}

func (fs *FS) GetAttr(name string, fc *fuse.Context) (*fuse.Attr, fuse.Status) {
	return nil, fuse.ENOENT
}

func (fs *FS) OpenDir(name string, fc *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	return nil, fuse.ENOENT
}

func (fs *FS) Open(name string, flagsd uint32, fc *fuse.Context) (nodefs.File, fuse.Status) {
	return nil, fuse.ENOENT
}

type ServerOptions struct {
	Mountpoint string
}

func NewServer(opts *ServerOptions) (*fuse.Server, error) {
	fs := &FS{FileSystem: pathfs.NewDefaultFileSystem()}
	nfs := pathfs.NewPathNodeFs(fs, nil)
	conn := nodefs.NewFileSystemConnector(nfs.Root(), nil)
	return fuse.NewServer(conn.RawFS(), opts.Mountpoint,
		&fuse.MountOptions{
			AllowOther: true,
		})
}
