// Package lzma is a thin wrapper around the C lzma library.
//
// The emphasis is on the word "thin". This package does not provide an
// idiomatic Go API; rather, it simply wraps C functions and types with
// analogous Go functions and types.
// A nice Go API should be built on top of this package.
//
// The documentation for each type and function in this package generally just
// contains a reference to
// to the underlying C type or function in the /src/liblzma/api/ directory of the
// upstream C repository. Full documentation for the type and function can be found
// by looking at the excellent documentation on the C side.
package lzma

/*
#cgo LDFLAGS: -llzma
#include <stdlib.h>
#include "upstream/src/liblzma/api/lzma.h"

// The lzma library requires that the stream be initialized to the value of the macro
// LZMA_STREAM_INIT. Because this is a macro it has no type. This function exists to cast the
// macro to the stream type.
lzma_stream new_stream() {
	lzma_stream strm = LZMA_STREAM_INIT;
	return strm;
}
*/
import "C"
import (
	"unsafe"
)

// Return corresponds to the lzma_ret type in base.h.
type Return int

const (
	Ok               Return = 0
	StreamEnd               = 1
	NoCheck                 = 2
	UnsupportedCheck        = 3
	GetCheck                = 4
	MemoryError             = 5
	MemoryLimitError        = 6
	FormatError             = 7
	OptionsError            = 8
	DataError               = 9
	BufferError             = 10
	ProgrammingError        = 11
	SeekNeeded              = 12
)

func (r Return) String() string {
	switch r {
	case Ok:
		return "OK"
	case StreamEnd:
		return "STREAM_END"
	case NoCheck:
		return "NO_CHECK"
	case UnsupportedCheck:
		return "UNSUPPORTED_CHECK"
	case GetCheck:
		return "GET_CHECK"
	case MemoryError:
		return "MEMORY_ERROR"
	case MemoryLimitError:
		return "MEMORY_LIMIT_ERROR"
	case FormatError:
		return "FORMAT_ERROR"
	case OptionsError:
		return "OPTIONS_ERROR"
	case DataError:
		return "DATA_ERROR"
	case BufferError:
		return "BUFFER_ERROR"
	case ProgrammingError:
		return "PROGRAMMING_ERROR"
	case SeekNeeded:
		return "SEEK_NEEDED"
	}
	return "UNKNOWN_RESULT"
}

// Action corresponds to the lzma_action type in base.h.
type Action int

const (
	Run         Action = 0
	SyncFlush          = 1
	FullFlush          = 2
	Finish             = 3
	FullBarrier        = 4
)

type cBuffer struct {
	start *C.uint8_t
	len   C.size_t
}

func (buf *cBuffer) set(p []byte) {
	// TODO: instead of allocating for each SetInput, allocate once and copy over?
	buf.clear()
	buf.start = (*C.uint8_t)(C.CBytes(p))
	buf.len = C.size_t(len(p))
}

func (buf *cBuffer) read(length int) []byte {
	return C.GoBytes(unsafe.Pointer(buf.start), C.int(length))
}

func (buf *cBuffer) clear() {
	if buf.start != nil {
		C.free(unsafe.Pointer(buf.start))
	}
	buf.start = nil
	buf.len = 0
}

// This was chosen arbitrarily but seems to work fine in practice
const outputBufferLength = 1024

// Stream wraps lzma_stream in base.h and the input and output buffers that the lzma_stream type
// requires to exist.
//
// The lzma_stream type operates on the two buffers but does not take ownership of them. This
// type thus contains handling for these buffers. This part of the package is the most Go-like
// because it needs to map from Go slices to C arrays, and ultimately hide the C implementation
// details.
type Stream struct {
	cStream C.lzma_stream
	input   cBuffer
	output  cBuffer
}

// NewStream returns a new stream.
func NewStream() *Stream {
	stream := Stream{
		cStream: C.new_stream(),
	}
	// TODO: this is very memory inefficient!
	p := make([]byte, outputBufferLength)
	stream.output.set(p)
	stream.cStream.next_out = stream.output.start
	stream.cStream.avail_out = stream.output.len
	return &stream
}

// AvailIn returns the number of bytes that have been placed in the input buffer using the SetInput
// method that have yet to be processed by the stream.
func (stream *Stream) AvailIn() int {
	return int(stream.cStream.avail_in)
}

// TotalIn returns the total number of bytes that have been read from the input buffer.
func (stream *Stream) TotalIn() int {
	return int(stream.cStream.total_in)
}

// AvailOut returns the number of bytes that the stream has written into the output buffer that
// have yet to be read using the Output method.
func (stream *Stream) AvailOut() int {
	return int(stream.cStream.avail_out)
}

// TotalOut returns the total number of bytes that have been written to the input buffer
func (stream *Stream) TotalOut() int {
	return int(stream.cStream.total_out)
}

// SetInput sets the input buffer of the stream to be the provided bytes. Note this overwrites
// any data that is already in the input buffer, so before calling SetInput it's best to verify
// that AvailIn returns 0.
func (stream *Stream) SetInput(p []byte) {
	stream.input.set(p)
	stream.cStream.next_in = stream.input.start
	stream.cStream.avail_in = stream.input.len
}

// Output returns all bytes that have been written to the output buffer by the stream, and resets
// the output buffer.
func (stream *Stream) Output() []byte {
	b := stream.output.read(int(stream.output.len - stream.cStream.avail_out))
	stream.cStream.next_out = stream.output.start
	stream.cStream.avail_out = stream.output.len
	return b
}

// Close closes the stream and releases C memory that has been allocated by the type.
func (stream *Stream) Close() {
	stream.input.clear()
	stream.output.clear()
	// TODO: move lzma_end to its own function
	C.lzma_end(&stream.cStream)
}

// EasyEncoder wraps lzma_easy_encoder in container.h.
func EasyEncoder(stream *Stream, preset int) Return {
	// TODO: do integrity checking
	return Return(C.lzma_easy_encoder(&stream.cStream, C.uint(preset), 0))
}

// StreamDecoder wraps lzma_stream_decoder in container.h.
func StreamDecoder(stream *Stream) Return {
	// TODO: do integrity checking
	return Return(C.lzma_stream_decoder(&stream.cStream, C.UINT64_MAX, 0))
}

// Code wraps lzma_code in base.h.
func Code(stream *Stream, action Action) Return {
	return Return(C.lzma_code(&stream.cStream, C.lzma_action(action)))
}
