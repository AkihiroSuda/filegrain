package dummycontent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	progressbar "github.com/cheggaaa/pb"
	"github.com/opencontainers/go-digest"
	"github.com/stevvooe/continuity"
)

type DummyRegularFileContent struct {
	Size    int64
	Digests []digest.Digest
}

// dummyRegularFile implements continuity.RegularFile
type dummyRegularFile struct {
	continuity.RegularFile
	dummyContent       []byte
	dummyContentDigest digest.Digest
}

func newDummyRegularfile(rf continuity.RegularFile) (*dummyRegularFile, error) {
	dummyContent := DummyRegularFileContent{
		Size:    rf.Size(),
		Digests: rf.Digests(),
	}
	dummyContentB, err := json.Marshal(dummyContent)
	if err != nil {
		return nil, err
	}
	return &dummyRegularFile{
		RegularFile:        rf,
		dummyContent:       dummyContentB,
		dummyContentDigest: digest.FromBytes(dummyContentB),
	}, nil
}

func (rf *dummyRegularFile) Size() int64 {
	return int64(len(rf.dummyContent))
}

func (rf *dummyRegularFile) Digests() []digest.Digest {
	return []digest.Digest{rf.dummyContentDigest}
}

type dummyContinuity struct {
	m        map[digest.Digest][]byte
	manifest *continuity.Manifest
}

func newDummyContinuity(manifest *continuity.Manifest) (*dummyContinuity, error) {
	dc := &dummyContinuity{
		m:        make(map[digest.Digest][]byte, 0),
		manifest: &continuity.Manifest{},
	}
	for _, r := range manifest.Resources {
		if rf, ok := r.(continuity.RegularFile); ok {
			dummyRF, err := newDummyRegularfile(rf)
			if err != nil {
				return nil, err
			}
			dc.m[dummyRF.dummyContentDigest] = dummyRF.dummyContent
			r = dummyRF
		}
		dc.manifest.Resources = append(dc.manifest.Resources, r)
	}
	return dc, nil
}

// dummyContinuityContentProvider implements continuity.ContentProvider
type dummyContinuityContentProvider struct {
	continuity *dummyContinuity
}

func (cp *dummyContinuityContentProvider) Reader(d digest.Digest) (io.ReadCloser, error) {
	b, ok := cp.continuity.m[d]
	if !ok {
		return nil, fmt.Errorf("not found: %s", d)
	}
	return ioutil.NopCloser(bytes.NewReader(b)), nil
}

func ApplyContinuityManifestWithDummyContents(dir string, manifest *continuity.Manifest) error {
	dummyContinuity, err := newDummyContinuity(manifest)
	if err != nil {
		return err
	}
	provider := &dummyContinuityContentProvider{
		continuity: dummyContinuity,
	}
	driver, err := continuity.NewSystemDriver()
	if err != nil {
		return err
	}
	ctx, err := continuity.NewContextWithOptions(dir,
		continuity.ContextOptions{
			Driver:   driver,
			Provider: provider,
		})
	logrus.Infof("Converting continuity manifest (%d resources) to a file-based database under %s",
		len(dummyContinuity.manifest.Resources), dir)
	logrus.Infof("NOTE: this operation is slow because it creates a bunch of files, but could be mitigated by future memDB-based implementation. So this slowness does not hurt for POC!")
	bar := progressbar.StartNew(len(dummyContinuity.manifest.Resources))
	for _, resource := range dummyContinuity.manifest.Resources {
		bar.Increment()
		if err := ctx.Apply(resource); err != nil {
			return err
		}
	}
	bar.Finish()
	return nil
}
