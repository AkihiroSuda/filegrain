package builder

import (
	"errors"
)

type fromOCIImageBuilder struct {
	source string
}

func NewBuilderWithOCIImage(source string) (Builder, error) {
	return &fromOCIImageBuilder{
		source: source,
	}, nil
}

func (b *fromOCIImageBuilder) Build(img, refName string) error {
	return errors.New("fromOCIImageBuilder: not implemented yet")
}
