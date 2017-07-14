package lazyfs

import (
	"fmt"
	"os"

	"github.com/AkihiroSuda/filegrain/continuityutil"
	continuitypb "github.com/stevvooe/continuity/proto"
)

func loadTree(opts Options) (*nodeManager, error) {
	imageManifest, err := loadImageManifest(opts)
	if err != nil {
		return nil, err
	}
	nm := newNodeManager("/")           // "/" = path sep (not root dir)
	nm.root.x = &continuitypb.Resource{ // set root content (unlikely to appear in the manifest)
		Mode: uint32(os.ModeDir | 0755),
	}
	for _, layer := range imageManifest.Layers {
		// TODO: support mixing up tar layers and continutiy layers..
		if layer.MediaType != continuityutil.MediaTypeManifestV0Protobuf {
			return nil, fmt.Errorf("unsupported layer mediaType: %s", layer.MediaType)
		}
		pb, err := loadContinuityPBManifest(opts, &layer)
		if err != nil {
			return nil, err
		}
		for _, resource := range pb.Resource {
			for _, path := range resource.Path {
				nm.insert(path, resource)
			}
		}
	}
	return nm, nil
}
