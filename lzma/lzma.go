// Package lzma is a thin wrapper around the C lzma library.
//
// The emphasis is on the word "thin". This package does not provide an
// idiomatic Go API; rather, it simply wraps C functions and types with
// analogous Go functions and types.
package lzma

/*
#cgo LDFLAGS: -llzma -L/home/james/git/xz/
#include "stdlib.h"
#include <stdio.h>
#include "../xz/src/liblzma/api/lzma.h"

// This function is needed to cast from the macro LZMA_STREAM_INIT to lzma_stream
// in a way the Go and C compilers understands.
lzma_stream new_stream() {
	lzma_stream strm = LZMA_STREAM_INIT;
	return strm;
}

void read_out(lzma_stream* strm, uint8_t* buf) {
	int i;
	for (i=-24; i<500; i++) {
		printf("%d ", strm->next_out[i]);
		buf[i] = strm->next_out[i];
	}
	fflush(stdout);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type Result int

const (
	Ok Result = 0
	StreamEnd = 1
	NoCheck = 2
	UnsupportedCheck = 3
	GetCheck = 4
	MemoryError = 5
	MemoryLimitError = 6
	FormatError = 7
	OptionsError = 8
	DataError = 9
	BufferError = 10
	ProgrammingError = 11
	SeekNeeded = 12
)

type Action int

const (
	Run Action = 0
	SyncFlush = 1
	FullFlush = 2
	Finish = 3
	FullBarrier = 4
)

// Stream wraps lzma_stream in base.h
type Stream struct {
	cStream C.lzma_stream
	input cBuffer
	output cBuffer
}

func NewStream() *Stream {
	return &Stream{
		cStream: C.new_stream(),
	}
}

type cBuffer struct {
	start *C.uint8_t
	len C.size_t
}

func (buf *cBuffer) set(p []byte) {
	// TODO: instead of allocating for each SetInput, allocate once and copy over?
	//if stream.cStream.next_in != nil {
	//	C.free(unsafe.Pointer(stream.cStream.next_in))
	//}
	buf.start = (*C.uint8_t)(C.CBytes(p))
	buf.len = C.size_t(len(p))
}

func (buf *cBuffer) read() []byte {
	return C.GoBytes(unsafe.Pointer(buf.start), C.int(buf.len))
}

// TODO: to go from the C buf to a go slice:
// To create a Go slice with the contents of C.my_buf:
//
// arr := C.GoBytes(unsafe.Pointer(&C.my_buf), C.BUF_SIZE)
// SetInput

// TODO Other direction:
// C.CBytes([]byte) unsafe.Pointer

func (stream *Stream) SetInput(p []byte) {
	stream.input.set(p)
	stream.cStream.next_in = stream.input.start
	stream.cStream.avail_in = stream.input.len
}

func (stream *Stream) SetOutputLen(length int) {
	// TODO: this is very memory inefficient!
	p := make([]byte, length)
	stream.output.set(p)
	stream.cStream.next_out = stream.output.start
	stream.cStream.avail_out = stream.output.len
}

func (stream *Stream) Output() []byte {


	fmt.Println(stream.output.read())
	fmt.Println(stream.output.read()[:int(stream.cStream.total_out)])
	fmt.Println("total_in", stream.cStream.total_in)
	fmt.Println("total_out", stream.cStream.total_out)
	fmt.Println("avail_out (actual)", stream.cStream.avail_out)

	return stream.output.read()[:int(stream.cStream.total_out)]
	//C.read_out(&stream.cStream, (*C.uint8_t)(unsafe.Pointer(&p[0])))

}

func (stream *Stream) Close() {
	//if stream.cStream.next_in != nil {
	//	C.free(unsafe.Pointer(stream.cStream.next_in))
	//}
	//if stream.cStream.next_out != nil {
	//	C.free(unsafe.Pointer(stream.cStream.next_out))
	//}
	C.lzma_end(&stream.cStream)
}

// EasyEncoder wraps lzma_easy_encoder in container.h
func EasyEncoder(stream *Stream, preset int) Result {
	// TODO: support check
	return Result(C.lzma_easy_encoder(&stream.cStream, C.uint(preset), 0))
}

// Code wraps lzma_code in base.h
func Code(stream *Stream, action Action) Result {
	return Result(C.lzma_code(&stream.cStream, C.lzma_action(action)))
}
