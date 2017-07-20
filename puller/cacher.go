package puller

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/go-digest"
	spec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/AkihiroSuda/filegrain/image"
)

type pullStatus int

const (
	pullStatusUnknown pullStatus = iota
	pullStatusPulling
	pullStatusPulled
)

type BlobCacher struct {
	cachePath string
	puller    Puller

	pullStatus     map[digest.Digest]pullStatus
	pullStatusCond *sync.Cond

	pulledBlobBytes uint64 // atomic
}

func NewBlobCacher(cachePath string, puller Puller) (*BlobCacher, error) {
	if _, err := os.Stat(cachePath); err != nil {
		return nil, err
	}
	cacher := &BlobCacher{
		cachePath:       cachePath,
		puller:          puller,
		pullStatus:      make(map[digest.Digest]pullStatus, 0),
		pullStatusCond:  sync.NewCond(&sync.Mutex{}),
		pulledBlobBytes: 0,
	}
	// TODO: load cacher.pulled
	return cacher, nil
}

func (p *BlobCacher) PullBlob(img string, d digest.Digest) (image.BlobReader, error) {
	if err := p.cacheBlobIfNotYet(img, d); err != nil {
		return nil, err
	}
	return p.openCachedBlob(img, d)
}

func (p *BlobCacher) cacheBlobIfNotYet(img string, d digest.Digest) error {
	alreadyCached := false
	for {
		p.pullStatusCond.L.Lock()
		st, ok := p.pullStatus[d]
		if ok && st == pullStatusPulling {
			p.pullStatusCond.Wait()
			p.pullStatusCond.L.Unlock()
		} else {
			p.pullStatusCond.L.Unlock()
			alreadyCached = st == pullStatusPulled
			break
		}
	}
	if !alreadyCached {
		return p.cacheBlob(img, d)
	}
	return nil
}

func (p *BlobCacher) cacheBlob(img string, d digest.Digest) error {
	// TODO: use hardlink when possible?
	logrus.Debugf("caching blob: %s", d)
	p.pullStatusCond.L.Lock()
	p.pullStatus[d] = pullStatusPulling
	p.pullStatusCond.L.Unlock()
	r, err := p.puller.PullBlob(img, d)
	if err != nil {
		return err
	}
	w, err := image.NewBlobWriter(p.cachePath, d.Algorithm())
	if err != nil {
		return err
	}
	copied, err := io.Copy(w, r)
	if err != nil {
		return err
	}
	if err := r.Close(); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	if dd := w.Digest(); dd == nil || *dd != d {
		return fmt.Errorf("expected %q, got %q", d, dd)
	}
	totalCopied := atomic.AddUint64(&p.pulledBlobBytes, uint64(copied))
	logrus.Infof("Total blob bytes pulled in this session: %d B", totalCopied)
	p.pullStatusCond.L.Lock()
	p.pullStatus[d] = pullStatusPulled
	p.pullStatusCond.L.Unlock()
	p.pullStatusCond.Broadcast()
	return nil
}

func (p *BlobCacher) openCachedBlob(img string, d digest.Digest) (image.BlobReader, error) {
	return image.GetBlobReader(p.cachePath, d)
}

func (p *BlobCacher) PullIndex(img string) (*spec.Index, error) {
	return p.puller.PullIndex(img)
}
