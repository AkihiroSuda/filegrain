# Filegrain: transport-agnostic, fine-grained content-addressable container image layout

Filegrain is a (long-term) proposal to extend [OCI Image Format](https://github.com/opencontainers/image-spec) to support CAS in the granularity of file, in a transport-agnostic way.

**Your feedback is welcome.**
 
## Pros and Cons

Pros:
* Higher concurrency in pulling image, in a transport-agnostic way
* Files can be lazy-pulled. i.e. Files can appear at the filesystem before it is actually pulled.
* Finer deduplication granularity

Cons:
* The `blobs` directory in the image can contain a large number of files. So, `readdir()` for the directory is likely to become slow. There are some proposals to mitigate this by sharding the `blobs` directory. ([opencontainers/image-spec#449](https://github.com/opencontainers/image-spec/issues/449))

## Format

Filegrain defines the image manifest which is almost identical to the OCI image manifest, but different in the following points:

 * Filegrain image manifest supports [continuity manifest](https://github.com/stevvooe/continuity) (`application/vnd.continuity.manifest.v0+pb`) as an [Image Layer Filesystem Changeset](https://github.com/opencontainers/image-spec/blob/aad7f240f0c544dcafc9cba98cbf0932a0d068ef/layer.md). Regular files in an image are stored as OCI blob and accessed via the digest value recorded in the continuity manifest. Filegrain still supports tar layers (`application/vnd.oci.image.layer.v1.tar` and its families), and it is even possible to put a continuity layer on top of tar layers, and vice versa. However, it is recommended to compose a manifest of a single continuity layer.
 * The media type of Filegrain image manifest is `application/vnd.filegrain.image.manifest.v0+json`, rather than `application/vnd.oci.image.manifest.v1+json` _at the moment_. If Filegrain can get merged to the OCI spec in the future, the media type should be `application/vnd.oci.image.manifest.vN+json`.
 
It is possible and strongly recommended to put both a Filegrain manifest file and an OCI manifest file in a single image.
i.e. the [image index](https://github.com/opencontainers/image-spec/blob/ab461b048bd1c8b6077d8e96936f706a518233c2/image-index.md) for such an image would be like this:
```json
{
	"schemaVersion": 2,
	"manifests": [
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			...
		}
		{
			"mediaType": "application/vnd.filegrain.image.manifest.v0+json",
			...
		}
	]
}
```

## Distribution

Filegrain is designed agnostic to transportation and hence can be distribeted in any way.

My personal recommendation is to just put the image directory to [IPFS](https://ipfs.io).
However, I intentionally designed Filegrain _not_ to use IPFS multiaddr/multihash.

## POC

N/A yet.

Plan:

 * Converter:
   * Convert OCI image to Filegrain
 * Lazy Puller and Mounter (FUSE):
   * Multiple lazy-puller backends (e.g. IPFS, git)
   * Emulate slow network so as to show the effect of lazy layer distribution

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

A. Yes, the blob system should be refined on the OCI side. Please refer to [opencontainers/image-spec#449](https://github.com/opencontainers/image-spec/issues/449).

Another idea is to just avoid putting files into the blobs directory, and use IPFS instead.
But this is not transport-agnostic.
