package lazyfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	spec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stevvooe/continuity"

	"github.com/AkihiroSuda/filegrain/continuityutil"
	"github.com/AkihiroSuda/filegrain/image"
	"github.com/AkihiroSuda/filegrain/lazyfs/dummycontent"
)

func loadContinuityManifest(opts Options, desc *spec.Descriptor) (*continuity.Manifest, error) {
	manifestBlob, err := loadBlobWithDescriptor(opts, desc)
	if err != nil {
		return nil, err
	}
	manifest, err := continuity.Unmarshal(manifestBlob)
	if err != nil {
		return nil, err
	}
	return manifest, nil
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

func loadUnderlier(opts Options, dir string) error {
	imageManifest, err := loadImageManifest(opts)
	if err != nil {
		return err
	}
	for _, layer := range imageManifest.Layers {
		// TODO: support mixing up tar layers and continutiy layers..
		if layer.MediaType != continuityutil.MediaTypeManifestV0Protobuf {
			return fmt.Errorf("unsupported layer mediaType: %s", layer.MediaType)
		}
		manifest, err := loadContinuityManifest(opts, &layer)
		if err != nil {
			return err
		}
		if err := dummycontent.ApplyContinuityManifestWithDummyContents(dir, manifest); err != nil {
			return err
		}
	}
	return nil
}
