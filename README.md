# xz compression in Go

This is a Go package for compressing and decompressing data in the xz format.
It works via a cgo wrapper around the C lzma2 library which is part of the 
[XZ Utils project](https://tukaani.org/xz/).

To build the package the lzma C library needs to be installed:

| OS | Requirements | Build status |
|---|---|---|
| MacOS | Works out of the box | [![MacOS build status](https://github.com/jamespfennell/xz/actions/workflows/macos.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions?query=branch%3Amain+workflow%3AMacOS)
| Debian/Ubuntu | Requires the apt package `liblzma-dev` | [![Debian build status](https://github.com/jamespfennell/xz/actions/workflows/debian.yml/badge.svg?branch=main)](https://github.com/jamespfennell/xz/actions?query=branch%3Amain+workflow%3ADebian)

In the future we're hoping `go build` will also be able to compile the
library to remove this dependency, though it's not 100% clear this is 
possible.

 
## API

The API follows the standard Go API for compression packages.

```
NewWriter(w io.Writer) *Writer
...
```

The package is currently (April 3, 2021) work in progress.

## Using other liblzma features

The Go xz package is an idiomatic Go API around the subpackage lzma which does the actual
wrapping of the C library.
The subpackage can be used for accessing many of the rich features and options of the lzma C
library which are not accessible though the main xz package.
However, the subpackage currently only has enough wrapping code to support the main xz use case.
If you want to access other C functions in the library via Go, feel free to add the wrappers and submit
a pull request.

## Thanks



## License

All files in this repository that are original (i.e., not from the upstream xz repository)
are released under the MIT license.
All C files from the upstream repository that are used to build the package are in the public domain.
