# xz compression in Go [![GoDoc](https://godoc.org/github.com/jamespfennell/xz?status.png)](https://godoc.org/github.com/jamespfennell/xz)

This is a Go package for compressing and decompressing data in the xz format.
It works via a cgo wrapper around the lzma2 C library, which is part of the 
[XZ Utils project](https://tukaani.org/xz/).
The package does not require the lzma2 library to be installed, and on 
    any system can be used simply with:
    
    go get github.com/jamespfennell/xz@v0.1.2


## Security notice

On March 29 2024 a [backdoor in the upstream xz repository was discovered](https://news.ycombinator.com/item?id=39865810).
This backdoor was the result of a successful ~2 year effort by a malicious actor to
    take control of the upstream xz repo.

This Go package uses a vendored snapshot of the upstream xz repository from March 2020, 2 years before this attack started.
Specifically the snapshot was taken at
    [upstream commit `2327a461`](https://git.tukaani.org/?p=xz.git;a=commit;h=2327a461e1afce862c22269b80d3517801103c1b).
This commit was the release commit for
    [version 5.2.5 of xz](https://www.mail-archive.com/xz-devel@tukaani.org/msg00359.html).
This commit is [pinned in this repository using a Git submodule](https://github.com/jamespfennell/xz/tree/main/internal/vendorc).
The C files themselves were copied from the upstream repository into this repository.

Thus, this Go package is still safe to use if you assume
    (a) the backdoor from March 2024 is the first successful compromise of upstream and
    (b) that _this_ repository has not been compromised by me, the maintainer, jamespfennell@.
Your security posture will determine whether these assumptions are safe to make.
Honestly at this point it's probably just easier to use zstd instead of xz.

Finally, we note that the upstream attacker made about ~750 commits over 2 years to upstream
    before being discovered.
Given the state of upstream, **we will never be updating the upstream snapshot used in this repository**.


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

The full API can be browsed on [pkg.go.dev](https://pkg.go.dev/github.com/jamespfennell/xz).

## Build information

The lzma2 C code is automatically compiled by `go build`.
The C code is highly portable and regularly compiled for numerous architectures as part of the XZ Utils project,
    so it's unlikely `go build` will encounter any issues.
As part of the CI, the package is built and tested in the following environments:

| OS | Architecture | C compiler | Build status |
|---|---|---|---|
| Linux   | x86 | Clang, GCC | [![Linux x86 build status](https://github.com/jamespfennell/xz/actions/workflows/linux.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/linux.yml?query=branch%3Amain)
| Linux   | arm | Clang, GCC | [![Linux ARM build status](https://travis-ci.com/jamespfennell/xz.svg?branch=main)](https://travis-ci.com/github/jamespfennell/xz)
| MacOS   | x86 | Clang, GCC | [![MacOS build status](https://github.com/jamespfennell/xz/actions/workflows/macos.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/macos.yml?query=branch%3Amain)
| Windows | x86 | GCC | [![Windows build status](https://github.com/jamespfennell/xz/actions/workflows/windows.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/windows.yml?query=branch%3Amain)

As an alternative to compiling the C files during `go build`, the package can statically link to a precompiled
lzma2 library if it is already present on the system.
To do this, use the following build invocation:
 
    CGO_CFLAGS=-DGOXZ_SKIP_C_COMPILATION CGO_LDFLAGS=-llzma go build ...
    
The lzma2 library is present on MacOS by default.
On Debian it can be installed through the apt package `liblzma-dev`.
The CI builds and tests the package using this static linking approach for both MacOS and Linux on x86.

## Advanced usage: the lzma sub-package

The main xz package only exposes a limited subset of the lzma2 C library.
More advanced features of the library can be accessed using the lzma Go sub-package,
    which is where the actual cgo wrapping happens.
Currently, the sub-package only has enough wrapping code to facilitate the main xz use cases,
    however it is easy to extend it to access any method of the C library.

This sub-package is pretty "low level" and mostly just maps Go function calls/data structures directly to
    C function calls/data structures.
One consequence of this is that the sub-package does not have an idiomatic Go API. 
For example, instead of returning error types it returns integer statuses, as the C code does.
Using the sub-package also involves being familiar with the lzma2 API.
Ideally, additional features of the lzma2 library would be exposed through an idiomatic Go API in the xz package;
    we are open to pull requests in this direction.

In order to extend the lzma sub-package it may be necessary to tweak the cgo setup, which
    we now describe.
   
One of the rules of cgo is that all dependent C files must be in the same directory as the Go file that references
them.
For this kind of project, this usually necessitates copying files from the upstream project into the directory.
The approach here is a little different: 
    for each C source file needed to compile the package,
    we add a C shim file in the lzma directory that includes the source file using an `#include` directive.
There are two benefits to this.
First, by wrapping each `#include` in an `#ifndef GOXZ_SKIP_C_COMPILATION` conditional we can 
    use the `CGO_CFLAGS` environment variable to essentially skip C compilation entirely.
(The shim file will evaluate to an empty source file.)
This enables users to build the package by statically linking to a prebuilt system lzma2 library instead
    of compiling the library from scratch.
Second, it means that non-trivial source files are not duplicated in source control.
This is one of those things that's not really important but "feels good".

The shim files are generated automatically using the vendor script:

    go run internal/vendorc/vendorc.go

The script does not include every C file in the lzma2 library.
This is because the xz package does not use every lzma2 feature, and we can skip compiling features we don't need.
Doing so cuts the compilation time about in half.
The catch is that some features of the lzma2 library
    (like the x86 filter) won't work unless additional source files are vendored in.
In this case you can just pass the `--all` flag to the script and every possible C file will be included.

## The goxz command

The internal/goxz subdirectory contains an example command line tool `goxz` built on top of the xz package.
It only exposes limited features of the lzma2 library; the standard `xz` command from XZ-Utils is
much richer.

## Thanks

The lzma2 C library was mostly written by Lasse Collin.
The documentation for this library is really excellent, which made this package so much easier to write.

## License

All files in the lzma/src tree are copied from the upstream lzma2 repository and are in the public domain.
Alice In Wonderland is used for testing only, and is in the public domain.
All other files in the repository are released under the MIT license.
