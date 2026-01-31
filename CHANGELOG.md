# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0] - 2026-01-31
### :sparkles: New Features
- [`108be6a`](https://github.com/rmhubbert/tuifade/commit/108be6a838d527ee72e40861771ed2f73165330a) - add Fader struct to enable session long caching *(commit by [@rmhubbert](https://github.com/rmhubbert))*


## [v0.1.0] - 2026-01-31
### :sparkles: New Features
- [`ec6d256`](https://github.com/rmhubbert/tuifade/commit/ec6d25647d02e44c695c43b51def2846a93bebcb) - add interpolation clamp and early return for no fade interpolation value *(commit by [@rmhubbert](https://github.com/rmhubbert))*

### :bug: Bug Fixes
- [`ddc8ce1`](https://github.com/rmhubbert/tuifade/commit/ddc8ce1c879e5d4e502c3d5afc6387e468058cd6) - incorrect colour assignment *(commit by [@rmhubbert](https://github.com/rmhubbert))*

### :recycle: Refactors
- [`68ab076`](https://github.com/rmhubbert/tuifade/commit/68ab07642f9c40f56e54a915c6022f6fc590e033) - add interpolation cache *(commit by [@rmhubbert](https://github.com/rmhubbert))*
- [`4fab202`](https://github.com/rmhubbert/tuifade/commit/4fab202e840212bef8d55e25b66c4046cf058da8) - remove caching & create local version of segments array *(commit by [@rmhubbert](https://github.com/rmhubbert))*
- [`d0cac0d`](https://github.com/rmhubbert/tuifade/commit/d0cac0da5d80c1e7a6ac7f0cf37fe8cc1833796a) - make interpolation structs private *(commit by [@rmhubbert](https://github.com/rmhubbert))*

### :white_check_mark: Tests
- [`d31f304`](https://github.com/rmhubbert/tuifade/commit/d31f30491a6e20bf1cb3fed6f76eda5e661527ec) - remove obsolete tests *(commit by [@rmhubbert](https://github.com/rmhubbert))*
- [`c92fab8`](https://github.com/rmhubbert/tuifade/commit/c92fab8dcc6dcd66f57c6e0011b27531ae97175c) - unit tests *(commit by [@rmhubbert](https://github.com/rmhubbert))*
- [`31df035`](https://github.com/rmhubbert/tuifade/commit/31df03580945a157a0363edecab52aa1fd457c44) - remove outdated tests *(commit by [@rmhubbert](https://github.com/rmhubbert))*


## [v0.0.7] - 2026-01-18
### :recycle: Refactors
- [`b53206c`](https://github.com/rmhubbert/tuifade/commit/b53206c03d864a0310d72d4a3f4f12d94713b822) - only output background ansi codes when the background is different from the default *(commit by [@rmhubbert](https://github.com/rmhubbert))*

### :white_check_mark: Tests
- [`da5488a`](https://github.com/rmhubbert/tuifade/commit/da5488af59a34543b5bf8ed39ae338389f71835a) - linting *(commit by [@rmhubbert](https://github.com/rmhubbert))*


## [v0.0.6] - 2026-01-17
### :zap: Performance Improvements
- [`3d47214`](https://github.com/rmhubbert/tuifade/commit/3d472146730fbdfd0ecbe0a00bbaf662b1b0cf09) - remove cache warm up and switch to ansi-parse String method *(commit by [@rmhubbert](https://github.com/rmhubbert))*

### :white_check_mark: Tests
- [`654cee1`](https://github.com/rmhubbert/tuifade/commit/654cee1ab54d186ee8d3041056d623a26d3a8e63) - add individual benchmark tests *(commit by [@rmhubbert](https://github.com/rmhubbert))*


## [v0.0.5] - 2026-01-17
### :zap: Performance Improvements
- [`b0d2780`](https://github.com/rmhubbert/tuifade/commit/b0d27804ad6bbf1e3821945748ed1070c847c470) - add colour cache to stop repeated conversion *(commit by [@rmhubbert](https://github.com/rmhubbert))*


## [v0.0.3] - 2026-01-17
### :recycle: Refactors
- [`5db2821`](https://github.com/rmhubbert/tuilum/commit/5db28214833c9ec10fd5f6b6917c0d46472afc9d) - use go-ansi-parser to manipulate input *(commit by [@rmhubbert](https://github.com/rmhubbert))*

### :white_check_mark: Tests
- [`2e2340e`](https://github.com/rmhubbert/tuilum/commit/2e2340e55437bfab720043af9b249808d116705e) - add unit and integration tests *(commit by [@rmhubbert](https://github.com/rmhubbert))*

### :wrench: Chores
- [`b75c401`](https://github.com/rmhubbert/tuilum/commit/b75c40151617be1cd489b638b774252bfaa97b58) - remove test cli *(commit by [@rmhubbert](https://github.com/rmhubbert))*


## [v0.0.2] - 2026-01-05
### :wrench: Chores
- [`0d8c0fb`](https://github.com/rmhubbert/tuilum/commit/0d8c0fb8535fd1a27ae998a83c792dd150aaafdd) - add basic scaffolding *(commit by [@rmhubbert](https://github.com/rmhubbert))*
- [`6432826`](https://github.com/rmhubbert/tuilum/commit/64328268e48ad166674596834376d7a15672c26c) - wip *(commit by [@rmhubbert](https://github.com/rmhubbert))*

[v0.0.2]: https://github.com/rmhubbert/tuilum/compare/v0.0.1...v0.0.2
[v0.0.3]: https://github.com/rmhubbert/tuilum/compare/v0.0.2...v0.0.3
[v0.0.5]: https://github.com/rmhubbert/tuifade/compare/v0.0.4...v0.0.5
[v0.0.6]: https://github.com/rmhubbert/tuifade/compare/v0.0.5...v0.0.6
[v0.0.7]: https://github.com/rmhubbert/tuifade/compare/v0.0.6...v0.0.7
[v0.1.0]: https://github.com/rmhubbert/tuifade/compare/v0.0.8...v0.1.0
[v0.2.0]: https://github.com/rmhubbert/tuifade/compare/v0.1.1...v0.2.0
