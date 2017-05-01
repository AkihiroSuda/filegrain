package builder

import (
	"fmt"
	"path/filepath"

	"github.com/AkihiroSuda/filegrain/cp"
	"github.com/AkihiroSuda/filegrain/image"
	"github.com/AkihiroSuda/filegrain/image/imageutil"
	"github.com/AkihiroSuda/filegrain/version"

	"github.com/Sirupsen/logrus"
	progressbar "github.com/cheggaaa/pb"
	"github.com/golang/protobuf/proto"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	spec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stevvooe/continuity"
	pb "github.com/stevvooe/continuity/proto"
)

type fromRootFSBuilder struct {
	source string
}

func NewBuilderWithRootFS(source string) (Builder, error) {
	_, err := image.ReadImageLayout(source)
	if err == nil {
		return nil, fmt.Errorf("source %q seems an OCI image, please specify a valid rootfs instead", source)
	}
	return &fromRootFSBuilder{
		source: source,
	}, nil
}

func (b *fromRootFSBuilder) Build(img, refName string) error {
	logrus.Infof("Initializing %s as an OCI image (spec %s)", img, specs.Version)
	if err := image.Init(img); err != nil {
		return err
	}
	contM, err := buildContinuityManifest(b.source)
	if err != nil {
		return err
	}
	contMDesc, err := putContinuityManifestBlobs(img, b.source, contM)
	if err != nil {
		return err
	}
	imageMDesc, err := putImageManifestBlobs(img, contMDesc)
	if err != nil {
		return err
	}
	logrus.Infof("Built image manifest %s", imageMDesc.Digest)
	if refName != "" {
		if imageMDesc.Annotations == nil {
			imageMDesc.Annotations = make(map[string]string, 0)
		}
		logrus.Infof("Tag: %s", refName)
		imageMDesc.Annotations[image.RefNameAnnotation] = refName
	}
	if err := image.PutManifestDescriptorToIndex(img, imageMDesc); err != nil {
		return err
	}
	return nil
}

func buildContinuityManifest(source string) (*continuity.Manifest, error) {
	ctx, err := continuity.NewContext(source)
	if err != nil {
		return nil, err
	}

	return continuity.BuildManifest(ctx)
}

func continuityManifestToPB(m *continuity.Manifest) (*pb.Manifest, error) {
	bytes, err := continuity.Marshal(m)
	if err != nil {
		return nil, err
	}
	var bm pb.Manifest
	if err := proto.Unmarshal(bytes, &bm); err != nil {
		return nil, err
	}
	return &bm, nil
}

// puts rootfs blobs and continuity manifest blob.
// returns the descriptor of the continuity manifest blob.
func putContinuityManifestBlobs(img, source string, manifest *continuity.Manifest) (*spec.Descriptor, error) {
	pbManifest, err := continuityManifestToPB(manifest)
	if err != nil {
		return nil, err
	}
	bar := progressbar.StartNew(len(pbManifest.Resource))
	for _, r := range pbManifest.Resource {
		bar.Increment()
		for _, ds := range r.Digest {
			d, err := digest.Parse(ds)
			if err != nil {
				return nil, err // FIXME: can be skipped, generally
			}
			blobPath := filepath.Join(img, "blobs", string(d.Algorithm()), d.Hex())
			if len(r.Path) == 0 {
				return nil, fmt.Errorf("no path for %s", d)
			}
			blobSourcePath := filepath.Join(source, r.Path[0])
			if err := cp.CopyFile(blobPath, blobSourcePath); err != nil {
				return nil, err
			}
		}
	}
	bar.Finish()
	manifestBytes, err := proto.Marshal(pbManifest)
	if err != nil {
		return nil, err
	}
	d, err := image.WriteBlob(img, manifestBytes)
	if err != nil {
		return nil, err
	}
	return &spec.Descriptor{
		MediaType: "application/vnd.continuity.manifest.v0+pb", // TODO: JSON
		Digest:    d,
		Size:      int64(len(manifestBytes)),
	}, nil
}

// puts image manifest blob and its deps (e.g. config).
// returns the descriptor of the image manifest blob.
func putImageManifestBlobs(img string, continuityManifest *spec.Descriptor) (*spec.Descriptor, error) {
	config := &spec.Image{
		Architecture: "amd64", // FIXME
		OS:           "linux", // FIXME
		RootFS: spec.RootFS{
			Type: "layers",
			DiffIDs: []digest.Digest{
				continuityManifest.Digest, // FIXME: ensure uncompressed
			},
		},
	}
	configDesc, err := imageutil.WriteJSONBlob(img, config, spec.MediaTypeImageConfig)
	if err != nil {
		return nil, err
	}
	manifest := &spec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		Config: *configDesc,
		Layers: []spec.Descriptor{
			*continuityManifest,
		},
		Annotations: map[string]string{
			version.VersionAnnotation: version.Version,
		},
	}
	desc, err := imageutil.WriteJSONBlob(img, manifest, spec.MediaTypeImageManifest)
	if err != nil {
		return nil, err
	}
	desc.Annotations = manifest.Annotations
	return desc, err
}
