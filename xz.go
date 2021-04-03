package main

import "C"
import (
	"bytes"
	"fmt"
	"github.com/jamespfennell/xz/lzma"
	"io"
	"os"
)

/*
#cgo LDFLAGS: -llzma -L/home/james/git/xz/
#include "xz/src/liblzma/api/lzma.h"
#include "example.c"

typedef struct Point {
    int x , y;
	lzma_stream strm;
} Point;

Point new_point() {
	Point a = {1, 2, LZMA_STREAM_INIT};
	return a;
}

// TODO: remove the void*
int thing(Point* p, void *next_in, size_t avail_in, void* next_out, size_t avail_out) {
	lzma_ret ret;
	p->strm.next_in = next_in;
	p->strm.avail_in = avail_in;
	p->strm.next_out = next_out;
	p->strm.avail_out = avail_out;
	ret = lzma_code(&(p->strm), LZMA_RUN);
	if (ret == LZMA_STREAM_END) {
		return 0;
	}
	return ret;
}
*/
// import "C"
type Writer struct {
	lzmaStream *lzma.Stream
}

func NewWriter() *Writer {
	return &Writer{
		lzmaStream: lzma.NewStream(),
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	w.lzmaStream.SetInput([]byte("my string to compress"))
	return 0, nil
}

func (w *Writer) Close() error {
	w.lzmaStream.Close()
	return nil
}

func main() {
	stream := lzma.NewStream()
	fmt.Println(stream)
	fmt.Println(lzma.EasyEncoder(stream, 6))
	stream.SetInput([]byte("my string to compress"))
	// return
	stream.SetOutputLen(100)
	fmt.Println("Run code: ", lzma.Code(stream, lzma.Run))
	stream.Output()

	fmt.Println("Finish code:", lzma.Code(stream, lzma.Finish))
	a := stream.Output()

	file, _ := os.Create("test.xz")
	defer file.Close()
	io.Copy(file, bytes.NewBuffer(a))
	file.Close()
	// Set the options using lzma_lzma_preset
	// What is lzma_easy_encoder?

	// Need to use the filter LZMA_FILTER_LZMA2


	stream.Close()
	/*
	// need something like coder_normal in coder.c line 631
	_ = NewWriter()
	point := CPoint{Point: C.new_point()}
	s := "my string to compress2"
	b := []byte(s)
	_ = b
	outB := make([]byte, 50)
	// rc := C.the_function(unsafe.Pointer(&b[0]), C.int(len(b)))
	fmt.Println(outB)
	fmt.Println(point)
	result := C.thing(&point.Point,
		unsafe.Pointer(&b[0]), C.ulong(len(b)),
		unsafe.Pointer(&outB[0]), C.ulong(len(outB)),
	)
	fmt.Println(outB)
	fmt.Println(point)
	fmt.Println("Result of lzma_code:", result)

	 */
}
