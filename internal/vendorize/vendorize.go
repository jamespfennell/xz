package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const cTemplate = `// This file was auto-generated by the vendorize.go script.
//
// This file is released under the MIT license but every source file it includes
// (directly and transitively) is in the public domain.
//
// Generated when the upstream repository was at commit %s.

#ifndef GOXZ_SKIP_C_COMPILATION
#include "upstream/%s"
#endif
`

const requiredUpstreamFiles = `
src/liblzma/check/check.c
src/liblzma/check/crc32_fast.c
src/liblzma/check/crc32_table.c
src/liblzma/check/crc64_fast.c
src/liblzma/check/crc64_table.c

src/liblzma/common/block_decoder.c
src/liblzma/common/block_encoder.c
src/liblzma/common/block_header_decoder.c
src/liblzma/common/block_header_encoder.c
src/liblzma/common/block_util.c
src/liblzma/common/common.c
src/liblzma/common/easy_encoder.c
src/liblzma/common/easy_preset.c
src/liblzma/common/filter_common.c
src/liblzma/common/filter_decoder.c
src/liblzma/common/filter_encoder.c
src/liblzma/common/filter_flags_decoder.c
src/liblzma/common/filter_flags_encoder.c
src/liblzma/common/index.c
src/liblzma/common/index_encoder.c
src/liblzma/common/index_hash.c
src/liblzma/common/stream_decoder.c
src/liblzma/common/stream_encoder.c
src/liblzma/common/stream_flags_common.c
src/liblzma/common/stream_flags_decoder.c
src/liblzma/common/stream_flags_encoder.c
src/liblzma/common/vli_decoder.c
src/liblzma/common/vli_encoder.c
src/liblzma/common/vli_size.c

src/liblzma/lz/lz_decoder.c
src/liblzma/lz/lz_encoder.c
src/liblzma/lz/lz_encoder_mf.c

src/liblzma/lzma/fastpos_table.c
src/liblzma/lzma/lzma2_decoder.c
src/liblzma/lzma/lzma2_encoder.c
src/liblzma/lzma/lzma_decoder.c
src/liblzma/lzma/lzma_encoder.c
src/liblzma/lzma/lzma_encoder_optimum_fast.c
src/liblzma/lzma/lzma_encoder_optimum_normal.c
src/liblzma/lzma/lzma_encoder_presets.c

src/liblzma/rangecoder/price_table.c
`

const usage = `This script vendorizes C files in the upstream xz repository to this repo.
It MUST be run with the repo root as the current working directory.
Usage:

	go run lzma/vendorize/vendorize.go [options]

Options:

`

var vendorizeAllFiles bool

func init() {
	// TODO: implement an optimize flag
	flag.BoolVar(&vendorizeAllFiles, "all", false,
		"vendorize all files in the upstream repo, not just those explicitly required")
}

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	gitHash, err := upstreamGitHash()
	if err != nil {
		fmt.Println("Failed to determine the Git hash of the upstream repo:", err)
		os.Exit(1)
	}

	var files []string
	if vendorizeAllFiles {
		files, err = listAllCFilesInUpstream()
		if err != nil {
			fmt.Println("Failed to list all C files in upstream:", err)
			os.Exit(1)
		}
	} else {
		files = split(requiredUpstreamFiles)
	}

	removeVendorizedCFiles()
	for _, file := range files {
		fmt.Println("Vendorizing", file)
		// TODO: validate that the file is in the public domain
		//  This doesn't cover header files but is probably good enough.
		//  Maybe we can transitively find header files that are included and audit them all
		if err := vendorizeCFile(file, gitHash); err != nil {
			fmt.Printf("Failed to vendorize C file %s: %s", file, err)
			os.Exit(1)
		}
	}
	fmt.Println("Running build and tests")
	if !runBuildAndTests() {
		fmt.Println("Build and tests failed! Run `go build -a goxz/goxz.go` to `go test ./...` to investigate.")
		os.Exit(1)
	}
	fmt.Println("Success")
}

func listAllCFilesInUpstream() ([]string, error) {
	var files []string
	roots := []string{"lzma/upstream/src/liblzma", "lzma/upstream/src/common"}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			name := d.Name()
			if !strings.HasSuffix(name, ".c") {
				return nil
			}
			// The *tablegen.c files are helper files with main functions. Including them means the library can't
			// compile
			if strings.HasSuffix(name, "tablegen.c") {
				return nil
			}
			// The next two files are only for compilers without the standard bool C library. We don't support these.
			if name == "crc32_small.c" {
				return nil
			}
			if name == "crc64_small.c" {
				return nil
			}
			if name == "stream_encoder_mt.c" {
				return nil
			}
			relPath := strings.TrimPrefix(path, "lzma/upstream/")
			files = append(files, relPath)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func runBuildAndTests() bool {
	return exec.Command("go", "build", "-a", "goxz/goxz.go").Run() == nil &&
		exec.Command("go", "test", "./...").Run() == nil
}

func vendorizeCFile(upstreamFile string, gitHash string) error {
	if _, err := os.Stat(filepath.Join("lzma", "upstream", upstreamFile)); err != nil {
		return err
	}
	vendorizedFile := strings.ReplaceAll("upstream/"+upstreamFile, "/", "__")
	w, err := os.Create(filepath.Join("lzma", vendorizedFile))
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, cTemplate, gitHash, upstreamFile); err != nil {
		return err
	}
	return w.Close()
}

func removeVendorizedCFiles() {
	entries, _ := os.ReadDir("lzma")
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "upstream__src__") && strings.HasSuffix(entry.Name(), ".c") {
			_ = os.Remove(filepath.Join("lzma", entry.Name()))
		}
	}
}

// upstreamGitHash determines the current hash of the upstream repo
func upstreamGitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = "lzma/upstream"
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	err := cmd.Run()
	return strings.TrimSpace(buffer.String()), err
}

// split splits the string on newlines, removes empty lines and lines with a // prefix, and returns the trimmed strings
func split(s string) []string {
	var r []string
	for _, l := range strings.Split(strings.TrimSpace(s), "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "//") {
			continue
		}
		r = append(r, l)
	}
	return r
}