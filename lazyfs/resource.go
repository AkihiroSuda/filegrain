package lazyfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	spec "github.com/opencontainers/image-spec/specs-go/v1"
	continuitypb "github.com/stevvooe/continuity/proto"

	"github.com/AkihiroSuda/filegrain/image"
)

func loadContinuityPBManifest(opts Options, desc *spec.Descriptor) (*continuitypb.Manifest, error) {
	manifestBlob, err := loadBlobWithDescriptor(opts, desc)
	if err != nil {
		return nil, err
	}
	var bm continuitypb.Manifest

	if err := proto.Unmarshal(manifestBlob, &bm); err != nil {
		return nil, err
	}
	return &bm, nil
}

func loadImageManifest(opts Options) (*spec.Manifest, error) {
	idx, err := opts.Puller.PullIndex(opts.Image)
	if err != nil {
		return nil, err
	}
	var imageManifestDesc *spec.Descriptor
	for _, m := range idx.Manifests {
		mRefName, ok := m.Annotations[image.RefNameAnnotation]
		if ok && mRefName == opts.RefName {
			imageManifestDesc = &m
			break
		}
	}
	if imageManifestDesc == nil {
		return nil, fmt.Errorf("unknown reference name: %q", opts.RefName)
	}
	imageManifestBlob, err := loadBlobWithDescriptor(opts, imageManifestDesc)
	if err != nil {
		return nil, err
	}
	var imageManifest spec.Manifest
	if err := json.Unmarshal(imageManifestBlob, &imageManifest); err != nil {
		return nil, err
	}
	return &imageManifest, nil
}

func loadBlobWithDescriptor(opts Options, desc *spec.Descriptor) ([]byte, error) {
	r, err := opts.Puller.PullBlob(opts.Image, desc.Digest)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}
