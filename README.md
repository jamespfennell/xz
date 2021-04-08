# xz compression in Go

This is a Go package for compressing and decompressing data in the xz format.
It works via a cgo wrapper around the C lzma2 library which is part of the 
[XZ Utils project](https://tukaani.org/xz/).

The underlying C library is cross-platform.
As part of the CI, the Go package is built and tested for the following operating systems and (64-bit) CPUs:

| OS | Architecture | Build status |
|---|---|---|
| Linux   | x86 | [![Linux x86 build status](https://github.com/jamespfennell/xz/actions/workflows/linux.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/linux.yml?query=branch%3Amain)
| Linux   | arm | [![Linux ARM build status](https://travis-ci.com/jamespfennell/xz.svg?branch=main)](https://travis-ci.com/github/jamespfennell/xz)
| MacOS   | x86 | [![MacOS build status](https://github.com/jamespfennell/xz/actions/workflows/macos.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/macos.yml?query=branch%3Amain)
| Windows | x86 | [![Windows build status](https://github.com/jamespfennell/xz/actions/workflows/windows.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions/workflows/windows.yml?query=branch%3Amain)

## API

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

## Using other liblzma features

The Go xz package is an idiomatic Go API around the subpackage lzma which does the actual
wrapping of the C library.
The subpackage can be used for accessing many of the rich features and options of the lzma C
library which are not accessible though the main xz package.
However, the subpackage currently only has enough wrapping code to support the main xz use case.
If you want to access other C functions in the library via Go, feel free to add the wrappers and submit
a pull request.

## The goxz command

The cmd subdirectory contains an example command line tool `goxz` build on top of the xz package.
It's very limited and only exists for demonstration purposes; the standard `xz` command from XZ-Utils
is recommended for actual use.

## Thanks


## License

All C files from the upstream repository that are used to build the package are in the public domain.
Alice In Wonderland (which is used for testing) is in the public domain.
All other files in this repository are original are released under the MIT license.
