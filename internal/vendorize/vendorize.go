package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	_ "embed"
)

//go:embed shim_template.c.tmpl
var cTemplate string

const usage = `This script vendorizes C files in the upstream xz repository to this repo.
It MUST be run with the repo root as the current working directory.
Usage:

	go run internal/vendorize/vendorize.go [options]

Options:

`

const thisDirectory = `internal/vendorize`
const upstreamRoot = `internal/vendorize/upstream`
const goXzCommand = `internal/goxz/goxz.go`
const lzmaDirectory = `lzma`
const requiredFiles = `required_files.txt`

var vendorizeAllFiles bool
var skipBuldAndTests bool
var optimizeFiles bool

func init() {
	flag.BoolVar(&vendorizeAllFiles, "all", false,
		"vendorize all files in the upstream repo, not just those explicitly required")
	flag.BoolVar(&skipBuldAndTests, "skip-build", false,
		"skip build and tests after vendorizing")
	flag.BoolVar(&optimizeFiles, "optimize", false,
		"optimize the files by removing source files not needed for the tests to pass")
}

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	var err error
	var files []string
	if vendorizeAllFiles {
		files, err = listAllCFilesInUpstream()
		if err != nil {
			fmt.Println("Failed to list all C files in upstream:", err)
			os.Exit(1)
		}
	} else {
		required, err := os.ReadFile(filepath.Join(thisDirectory, requiredFiles))
		if err != nil {
			fmt.Printf("Failed to read required_files.txt: %s\n", err)
			os.Exit(1)
		}
		files = split(string(required))
	}

	if optimizeFiles {
		files = optimize(files)
	}

	if !vendorize(files, true) {
		os.Exit(1)
	}

	if optimizeFiles {
		fmt.Println("---REQUIRED FILE LIST FOLLOWS---")
		for _, file := range files {
			fmt.Println(file)
		}
	}
}

func optimize(files []string) []string {
	fmt.Println("Optimizing files")
	var required []string
	for i, file := range files {
		fmt.Printf("%d/%d\n", i+1, len(files))
		thisFiles := make([]string, i)
		copy(thisFiles, files[:i])
		thisFiles = append(thisFiles, files[i+1:]...)
		fmt.Printf("Testing the build without %s\n", file)
		if vendorize(thisFiles, false) {
			fmt.Println("!!Passed!!")
		} else {
			fmt.Println("  Failed")
			required = append(required, file)
		}
	}
	fmt.Printf("Required files: %d/%d\n", len(required), len(files))
	if len(required) < len(files) {
		fmt.Println("Performing another optimize round")
		return optimize(required)
	}
	return required
}

func vendorize(files []string, verbose bool) bool {
	removeVendorizedCFiles()
	for _, file := range files {
		if verbose {
			fmt.Println("Vendorizing", file)
		}
		if err := vendorizeCFile(file); err != nil {
			fmt.Printf("Failed to vendorize C file %s: %s", file, err)
			os.Exit(1)
		}
	}
	if skipBuldAndTests {
		return true
	}
	if verbose {
		fmt.Println("Running build and tests")
	}
	if !runBuildAndTests() {
		if verbose {
			fmt.Printf("Build and tests failed! Run `go build -a %s` to `go test ./...` to investigate.\n",
				goXzCommand)
		}
		return false
	}
	if verbose {
		fmt.Println("Success")
	}
	return true
}

func listAllCFilesInUpstream() ([]string, error) {
	var files []string
	roots := []string{"src/liblzma", "src/common"}
	for _, root := range roots {
		root = filepath.Join(upstreamRoot, root)
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			name := d.Name()
			if !strings.HasSuffix(name, ".c") && !strings.HasSuffix(name, ".h") {
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
			relPath := strings.TrimPrefix(path, upstreamRoot + "/")
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
	return exec.Command("go", "build", "-a", goXzCommand).Run() == nil &&
		exec.Command("go", "test", "./...").Run() == nil
}

func vendorizeCFile(upstreamFile string) error {
	b, err := os.ReadFile(filepath.Join(upstreamRoot, upstreamFile))
	if err != nil {
		return err
	}
	// TODO: validate that the file is in the public domain
	//  This doesn't cover header files but is probably good enough.
	//  Maybe we can transitively find header files that are included and audit them all
	vendorizedFile := filepath.Join(lzmaDirectory, upstreamFile)
	if err := os.MkdirAll(filepath.Dir(vendorizedFile), 0777); err != nil {
		return err
	}
	if err := os.WriteFile(vendorizedFile, b, 0666); err != nil {
		return err
	}
	if !strings.HasSuffix(upstreamFile, ".c") {
		return nil
	}
	shimFile := "shim__" + strings.ReplaceAll(upstreamFile, "/", "__")
	content := fmt.Sprintf(cTemplate, upstreamFile)
	return os.WriteFile(filepath.Join(lzmaDirectory, shimFile), []byte(content), 0666)
}

func removeVendorizedCFiles() {
	_ = os.RemoveAll(filepath.Join(lzmaDirectory, "src"))
	entries, _ := os.ReadDir(lzmaDirectory)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "shim__src__") && strings.HasSuffix(entry.Name(), ".c") {
			_ = os.Remove(filepath.Join(lzmaDirectory, entry.Name()))
		}
	}
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
