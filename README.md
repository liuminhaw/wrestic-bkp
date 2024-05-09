# Project Name
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Description
Wrapper of [restic](https://github.com/restic/restic) with configurations

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Scripts implementation](#scripts-implementation)
- [License](#license)
- [Contact](#contact)

## Installation
### Pre-built binary
Download prebuilt binary from project [releases](https://github.com/liuminhaw/wrestic-bkp/releases)
### Build from source
1. Clone or download source code
1. Run `go build` inside cloned directory

## Usage
`BackupName`: name defined in config file
### Restic run
- Init repository
  ```bash
  ./wrestic-bkp run init BackupName [flags]
  ```
- Backup 
  ```bash
  ./wrestic-bkp run backup BackupName [flags]
  ```
- Show snapshots
  ```bash
  ./wrestic-bkp run snapshots BackupName [flags]
  ```
- Check backups integrity and consistency
  ```bash
  ./wrestic-bkp run check BackupName [flags]  
  ```
### Config 
Show configuration file content
```bash
./wrestic-bkp config show [BackupName] [flags]
```

## Scripts implementation
Scripts implementation of `wrestic-bkp` before migrating using Golang: [scripts](./scritps)

## License
This project is licensed under the [MIT License](LICENSE).

## Contact
- GitHub: [liuminhaw](https://github.com/liuminhaw)
