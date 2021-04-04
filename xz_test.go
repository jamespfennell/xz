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
