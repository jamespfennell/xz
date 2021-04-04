package main

import (
	"fmt"
	"github.com/jamespfennell/xz"
	"io"
	"os"
)

func main() {
	inputFile, err := os.Open("test.tar")
	if err != nil {
		fmt.Println("Failed to open input file:", err)
		os.Exit(1)
	}
	outputFile, err := os.Create("test.xz")
	if err != nil {
		fmt.Println("Failed to create output file:", err)
		os.Exit(1)
	}
	w := xz.NewWriterLevel(outputFile, 9)
	if _, err := io.Copy(w, inputFile); err != nil {
		fmt.Println("Failed to compress data:", err)
	}
	if err := w.Close(); err != nil {
		fmt.Println("Failed to compress data:", err)
	}
	inputFile.Close()
	outputFile.Close()
	return
}
