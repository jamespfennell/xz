# xz
[WIP] Go bindings for xz compression 

## Notes

We can use the `xz` command line tool's usage of the `lzma` library/API as a template: [link to function](https://git.tukaani.org/?p=xz.git;a=blob;f=src/xz/coder.c;h=85f954393d8bf0df73eeaf90669f65cc4705ef4e;hb=e7da44d5151e21f153925781ad29334ae0786101#l629).

The Go package API will be the same as Go's `gzip` package and [`zstd`](https://github.com/DataDog/zstd):

```
const (
    BestSpeed          = 0
    BestCompression    = 9
    DefaultCompression = 6
)

type Writer struct {
    ...
}

func NewWriter(w io.Writer) *Writer {
    return NewWriterLevel(w, DefaultCompression)
}

func NewWriterLevel(w io.Writer, level int) *Writer {
    ...
}
```
