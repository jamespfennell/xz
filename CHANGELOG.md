# Changelog

## [Unreleased]

## [0.1.2] - 2021-04-21

- lzma2 version: 5.2.5.

- Fixed the vendoring process. It turns out that Go modules
    do not support Git submodules and so upstream C files
    still need to be copied into the repo.
  
## [0.1.1] - 2021-04-11

- lzma2 version: 5.2.5.

- Updated Github documentation and added
    documentation on pkg.dev.go.

## [0.1.0] - 2021-04-11

- lzma2 version: 5.2.5.

- Support for compressing in the xz format with the
    10 different compression levels.
    
- Support for decompressing xz streams.

[unreleased]: https://github.com/jamespfennell/xz/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/jamespfennell/xz/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/jamespfennell/xz/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/jamespfennell/xz/compare/d164ed5c1f3e59bdb117c87312078543522ab99a...v0.1.0
