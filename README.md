# xz compression in Go

This is a Go package for compressing and decompressing data in the xz format.
It works via a cgo wrapper around the C lzma2 library which is part of the 
[XZ Utils project](https://tukaani.org/xz/).

## Usage

The API follows the standard Go API for compression packages.

```
const (
	BestSpeed          = 0
	BestCompression    = 9
	DefaultCompression = 6
)

// NewWriter creates a io.WriteCloser that xz-compresses input with the default 
// compression level and writes the output to w.
func NewWriter(w io.Writer) *Writer

// NewWriterLevel creates a io.WriteCloser that xz-compresses input with the prescribed 
// compression level and writes the output to w. The level should be between 
// BestSpeed and BestCompression inclusive; if it isn't, the level will be rounded
// up or down accordingly.
func NewWriterLevel(w io.Writer, level int) *Writer

// NewReader creates a new io.ReadCloser that reads xz-compressed input from r
// and returns decompressed output.
func NewReader(r io.Reader) *Reader
```

## Build information

The underlying lzma2 C library is cross-platform, and by default it is compiled during `go build`.
As part of the CI, the package is built and tested in the following environments:

| OS | Architecture | C compiler | Build status |
|---|---|---|---|
| Linux   | x86 | Clang, GCC | [![Linux x86 build status](https://github.com/jamespfennell/xz/actions/workflows/linux.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/linux.yml?query=branch%3Amain)
| Linux   | arm | Clang, GCC | [![Linux ARM build status](https://travis-ci.com/jamespfennell/xz.svg?branch=main)](https://travis-ci.com/github/jamespfennell/xz)
| MacOS   | x86 | Clang, GCC | [![MacOS build status](https://github.com/jamespfennell/xz/actions/workflows/macos.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/macos.yml?query=branch%3Amain)
| Windows | x86 | GCC | [![Windows build status](https://github.com/jamespfennell/xz/actions/workflows/windows.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/windows.yml?query=branch%3Amain)

Given the wide distribution of the XZ Utils software, and the high quality of the C code,
    we suspect `go build` will work out of the box on any 64-bit platform.

As an alternative to compiling the C files during `go build`, the package can statically link to a precompiled
lzma library if it is already present on the system.
To do this, use the following build invocation:
 
    CGO_CFLAGS=-DGOXZ_SKIP_C_COMPILATION CGO_LDFLAGS=-llzma go build ...
    
The lzma library is present on MacOS by default.
On Debian it can be installed through the apt package `liblzma-dev`.
The CI builds and tests the package using this static linking approach for both MacOS and Linux on x86.

## The lzma sub-package

This section is targeted at people who want to use features of the C lzma library that are not exposed
    through the xz package API described above.

The lzma C library is wrapped using cgo in the lzma Go sub-package.
This sub-package is pretty "low level" and mostly just maps Go function calls/data structures directly to
    C function calls/data structures.
One of the consequences of this is that it does not have an idiomatic Go API. 
For example, instead of returning error types it returns integer statuses, as the C code does.
The main xz package wraps the sub-package and provides the idiomatic Go API.

One of the rules of cgo is that all dependent C files must be in the same directory as the Go file that references
them.
For this kind of project, this usually necessitates copying files from the upstream project into the directory.
The approach here is a little different: 
    for each C source file needed to compile the package,
    we add a tiny C shim file in the lzma directory that includes the source file using an `#include` directive.
There are two benefits to this.
First, by wrapping each `#include` in an `#ifndef GOXZ_SKIP_C_COMPILATION` conditional we can 
    use the `CGO_CFLAGS` environment variable to essentially skip C compilation entirely.
(The shim file will evaluate to an empty source file.)
This enables users to build the package by statically linking to a prebuilt system lzma library instead
    of compiling the library from scratch.
Second, it means that non-trivial source files are not duplicated in source control.
This is one of those things that's not really important but "feels good".

The shim files are generated automatically using the vendorize script:

    go run lzma/vendorize/vendorize.go

The script does not include every C file in the lzma2 library.
This is because the xz package does not use every lzma2 feature, and we can skip compiling features we don't need.
Doing so cuts the compilation time about in half.
The catch is that some features of the lzma library
    (like CRC64 checking) won't work unless additional source files are vendored in.
In this case you can just pass the `--all` flag to the script and every possible C file will be included.

## The goxz command

The cmd subdirectory contains an example command line tool `goxz` built on top of the xz package.
It only exposes limited features of the lzma library; the standard `xz` command from XZ-Utils is
much richer.

## Thanks

The C lzma library was written by Lasse Collin.
The documentation for this library is really excellent, which made this package so much easier to write.

## License

All files in this repository, except Alice In Wonderland, are original and are released under the MIT license.
Alice In Wonderland is used for testing only, and is in the public domain.
Building the package involves pulling in C files from the upstream repository
    via the Git submodule at `lzma/upstream`.
These C files are all in the public domain.
