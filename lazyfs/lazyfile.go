package lazyfs

import (
	"io"

	"github.com/AkihiroSuda/filegrain/image"
	"github.com/Sirupsen/logrus"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/opencontainers/go-digest"
)

type file struct {
	nodefs.File
	opts       Options
	item       *treeItem
	blobReader image.BlobReader
}

func newFile(opts Options, item *treeItem) (nodefs.File, fuse.Status) {
	if len(item.resource.Digest) == 0 {
		logrus.Errorf("no digest for %v", item.resource.Path)
		return nil, fuse.EIO
	}
	dgst := digest.Digest(item.resource.Digest[0])
	blobReader, err := opts.Puller.PullBlob(opts.Image, dgst)
	if err != nil {
		logrus.Errorf("error while pulling %s (%v): %v", dgst, item.resource.Path, err)
		return nil, fuse.EIO
	}
	f := &file{
		File:       nodefs.NewDefaultFile(),
		item:       item,
		opts:       opts,
		blobReader: blobReader,
	}
	cached := &nodefs.WithFlags{
		File:      f,
		FuseFlags: fuse.FOPEN_KEEP_CACHE,
	}
	return cached, fuse.OK
}

func (f *file) GetAttr(out *fuse.Attr) fuse.Status {
	*out = *fuseAttrFromTreeItem(f.item)
	return fuse.OK
}

func (f *file) Release() {
	if err := f.blobReader.Close(); err != nil {
		logrus.Errorf("error while closing %v: %v",
			f.item.resource.Path, err)
	}
}

func (f *file) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	if n, err := f.blobReader.ReadAt(buf, off); err == io.EOF {
		buf = buf[:n]
	} else if err != nil {
		logrus.Errorf("error while reading %d bytes at %d (%v): %v",
			len(buf), off, f.item.resource.Path, err)
		return nil, fuse.EIO
	}
	return fuse.ReadResultData(buf), fuse.OK
}
