package xz

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"testing"
)

func TestWriterAgreesWithShellCmd(t *testing.T) {
	for compression := BestSpeed; compression <= BestCompression; compression++ {
		t.Run(fmt.Sprintf("compression level %d", compression), func(t *testing.T) {
			input := []byte("my string to compress")
			var output bytes.Buffer
			w := NewWriterLevel(&output, compression)
			_, err := io.Copy(w, bytes.NewReader(input))
			nilErrOrFail(t, err, "writing to the xz writer")
			nilErrOrFail(t, w.Close(), "closing the xz writer")

			expectBytesEqual(t, xzShellCmdDecompress(t, output.Bytes()), input)
		})
	}
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
