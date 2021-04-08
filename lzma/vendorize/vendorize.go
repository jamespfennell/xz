package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const template = `// This file was auto-generated by the vendorize.go script.
//
// This file is released under the MIT license but every source file it includes
// (directly and transitively) is in the public domain.
//
// Generated when the upstream repository was at commit %s.

#ifndef GOXZ_SKIP_C_COMPILATION
#include "upstream/%s"
#endif
`

func main() {
	gitHash, err := upstreamGitHash()
	if err != nil {
		fmt.Println("Failed to determine the Git hash of the upstream repo:", err)
		os.Exit(1)
	}
	entries, _ := os.ReadDir("lzma")
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "upstream__src__") && strings.HasSuffix(entry.Name(), ".c") {
			os.Remove(filepath.Join("lzma", entry.Name()))
		}
	}
	var srcs []string
	// TODO: all of the warnings are from the range_decoder.h, maybe we can remove it because the Makefile says
	//  it's only for LZMA1
	roots := []string{"lzma/upstream/src/liblzma", "lzma/upstream/src/common"}
	for _, root := range roots {
		filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if !strings.HasSuffix(info.Name(), ".c") {
				return nil
			}
			if strings.HasSuffix(info.Name(), "tablegen.c") {
				return nil
			}
			// TODO: try not to include every file...!
			//  Probably we can just run the script 88 times skipping each source file, and see which of the 88
			//  runs succeed.
			if info.Name() == "crc32_small.c" {
				return nil
			}
			if info.Name() == "crc64_small.c" {
				return nil
			}
			if info.Name() == "stream_encoder_mt.c" {
				return nil
			}
			relPath := strings.TrimPrefix(path, "lzma/upstream/")
			println(relPath)
			srcs = append(srcs, relPath)
			return nil
		})
	}
	for _, src := range srcs {
		if _, err := os.Stat(filepath.Join("lzma", "upstream", src)); err != nil {
			fmt.Println("Cannot find", src)
			os.Exit(1)
		}
		vFileName := strings.ReplaceAll("upstream/"+src, "/", "__")
		w, _ := os.Create(filepath.Join("lzma", vFileName))
		fmt.Fprintf(w, template, gitHash, src)
		w.Close()
	}
}

func upstreamGitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = "lzma/upstream"
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	err := cmd.Run()
	return strings.TrimSpace(buffer.String()), err
}