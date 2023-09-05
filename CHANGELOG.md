# Changelog

## [0.2.1] - 2023-09-05

### Fixed
- Fix error message when using mount action mount point with no paths setup([#40](https://github.com/liuminhaw/wrestic-bkp/pull/40)) 

## [0.2.0] - 2023-08-22

### Changed

- **Breaking:** Change config `src` to list format for multiple source paths backup ([#26](https://github.com/liuminhaw/wrestic-bkp/pull/26))
- **Breaking:** Change script name ([#34](https://github.com/liuminhaw/wrestic-bkp/pull/34))

### Added

- Add type s3 as new backup destination ([#29](https://github.com/liuminhaw/wrestic-bkp/pull/29))
- Add `default_password` config to mount point setting ([#31](https://github.com/liuminhaw/wrestic-bkp/pull/31))
- Add `paths` config to mount point setting ([#32](https://github.com/liuminhaw/wrestic-bkp/pull/32))
- Perform `restic check` after each backup is finished ([#28](https://github.com/liuminhaw/wrestic-bkp/pull/28))

### Fixed

- Fix script helper message format ([#36](https://github.com/liuminhaw/wrestic-bkp/pull/36))

## [0.1.2] - 2023-07-08

### Added

- Document changes with `CHANGELOG`

### Removed

- Remove `setup` script

### Fixed

- Fix file path conditioning ([#11](https://github.com/liuminhaw/restic-bkp/pull/11))

## [0.1.1] - 2022-06-06

### Changed

- Refacor structure with functions

### Added

- Add `mount` usage ([#6](https://github.com/liuminhaw/restic-bkp/pull/6))
- Add `--config` option to specify config file ([#3](https://github.com/liuminhaw/restic-bkp/pull/3))

## [0.1.0] - 2022-02-07

_:seedling: Initial release._

[0.2.1]: https://github.com/liuminhaw/wrestic-bkp/releases/tag/v0.2.1

[0.2.0]: https://github.com/liuminhaw/wrestic-bkp/releases/tag/v0.2.0

[0.1.2]: https://github.com/liuminhaw/restic-bkp/releases/tag/v0.1.2

[0.1.1]: https://github.com/liuminhaw/restic-bkp/releases/tag/v0.1.1

[0.1.0]: https://github.com/liuminhaw/restic-bkp/releases/tag/v0.1.0
