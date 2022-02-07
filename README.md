# restic-bkp
Backup files with custom configuration using restic

## Pre-requisite
restic need to be installed before using this script  
**reference:** [Installation](https://restic.readthedocs.io/en/latest/020_installation.html)

## Version
### v0.1.0
- Use new config format: json
    - Support multiple sources and destinations configuration
- New configuration setting: forget policy
- Moving from [some-script](https://github.com/liuminhaw/some-scripts) repositroy to separate independent repository

**More versions information:** [Versions Documentation](https://mewing-pisces-b2c.notion.site/restic-bkp-1b852fb0215a4e1d8df00081f3050e1b) 

## Setup
#### Configuration
Set backup information (source, destination, password, exclusion, sna)
- password file
- exclusive file
- source
- destination
- snapshot policy

#### Usage
```
Usage: restic-bkp.sh [--help] [--version] [--type=local|sftp] backup|init|snapshots
    --help                      Display this help message and exit
    --type=[local|sftp]         
    --type [local|sftp]         Specify backup destination type: (local, sftp)
                                Default type: local
    --version                   Show version information
    action                      Command to execute: (backup, init, snapshots)
                                backup: create new backup snapshot
                                init: prepare backup destination directory
                                snapshots: list previous snapshots
```

#### Exclusion
The `exclude_file` key in config file (default: excludes.txt)<br>
Create the file with listed files and directories for exclusion
```
List
excluded 
files
and
diretories
here
``` 

#### Password
The `password_file` key in config file (default: .restic.pass)<br>
Password is needed when restic process
```
defaultPasswordToEncryptResticSnapshot
```
