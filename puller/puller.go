package puller

import (
	"github.com/opencontainers/go-digest"
	spec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/AkihiroSuda/filegrain/image"
)

type Puller interface {
	PullBlob(img string, d digest.Digest) (image.BlobReader, error)
	PullIndex(img string) (*spec.Index, error)
}
