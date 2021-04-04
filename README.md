# xz compression in Go

This is a Go package for compressing and decompressing data in the xz format.
It works via a cgo wrapper around the C xz program (or, more precisely, the C lzma2 library).

To build the package the lzma C library needs to be installed, which
on Debian means installing the apt package `liblzma-dev`.
In the future we're hoping `go build` will also be able to compile the
library to remove this dependency, though it's not 100% clear this is 
possible.

The API follows the standard Go API for compression packages.

```
NewWriter(w io.Writer) *Writer
...
```

The package is currently (April 3, 2021) work in progress:

- Only the writer/compressor is implemented



## Notes

The Go package API will be the same as Go's `gzip` package
and [`zstd`](https://github.com/DataDog/zstd)

