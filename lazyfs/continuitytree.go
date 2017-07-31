package lazyfs

import (
	"fmt"
	"os"

	"github.com/AkihiroSuda/filegrain/continuityutil"
	continuitypb "github.com/stevvooe/continuity/proto"
)

type treeItem struct {
	resource *continuitypb.Resource
	ino      uint64 // inode number
}

func loadTree(opts Options) (*nodeManager, error) {
	imageManifest, err := loadImageManifest(opts)
	if err != nil {
		return nil, err
	}
	nm := newNodeManager("/") // "/" = path sep (not root dir)
	nm.root.x = &treeItem{
		resource: &continuitypb.Resource{ // set root content (unlikely to appear in the manifest)
			Mode: uint32(os.ModeDir | 0755),
		},
	}
	ino := uint64(0)
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
			item := &treeItem{
				resource: resource,
				ino:      ino,
			}
			for _, path := range resource.Path {
				nm.insert(path, item)
			}
			ino++
		}
	}
	return nm, nil
}
