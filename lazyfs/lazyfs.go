package lazyfs

import (
	"os"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/AkihiroSuda/filegrain/puller"
)

// FS is a READ-ONLY filesystem with lazy-pull feature.
//
// FS implements github.com/hanwen/go-fuse/fuse/pathfs.FileSystem
//
// Supported objects:
//  - directories
//  - regular files (including hardlinks) (excepts XAttrs)
//  - symbolic links
type FS struct {
	pathfs.FileSystem
	opts Options
	tree *nodeManager
}

func fuseAttrFromTreeItem(item *treeItem) *fuse.Attr {
	mode := item.resource.Mode & uint32(os.ModePerm)
	siz := item.resource.Size
	switch item.resource.Mode & uint32(os.ModeType) {
	case uint32(os.ModeDir):
		mode |= syscall.S_IFDIR
	case uint32(os.ModeSymlink):
		mode |= syscall.S_IFLNK
	case 0:
		mode |= syscall.S_IFREG
	}
	return &fuse.Attr{
		Ino:  item.ino,
		Mode: mode,
		Size: siz,
		// Times are not supported in current continuity
	}
}

func (fs *FS) lookup(name string) (*treeItem, fuse.Status) {
	n := fs.tree.lookup(name)
	if n == nil {
		return nil, fuse.ENOENT
	}
	res, ok := n.x.(*treeItem)
	if !ok {
		logrus.Errorf("can't convert %#v to *treeItem while looking up %q", n.x, name)
		return nil, fuse.EIO
	}
	return res, fuse.OK
}

func (fs *FS) GetAttr(name string, fc *fuse.Context) (*fuse.Attr, fuse.Status) {
	item, st := fs.lookup(name)
	if st != fuse.OK {
		return nil, st
	}
	attr := fuseAttrFromTreeItem(item)
	return attr, fuse.OK
}

func (fs *FS) OpenDir(name string, fc *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	n := fs.tree.lookup(name)
	if n == nil {
		return nil, fuse.ENOENT
	}
	var ents []fuse.DirEntry
	for k, v := range n.m {
		item, ok := v.x.(*treeItem)
		if !ok {
			logrus.Errorf("can't convert %#v to *treeItem while opendir %q, %q", n.x, name, k)
			return nil, fuse.EIO
		}
		mode := item.resource.Mode // FIXME?
		ents = append(ents, fuse.DirEntry{
			Name: k,
			Mode: mode,
		})
	}
	return ents, fuse.OK
}

func (fs *FS) Open(name string, flags uint32, fc *fuse.Context) (nodefs.File, fuse.Status) {
	item, st := fs.lookup(name)
	if st != fuse.OK {
		return nil, st
	}
	return newFile(fs.opts, item)
}

func (fs *FS) Readlink(name string, fc *fuse.Context) (string, fuse.Status) {
	item, st := fs.lookup(name)
	if st != fuse.OK {
		return "", st
	}
	return item.resource.Target, fuse.OK
}

type Options struct {
	Mountpoint string
	Puller     puller.Puller
	Image      string
	RefName    string
}

func NewFS(opts Options) (*FS, error) {
	tree, err := loadTree(opts)
	if err != nil {
		return nil, err
	}
	fs := &FS{
		FileSystem: pathfs.NewReadonlyFileSystem(pathfs.NewDefaultFileSystem()),
		opts:       opts,
		tree:       tree,
	}
	return fs, nil
}

func NewServer(fs *FS, opts *nodefs.Options) (*fuse.Server, error) {
	nfs := pathfs.NewPathNodeFs(fs, nil)
	conn := nodefs.NewFileSystemConnector(nfs.Root(), opts)
	return fuse.NewServer(conn.RawFS(), fs.opts.Mountpoint,
		&fuse.MountOptions{
			FsName: fs.opts.Mountpoint + ":" + fs.opts.RefName,
			Name:   "filegrain.lazyfs",
		})
}
