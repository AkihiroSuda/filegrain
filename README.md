# :warning: FILEgrain is abandoned in favor of stargz/CRFS. See [containerd#3731](https://github.com/containerd/containerd/issues/3731) and https://github.com/ktock/remote-snapshotter

- - -
# FILEgrain: transport-agnostic, fine-grained content-addressable container image layout

[![Build Status](https://travis-ci.org/AkihiroSuda/filegrain.svg)](https://travis-ci.org/AkihiroSuda/filegrain)
[![GoDoc](https://godoc.org/github.com/AkihiroSuda/filegrain?status.svg)](https://godoc.org/github.com/AkihiroSuda/filegrain)

FILEgrain is a (long-term) proposal to extend [OCI Image Format](https://github.com/opencontainers/image-spec) to support CAS in the granularity of file, in a transport-agnostic way.

**Your feedback is welcome.**

## Talks

- [Open Source Summit North America (September 11, 2017, Los Angeles)](https://ossna2017.sched.com/event/BDpM/filegrain-transport-agnostic-fine-grained-content-addressable-container-image-layout-akihiro-suda-ntt)
 
## Pros and Cons

Pros:
* Higher concurrency in pulling image, in a transport-agnostic way
* Files can be lazy-pulled. i.e. Files can appear at the filesystem before it is actually pulled.
* Finer deduplication granularity

Cons:
* The `blobs` directory in the image can contain a large number of files. So, `readdir()` for the directory is likely to become slow. This could be mitigated by using [external blob stores](#future-support-for-ipfs-blob-store) though.

## Format

FILEgrain defines the image manifest which is almost identical to the OCI image manifest, but different in the following points:

 * FILEgrain image manifest supports [continuity manifest](https://github.com/containerd/continuity) (`application/vnd.continuity.manifest.v0+pb` and `...+json`) as an [Image Layer Filesystem Changeset](https://github.com/opencontainers/image-spec/blob/master/layer.md). Regular files in an image are stored as OCI blob and accessed via the digest value recorded in the continuity manifest. FILEgrain still supports tar layers (`application/vnd.oci.image.layer.v1.tar` and its families), and it is even possible to put a continuity layer on top of tar layers, and vice versa. Tar layers might be useful for enforcing a lot of small files to be downloaded in batch (as a single tar file).
 * FILEgrain image manifest SHOULD have an annotation `filegrain.version=20170501`, in both the manifest JSON itself and the image index JSON. This annotation WILL change in future versions.
 
It is possible and recommended to put both a FILEgrain manifest file and an OCI manifest file in a single image.

## Example
[image index](https://github.com/opencontainers/image-spec/blob/latest/image-index.md):
(The second entry is a FILEgrain manifest)
```json
{
    "schemaVersion": 2,
    "manifests": [
	{
	    "mediaType": "application/vnd.oci.image.manifest.v1+json",
	    ...
	},
	{
	    "mediaType": "application/vnd.oci.image.manifest.v1+json",
	    ...,
	    "annotations": {
		"filegrain.version": "20170501"
	    }
	}
    ]
}
```

[image manifest](https://github.com/opencontainers/image-spec/blob/latest/image-manifest.md):
(a continuity layer on top of a tar layer)
```json
{
    "schemaVersion": 2,
    "layers": [
	{
	    "mediaType": "application/vnd.continuity.manifest.v0+json",
	    ...
	},
	{
	    "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
	    ..,
	}
    ],
    "annotations": {
	"filegrain.version": "20170501"
    }
}
```

## Distribution

FILEgrain is designed agnostic to transportation and hence can be distribeted in any way.

My personal recommendation is to just put the image directory to [IPFS](https://ipfs.io).
However, I intentionally designed FILEgrain _not_ to use IPFS multiaddr/multihash.

### Future support for IPFS blob store

So as to avoid putting a lot file into a single OCI blob directory, it might be good to consider using IPFS as an additional blob store.

IPFS store support is not yet undertaken, but it would be like this:

```json
{
    "schemaVersion": 2,
    "layers": [
	{
		"mediaType": "application/vnd.continuity.manifest.v0+json",
		...,
		"annotations": {
			"filegrain.ipfs": "QmFooBar"
		}
	}
    ],
    "annotations": {
	"filegrain.version": "2017XXXX"
    }
}
```

In this case, the layer SHOULD be fetch via IPFS multihash, rather than the digest values specified in the continuity manifest.
Also, the continuity manifest MAY omit digest values, since IPFS provides them redundantly.

Note that this is different from just putting the `blobs` directory onto IPFS, which would still create a lot of files on a single directory, when pulled from non-FILEgrain implementation.


## POC

Builder:

- [ ] Build a FILEgrain image from an existing OCI image (`--source-type oci-image`)
- [X] Build a FILEgrain image from an existing Docker image  (`--source-type docker-image`)
- [X] Build a FILEgrain image from a raw rootfs directory (`--source-type rootfs`)

Lazy Puller:

- [X] OCI-style directory on a generic filesystem (`blobs/sha256/deadbeef..`)
- [ ] Docker registry
- [ ] IPFS multihash (See [Future support for IPFS blob store](#future-support-for-ipfs-blob-store) section)

Mounter:

- [X] Read-only mount using FUSE (Linux)

Writable mount is not planned at the moment, as FILEgrain is optimized for "cattles" rather than "pets".
Users should use bind-mount or some union filesystems for `/tmp`, `/run`, and `/home`.

### POC Usage

Install FILEgrain binary:

```console
$ go get github.com/AkihiroSuda/filegrain
```

Convert a Docker image (e.g. `java:8`) to a FILEgrain image `/tmp/filegrain-image`:

```console
# filegrain build -o /tmp/filegrain-image --source-type docker-image java:8
```

Prepare an OCI bundle `/tmp/bundle.sh `from [`./oci-runtime-bundle.template`](./oci-runtime-bundle.template/README.md):
```console
# cp -r ./oci-runtime-bundle.template /tmp/bundle
# cd /tmp/bundle
# ./prepare.sh
```

Mount the local FILEgrain image `/tmp/filegrain-image` on `/tmp/bundle/rootfs`:
```console
# filegrain mount /tmp/filegrain-image /tmp/bundle/rootfs
```
In future, `filegrain mount` should support mounting remote images over Docker Registry HTTP API as well.

Open another terminal, and start runC with the bundle `/tmp/bundle`:
```console
# cd /tmp/bundle
# runc run foo
```
Instead of runc, you will be able to use `docker run` as well when Docker supports running an arbitrary OCI runtime bundle.

The container starts without pulling all the blobs.
Pulled blobs can be found on `/tmp/filegrain-blobcacheXXXXX`:

```console
# du -hs /tmp/filegrain-blobcache*
```

This directory grows as you `read(2)` files within the container rootfs.

### POC Benchmark

Please refer to [#17](https://github.com/AkihiroSuda/filegrain/issues/17).

e.g. Pulling 352MB of blobs is enough for using NLTK with 8.3GB `kaggle/python` image.

## Similar work

### Lazy distribution
- [Harter, Tyler, et al. "Slacker: Fast Distribution with Lazy Docker Containers." FAST. 2016.](https://www.usenix.org/conference/fast16/technical-sessions/presentation/harter)
- [Lestaris, George. "Alternatives to layer-based image distribution: using CERN filesystem for images." Container Camp UK. 2016.](http://www.slideshare.net/glestaris/alternatives-to-layerbased-image-distribution-using-cern-filesystem-for-images)
- [Blomer, Jakob, et al. "A Novel Approach for Distributing and Managing Container Images: Integrating CernVM File System and Mesos." MesosCon NA. 2016.](https://mesosconna2016.sched.com/event/6jtr/a-novel-approach-for-distributing-and-managing-container-images-integrating-cernvm-file-system-and-mesos-jakob-blomer-cern-jie-yu-artem-harutyunyan-mesosphere)

## FAQ

**Q. Why not just use IPFS directory? It is CAS in the granularity of file.**

A. Because IPFS does not support metadata of files. Also, this way is not transport-agnostic.

**Q. Usecases for lazy-pulling?**

A. Here are some examples I can come up with:

- Huge web application composed of a lot of static HTML and graphic files
- Huge scientific data (a content-addressable image with full code and full data would be great for reproducible research)
- Huge OS image (e.g. Windows Server, Linux with VNC desktop)
- Huge runtime (e.g. Java, dotNET)
- Huge image that is composed of multiple software stack for integration testing

Please also refer to the list of [similar work about lazy distribution](#similar-work).

**Q. Isn't it a bad idea to put a lot of file into a single blobs directory?**

A. This could be mitigated by avoid putting file into the OCI blob store, and use an external blob store instead e.g. IPFS. (go-ipfs supports [sharding](https://github.com/ipfs/go-ipfs/pull/3042)), although not transport-agnostic.
See also [an idea about future support for IPFS blob store](#future-support-for-ipfs-blob-store).

Also, there is an idea to implement sharding to the OCI native blob store: [opencontainers/image-spec#449](https://github.com/opencontainers/image-spec/issues/449).
