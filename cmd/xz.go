package main

import (
	"bytes"
	"github.com/jamespfennell/xz"
	"io"
	"os"
)

func main() {
	a := []byte("my string to compress part 2")
	file, _ := os.Create("test.xz")

	w := xz.NewWriter(file)
	io.Copy(w, bytes.NewReader(a))

	w.Close()

	file.Close()
	return
}
