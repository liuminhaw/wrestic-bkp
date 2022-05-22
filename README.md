# restic-bkp
Backup files with custom configuration using restic

## Pre-requisite
1. restic need to be installed before executing the script  
**reference:** [Installation](https://restic.readthedocs.io/en/latest/020_installation.html)
1. jq need to be installed before executing the script
**reference:** [Installation](https://stedolan.github.io/jq/download/)

## Version
### v0.1.1
- Add cli --config option to specify which file to read in as configuration setting 
    - Default config file is `config.json` if `--config` option is not used

**More versions information:** [Versions Documentation](https://mewing-pisces-b2c.notion.site/restic-bkp-1b852fb0215a4e1d8df00081f3050e1b) 

## Usage
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

## Setup
### Configuration
Set backup information (source, destination, password, exclusion, snapshot policy)
- [password file](#password) - Filename where encryption password is stored 
    - Set to `.restic.pass` as default, modify its content for custom password
- [exclusive file](#exclusion) - Files or direcotries to skip while backing up, one file / directory per line
    - Set to `excludes.txt` as default, modify its content for custom settings
- [type](#type) - Current support types: `local` `sftp`
    - source - Backup source path
    - destination - Backup destination path
    - host (`sftp` type only) - host name set in ssh config file
- [snapshot policy](#snapshot-policy) - Rules to indicate which snapshots to keep / remove

Configuration file can be copied and edit from `config.template`

#### Password
The `password_file` key in config file (default: .restic.pass)<br>
Password is needed when restic process
```
defaultPasswordToEncryptResticSnapshot
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

#### Type
Supported type making restic backup

##### local
At local direcotry of disk. Local is the default backup type if no `--type` option is specified.

Local configuration block is consist of an array with pairs of `src` and `dest` set.
- source - Backup source path
- destination - Backup destination path

```json
"local": [
		{
			"src": "Syncing source path",
			"dest": "Syncing destination path"
		},
		{
			"src": "Syncing source path",
			"dest": "Syncing destination path"
		}
]
```

##### sftp
Backup via SFTP with SSH. To use `sftp` type on backup, set `--type` option with `sftp` value

SFTP configuration block is consist of an array with pairs of `host`, `src`, and `dest` set. 
- source - Backup source path
- destination - Backup destination path
- host (`sftp` type only) - host name set in ssh config file 

```json
"sftp": [
		{
			"host": "Host from ssh config file",
			"src": "Syncing source path",
			"dest": "Syncing destination path"
		},
		{
			"host": "Host from ssh config file",
			"src": "Syncing source path",
			"dest": "Syncing destination path"
		}
]

```

#### Snapshot policy
Set policies to indicate which snapshots to keep and which snapshots to remove

Snapshot policy configuration block is consist of keeping options to indicate the keeping rules

Valid options<br>
For more policy explanation, please checkout restic's [document page](https://restic.readthedocs.io/en/stable/060_forget.html#removing-snapshots-according-to-a-policy)
- `keep_last`
- `keep_hourly`
- `keep_daily`
- `keep_weekly`
- `keep_monthly`
- `keep_yearly`
- `keep_within`
- `keep_within_hourly`
- `keep_within_daily`
- `keep_within_monthly`
- `keep_within_yearly`


```json
"snapshots_policy": {
		"keep_daily": 7,
		"keep_weekly": 5,
		"keep_monthly": 12,
		"keep_yearly": 3
}

```




