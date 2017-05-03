package puller

import (
	"github.com/opencontainers/go-digest"
	spec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/AkihiroSuda/filegrain/image"
)

// LocalPuller lacks caching. Use with BlobCacher.
type LocalPuller struct {
}

func NewLocalPuller() *LocalPuller {
	return &LocalPuller{}
}

func (p *LocalPuller) PullBlob(img string, d digest.Digest) (image.BlobReader, error) {
	return image.GetBlobReader(img, d)
}

func (p *LocalPuller) PullIndex(img string) (*spec.Index, error) {
	return image.ReadIndex(img)
}
