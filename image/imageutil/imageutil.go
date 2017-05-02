package imageutil

import (
	"encoding/json"

	spec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/AkihiroSuda/filegrain/image"
)

func WriteJSONBlob(img string, x interface{}, mediaType string) (*spec.Descriptor, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}
	d, err := image.WriteBlob(img, b)
	if err != nil {
		return nil, err
	}
	return &spec.Descriptor{
		MediaType: mediaType,
		Digest:    d,
		Size:      int64(len(b)),
	}, nil
}
