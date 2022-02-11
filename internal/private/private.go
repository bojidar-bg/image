package private

import (
	"context"
	"io"

	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/types"
)

// ImageDestination is an internal extension to the types.ImageDestination
// interface.
type ImageDestination interface {
	types.ImageDestination

	// PutBlobWithOptions writes contents of stream and returns data representing the result.
	// inputInfo.Digest can be optionally provided if known; if provided, and stream is read to the end without error, the digest MUST match the stream contents.
	// inputInfo.Size is the expected length of stream, if known.
	// inputInfo.MediaType describes the blob format, if known.
	// WARNING: The contents of stream are being verified on the fly.  Until stream.Read() returns io.EOF, the contents of the data SHOULD NOT be available
	// to any other readers for download using the supplied digest.
	// If stream.Read() at any time, ESPECIALLY at end of input, returns an error, PutBlob MUST 1) fail, and 2) delete any data stored so far.
	PutBlobWithOptions(ctx context.Context, stream io.Reader, inputInfo types.BlobInfo, options PutBlobOptions) (types.BlobInfo, error)

	// TryReusingBlobWithOptions checks whether the transport already contains, or can efficiently reuse, a blob, and if so, applies it to the current destination
	// (e.g. if the blob is a filesystem layer, this signifies that the changes it describes need to be applied again when composing a filesystem tree).
	// info.Digest must not be empty.
	// If the blob has been successfully reused, returns (true, info, nil); info must contain at least a digest and size, and may
	// include CompressionOperation and CompressionAlgorithm fields to indicate that a change to the compression type should be
	// reflected in the manifest that will be written.
	// If the transport can not reuse the requested blob, TryReusingBlob returns (false, {}, nil); it returns a non-nil error only on an unexpected failure.
	TryReusingBlobWithOptions(ctx context.Context, info types.BlobInfo, options TryReusingBlobOptions) (bool, types.BlobInfo, error)
}

// PutBlobOptions are used in PutBlobWithOptions.
type PutBlobOptions struct {
	// Cache to look up blob infos.
	Cache types.BlobInfoCache
	// Denotes whether the blob is a config or not.
	IsConfig bool
	// Indicates an empty layer.
	EmptyLayer bool
	// The corresponding index in the layer slice.
	LayerIndex *int
}

// TryReusingBlobOptions are used in TryReusingBlobWithOptions.
type TryReusingBlobOptions struct {
	// Cache to look up blob infos.
	Cache types.BlobInfoCache
	// Use an equivalent of the desired blob.
	CanSubstitute bool
	// Indicates an empty layer.
	EmptyLayer bool
	// The corresponding index in the layer slice.
	LayerIndex *int
	// The reference of the image that contains the target blob.
	SrcRef reference.Named
}

// ImageSourceChunk is a portion of a blob.
// This API is experimental and can be changed without bumping the major version number.
type ImageSourceChunk struct {
	Offset uint64
	Length uint64
}

// ImageSourceSeekable is an image source that permits to fetch chunks of the entire blob.
// This API is experimental and can be changed without bumping the major version number.
type ImageSourceSeekable interface {
	// GetBlobAt returns a stream for the specified blob.
	// The specified chunks must be not overlapping and sorted by their offset.
	GetBlobAt(context.Context, types.BlobInfo, []ImageSourceChunk) (chan io.ReadCloser, chan error, error)
}

// ImageDestinationPartial is a service to store a blob by requesting the missing chunks to a ImageSourceSeekable.
// This API is experimental and can be changed without bumping the major version number.
type ImageDestinationPartial interface {
	// PutBlobPartial writes contents of stream and returns data representing the result.
	PutBlobPartial(ctx context.Context, stream ImageSourceSeekable, srcInfo types.BlobInfo, cache types.BlobInfoCache) (types.BlobInfo, error)
}

// BadPartialRequestError is returned by ImageSourceSeekable.GetBlobAt on an invalid request.
type BadPartialRequestError struct {
	Status string
}

func (e BadPartialRequestError) Error() string {
	return e.Status
}