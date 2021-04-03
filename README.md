# xz compression in Go

This Go package enables compressing and decompressing data in the xz format.
It works via a cgo wrapper around the C xz program (or, more precisely, the C lzma2 library).

The API follows the standard Go API for compression packages.

```
NewWriter(w io.Writer) *Writer
...
```

The package is currently (April 3, 2021) work in progress:

- Only the writer/compressor is implemented
- Buffering is really inefficient
- Memory leaks are everywhere  
- The lzma library needs to be manually compiled and the `.a` file put in
    the directory `/home/james/git/xz`...hopefully we can fix this and
    have `go build` compile the C files. Otherwise, using the package will
    require the system have the lzma library already installed.



## Notes

The Go package API will be the same as Go's `gzip` package
and [`zstd`](https://github.com/DataDog/zstd)

