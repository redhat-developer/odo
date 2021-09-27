# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.0.0] - 2020-08-03

### Added

- Add support for all os that go support
  - aix
  - hurd
  - js
  - nacl
  - plan9
  - zos
- windows: Add env support

### Changed

- windows: Read windows registry instead of OLE

## [v0.3.0] - 2020-06-03

### Added

- Add FreeBSD/OpenBSD support (#12)

### Changed

- unix: Detect via locale.conf instead of locale command (#14)

## [v0.2.0] - 2020-04-21

### Added

- Add system v support (#8)
- Add full windows LCID support (#10)

## v0.1.0 - 2020-02-20

### Added

- Support Linux, macOS X and Windows platforms
- Support Detect and DetectAll

[v1.0.0]: https://github.com/Xuanwo/go-locale/compare/v0.3.0...v1.0.0
[v0.3.0]: https://github.com/Xuanwo/go-locale/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/Xuanwo/go-locale/compare/v0.1.0...v0.2.0
