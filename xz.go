// Package xz implements xz compression and decompression.
package xz

import (
	"bytes"
	"fmt"
	"github.com/jamespfennell/xz/lzma"
	"io"
)

const (
	BestSpeed          = 0
	BestCompression    = 9
	DefaultCompression = 6
)

// LzmaError may be returned if the underlying lzma library returns an error code during compression or decompression.
// Receiving this error indicates a bug in the xz package, and a bug report would be appreciated.
type LzmaError struct {
	result lzma.Return
}

func (err LzmaError) Error() string {
	return fmt.Sprintf(
		"lzma library returned a %s error. This indicates a bug in the Go xz package", err.result)
}

// Writer is an io.WriteCloser that xz-compresses its input and writes it to an underlying io.Writer
type Writer struct {
	lzmaStream *lzma.Stream
	w          io.Writer
	// TODO: lastErr
}

// NewWriter creates a Writer that compresses with the default compression level of DefaultCompression and writes the
// output to w.
func NewWriter(w io.Writer) *Writer {
	return NewWriterLevel(w, DefaultCompression)
}

// NewWriterLevel creates a Writer that compresses with the prescribed compression level and writes the output to w.
// The level should be between BestSpeed and BestCompression inclusive; if it isn't, the level will be rounded up
// or down accordingly.
func NewWriterLevel(w io.Writer, level int) *Writer {
	if level < BestSpeed {
		fmt.Printf("xz library: unexpected negative compression level %d; using level 0\n", level)
		level = BestSpeed
	}
	if level > BestCompression {
		fmt.Printf("xz library: unexpected compression level %d bigger than 9; using level 9\n", level)
		level = BestCompression
	}
	s := lzma.NewStream()
	if ret := lzma.EasyEncoder(s, level); ret != lzma.Ok {
		fmt.Printf("xz library: unexpected result from encoder initialization: %s\n", ret)
	}
	return &Writer{
		lzmaStream: s,
		w:          w,
	}
}

// Write accepts p for compression.
//
// Because of internal buffering and the mechanics of xz, the compressed version of p is not guaranteed to have been
// written to the underlying io.Writer when the function returns.
func (z *Writer) Write(p []byte) (int, error) {
	z.lzmaStream.SetInput(p)
	return z.consumeInput()
}

// Close finishes processing any input that has yet to be compressed, writes all remaining output to the underlying
// io.Writer, and frees memory resources associated to the Writer.
func (z *Writer) Close() error {
	if _, err := z.consumeInput(); err != nil {
		// TODO: this is a bug. We need to close resources in this case
		return err
	}
	for {
		if z.lzmaStream.AvailOut() == 0 {
			if _, err := z.w.Write(z.lzmaStream.Output()); err != nil {
				// TODO: this is a bug. We need to close resources in this case
				return err
			}
		}
		result := lzma.Code(z.lzmaStream, lzma.Finish)
		if result == lzma.StreamEnd {
			break
		}
		if result != lzma.Ok {
			return LzmaError{result: result}
		}
	}
	if _, err := z.w.Write(z.lzmaStream.Output()); err != nil {
		// TODO: this is a bug. We need to close resources in this case
		return err
	}
	z.lzmaStream.Close()
	return nil
}

func (z *Writer) consumeInput() (int, error) {
	start := z.lzmaStream.TotalIn()
	var err error
	for {
		if z.lzmaStream.AvailIn() == 0 {
			break
		}
		if z.lzmaStream.AvailOut() == 0 {
			if _, err = z.w.Write(z.lzmaStream.Output()); err != nil {
				break
			}
		}
		result := lzma.Code(z.lzmaStream, lzma.Run)
		if result != lzma.Ok {
			err = LzmaError{result: result}
			break
		}
	}
	return z.lzmaStream.TotalIn() - start, err
}

// Reader is an io.ReadCloser that xz-decompresses from an underlying io.Reader.
type Reader struct {
	lzmaStream    *lzma.Stream
	r             io.Reader
	buf           bytes.Buffer
	inputFinished bool
	lastErr       error
}

// NewReader creates a new Reader that reads compressed input from r.
func NewReader(r io.Reader) *Reader {
	s := lzma.NewStream()
	if ret := lzma.StreamDecoder(s); ret != lzma.Ok {
		fmt.Printf("xz library: unexpected result from decoder initialization: %s\n", ret)
	}
	return &Reader{
		lzmaStream: s,
		r:          r,
	}
}

// Read decompresses output from the underlying io.Reader and returns up to len(p) uncompressed bytes.
func (z *Reader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if z.lastErr != nil {
		return 0, z.lastErr
	}
	if z.buf.Len() < len(p) {
		// We have no idea how much data to request from the underlying io.Reader, so just cargo cult from the caller...
		z.lastErr = z.populateBuffer(len(p))
		if z.lastErr != nil {
			return 0, z.lastErr
		}
	}
	var n int
	n, z.lastErr = z.buf.Read(p)
	return n, z.lastErr
}

func (z *Reader) populateBuffer(sizeHint int) error {
	if z.inputFinished {
		return nil
	}

	q := make([]byte, sizeHint)
	m, err := z.r.Read(q)
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF {
		z.inputFinished = true
	}
	z.lzmaStream.SetInput(q[:m])

	var outputs [][]byte
	action := lzma.Run
	for {
		// When decoding with lzma.Run, lzma requires the input buffer be non-empty. So if it is empty, either return
		// or transition to lzma.Finish.
		if action == lzma.Run && z.lzmaStream.AvailIn() == 0 {
			if !z.inputFinished {
				break
			}
			action = lzma.Finish
		}
		result := lzma.Code(z.lzmaStream, action)
		// The output buffer is not necessarily full, but because we're decompressing it often is so for simplicity
		// just copy and clear it.
		outputs = append(outputs, z.lzmaStream.Output())
		if result == lzma.StreamEnd && action == lzma.Finish {
			break
		}
		if result.IsErr() {
			return LzmaError{result: result}
		}
	}

	var totalNewLen int
	for _, output := range outputs {
		totalNewLen += len(output)
	}
	z.buf.Grow(totalNewLen)
	for _, output := range outputs {
		// the error on this Write is always nil
		z.buf.Write(output)
	}
	return nil
}

// Close released resources associated to this Reader.
func (z *Reader) Close() error {
	z.lzmaStream.Close()
	return nil
}
