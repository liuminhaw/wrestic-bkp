# Changelog

## [0.4.1] - 2024-05-10

### Changed

- Build release executable with `CGO` option disabled to increase portability ([#64](https://github.com/liuminhaw/wrestic-bkp/pull/64)) (Min-Haw, Liu)
- Add docs and config file into release package ([#65](https://github.com/liuminhaw/wrestic-bkp/pull/65)) (Min-Haw, Liu)

## [0.4.0] - 2024-01-01

### Changed

- Move origin bash script implementation to `scripts` directory ([#54](https://github.com/liuminhaw/wrestic-bkp/pull/54), [#56](https://github.com/liuminhaw/wrestic-bkp/pull/56))

### Added

- Golang implementation on local, s3, and sftp with init, backup, check, and snapshots actions ([#54](https://github.com/liuminhaw/wrestic-bkp/pull/54))

## [0.3.0] - 2023-10-24

### Changed

- Rename config template from `config.template` to `config.template.json` ([`2f49eb6`](https://github.com/liuminhaw/wrestic-bkp/pull/46/commits/2f49eb6))

### Added

- Add aws profile setting as a new way for s3 authentication ([#48](https://github.com/liuminhaw/wrestic-bkp/pull/48))
- Add `src_in_one` config to backup listed source paths together ([#46](https://github.com/liuminhaw/wrestic-bkp/pull/46))
- Add `tags` config for tagging and filtering backups ([#50](https://github.com/liuminhaw/wrestic-bkp/pull/50))

## [0.2.2] - 2023-10-08

### Fixed

- Fix mismatched version information ([#45](https://github.com/liuminhaw/wrestic-bkp/pull/45))

## [0.2.1] - 2023-09-05

### Fixed

- Fix error message when using mount action mount point with no paths setup ([#40](https://github.com/liuminhaw/wrestic-bkp/pull/40))

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

[0.4.1]: https://github.com/liuminhaw/wrestic-bkp/releases/tag/v0.4.1

[0.4.0]: https://github.com/liuminhaw/wrestic-bkp/releases/tag/v0.4.0

[0.3.0]: https://github.com/liuminhaw/restic-bkp/releases/tag/v0.3.0

[0.2.2]: https://github.com/liuminhaw/wrestic-bkp/releases/tag/v0.2.2

[0.2.1]: https://github.com/liuminhaw/wrestic-bkp/releases/tag/v0.2.1

[0.2.0]: https://github.com/liuminhaw/wrestic-bkp/releases/tag/v0.2.0

[0.1.2]: https://github.com/liuminhaw/restic-bkp/releases/tag/v0.1.2

[0.1.1]: https://github.com/liuminhaw/restic-bkp/releases/tag/v0.1.1

[0.1.0]: https://github.com/liuminhaw/restic-bkp/releases/tag/v0.1.0
