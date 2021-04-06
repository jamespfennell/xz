package xz

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os/exec"
	"testing"
)

//go:embed alice_in_wonderland.txt
var aliceInWonderland []byte

const smallString = "my string to compress"

type testCase struct {
	input       []byte
	compression int
}

func runOverAllTestCases(t *testing.T, fn func(*testing.T, testCase)) {
	strings := [][]byte{
		[]byte(smallString),
		aliceInWonderland,
	}
	for _, input := range strings {
		for compression := BestSpeed; compression <= BestCompression; compression++ {
			t.Run(
				fmt.Sprintf(
					"input %s... / compression level %d", string(input)[:10], compression),
				func(t *testing.T) {
					fn(t, testCase{
						input:       input,
						compression: compression,
					})
				},
			)
		}
	}
}

func TestWriterReaderRoundTrip(t *testing.T) {
	runOverAllTestCases(t, func(t *testing.T, tc testCase) {
		var output bytes.Buffer
		w := NewWriterLevel(&output, tc.compression)
		_, err := io.Copy(w, bytes.NewReader(tc.input))
		nilErrOrFail(t, err, "writing to the xz writer")
		nilErrOrFail(t, w.Close(), "closing the xz writer")

		r := NewReader(&output)
		reconstructedInput, err := io.ReadAll(r)
		nilErrOrFail(t, err, "reading from xz reader")
		expectBytesEqual(t, reconstructedInput, tc.input)
	})
}

func TestMisbehavingReaders(t *testing.T) {
	for i, newReader := range []func(io.Reader) io.Reader{
		func(r io.Reader) io.Reader {
			return &reluctantReader{
				reluctance: 5,
				r:          r,
			}
		},
		func(r io.Reader) io.Reader {
			return &slowReader{
				r: r,
			}
		},
		func(r io.Reader) io.Reader {
			return &drawnOutReader{
				r: r,
			}
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var output bytes.Buffer
			w := NewWriter(&output)
			_, err := io.Copy(w, bytes.NewReader(aliceInWonderland))
			nilErrOrFail(t, err, "writing to the xz writer")
			nilErrOrFail(t, w.Close(), "closing the xz writer")

			r := NewReader(newReader(&output))
			reconstructedInput, err := io.ReadAll(r)
			nilErrOrFail(t, err, "reading from xz reader")
			expectBytesEqual(t, reconstructedInput, aliceInWonderland)
		})
	}
}

func TestWriterAgreesWithShellCmd(t *testing.T) {
	runOverAllTestCases(t, func(t *testing.T, tc testCase) {
		var output bytes.Buffer
		w := NewWriterLevel(&output, tc.compression)
		_, err := io.Copy(w, bytes.NewReader(tc.input))
		nilErrOrFail(t, err, "writing to the xz writer")
		nilErrOrFail(t, w.Close(), "closing the xz writer")

		expectBytesEqual(t, xzShellCmdDecompress(t, output.Bytes()), tc.input)
	})

}

func TestReaderAgreesWithShellCmd(t *testing.T) {
	runOverAllTestCases(t, func(t *testing.T, tc testCase) {
		output := xzShellCmdCompress(t, tc.compression, tc.input)

		r := NewReader(bytes.NewReader(output))
		reconstructedInput, err := io.ReadAll(r)
		nilErrOrFail(t, err, "reading from xz reader")
		expectBytesEqual(t, reconstructedInput, tc.input)
	})

}

// xzShellCmdDecompress uses the xz command line program to
// decompress data. It is used to verify that the Go xz writer and the
// program give compatible results, and thus, under the assumption
// that the program is correct, that the Go xz writer is correct.
func xzShellCmdDecompress(t *testing.T, b []byte) []byte {
	cmd := exec.Command("xz", "-d")
	cmd.Stdin = bytes.NewReader(b)
	stdout, err := cmd.Output()
	nilErrOrFail(t, err, "decompressing using the xz shell command")
	return stdout
}

// xzShellCmdCompress uses the xz command line program to
// compress data. It is used to verify that the Go xz writer and the
// program give compatible results, and thus, under the assumption
// that the program is correct, that the Go xz writer is correct.
func xzShellCmdCompress(t *testing.T, level int, b []byte) []byte {
	cmd := exec.Command("xz", "-z", fmt.Sprintf("-%d", level))
	cmd.Stdin = bytes.NewReader(b)
	stdout, err := cmd.Output()
	nilErrOrFail(t, err, "compressing using the xz shell command")
	return stdout
}

func nilErrOrFail(t *testing.T, err error, action string) {
	if err != nil {
		t.Fatalf("Unexpected error while %s: %s", action, err)
	}
}

func expectBytesEqual(t *testing.T, a, b []byte) {
	if len(a) != len(b) {
		t.Errorf("Byte array input lengths not equal: %d != %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("Byte array differs at index %d: %d != %d", i, a[i], b[i])
		}
	}
}

// reluctantReader returns input from an underlying io.Reader once every reluctance reads. Other times,
// it returns 0, nil
type reluctantReader struct {
	reluctance int
	r          io.Reader

	// Implementation details
	pos int
}

func (z *reluctantReader) Read(p []byte) (int, error) {
	defer func() {
		// We only count non-trivial reads for reluctance purposes. Otherwise, if the consumer is also being conniving
		// and only issuing non-trivial reads every n reads we may get an infinite loop. Put another way, this
		// condition guarantees we'll make some progress every reluctance read.
		if len(p) > 0 {
			z.pos++
		}
	}()
	if z.pos == z.reluctance {
		z.pos = -1
		return z.r.Read(p)
	}
	return 0, nil
}

// slowReader returns a single byte of input from an underlying io.Reader on every Read.
type slowReader struct {
	r io.Reader

	// Implementation details
	buf     []byte
	lastErr error
}

func (z *slowReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for {
		if len(z.buf) > 0 {
			p[0] = z.buf[0]
			z.buf = z.buf[1:]
			return 1, nil
		}
		if z.lastErr != nil {
			return 0, z.lastErr
		}
		z.buf = make([]byte, len(p))
		var n int
		n, z.lastErr = z.r.Read(z.buf)
		if n == 0 && z.lastErr == nil {
			return 0, nil
		}
		z.buf = z.buf[:n]
	}
}

// drawnOutReader follows the non-recommended approach for marking io.EOF: it first returns the last bytes with
// a nil error, and on the next read returns 0, io.EOF.
type drawnOutReader struct {
	r io.Reader

	// Implementation details
	lastErr error
}

func (z *drawnOutReader) Read(p []byte) (int, error) {
	if z.lastErr != nil {
		return 0, z.lastErr
	}
	n, err := z.r.Read(p)
	if err != nil {
		z.lastErr = err
	}
	if err == io.EOF {
		err = nil
	}
	return n, err
}
