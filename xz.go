package xz

import "C"
import (
	"fmt"
	"github.com/jamespfennell/xz/lzma"
	"io"
)

const (
	BestSpeed          = 0
	BestCompression    = 9
	DefaultCompression = 6
)

type LzmaError struct {
	result lzma.Return
}

func (err LzmaError) Error() string {
	return fmt.Sprintf("lzma library returned a %s error", err.result)
}

type Writer struct {
	lzmaStream *lzma.Stream
	w          io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return NewWriterLevel(w, DefaultCompression)
}

func NewWriterLevel(w io.Writer, level int) *Writer {
	if level < BestSpeed {
		fmt.Printf("xz library: unexpected negative compression level %d; using level 0", level)
		level = BestSpeed
	}
	if level > BestCompression {
		fmt.Printf("xz library: unexpected compression level %d bigger than 9; using level 9", level)
		level = BestCompression
	}
	s := lzma.NewStream()
	// TODO: leave a comment if there's an error
	lzma.EasyEncoder(s, level)
	// TODO: configurable? See what zstd and gzip do here
	//  Just benchmark on some huge files and see what happens
	s.SetOutputLen(500)
	return &Writer{
		lzmaStream: s,
		w:          w,
	}
}

func (z *Writer) Write(p []byte) (int, error) {
	z.lzmaStream.SetInput(p)
	return z.consumeInput()
}

func (z *Writer) Close() error {
	if _, err := z.consumeInput(); err != nil {
		return err
	}
	for {
		if z.lzmaStream.AvailOut() == 0 {
			if _, err := z.w.Write(z.lzmaStream.Output()); err != nil {
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
