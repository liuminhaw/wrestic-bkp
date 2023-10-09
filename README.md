# restic-bkp
Backup files with custom configuration using restic

## Roadmap
[Project Roadmap](https://mewing-pisces-b2c.notion.site/restic-bkp-1b852fb0215a4e1d8df00081f3050e1b) 

## Pre-requisite
1. `restic` need to be installed before executing the script <br>
**reference:** [Installation](https://restic.readthedocs.io/en/latest/020_installation.html)
1. `jq` need to be installed before executing the script <br>
**reference:** [Installation](https://stedolan.github.io/jq/download/)


## Usage
```
Usage:  wrestic-bkp.sh [--help] [--version] [--config=CONFIG_FILE] [--type=local|sftp|s3] backup|init|snapshots
        wrestic-bkp.sh [--help] [--version] [--config=CONFIG_FILE] mount MP

    --help                      Display this help message and exit
    --config=CONFIG_FILE
    --config CONFIG_FILE        Specify which configuration file to use when running the script
                                Default config file: config.json
    --type=[local|sftp|s3]
    --type [local|sftp|s3]      Specify backup destination type: (local, sftp)
                                Default type: local
    --version                   Show version information

    backup                      Create new backup snapshot
    init                        Prepare backup destination repository
    mount MP                    Mounting backup repository for browsing or restoring
                                MP: Mount point name in configuration .restore block
    snapshots                   List history snapshots
```

## Setup
### Configuration
Set backup information (source, destination, password, exclusion, snapshot policy)
- [password file](#password) - Filename where encryption password is stored 
    - Set to `.restic.pass` as default, modify its content for custom password
- [exclusive file](#exclusion) - Files or direcotries to skip while backing up, one file / directory per line
    - Set to `excludes.txt` as default, modify its content for custom settings
- [type](#type) - Current support types: `local` `sftp` `s3`
- [mount](#mount) - mounts configuration to read using `mount` action
- [snapshot policy](#snapshot-policy) - Rules to indicate which snapshots to keep / remove

Configuration file can be copied and edit from `config.template`

#### Password
The `password_file` key in config file (default: .restic.pass)<br>
Password is needed when restic process
```text
defaultPasswordToEncryptResticSnapshot
```

#### Exclusion
The `exclude_file` key in config file (default: excludes.txt)<br>
Create the file with listed files and directories for exclusion

```text
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
At local directory of disk. Local is the default backup type if no `--type` option is specified.

Local configuration block is consisted of an array with pairs of `src` and `dest` set.
- source: Backup source paths, list of paths to backup
- destination: Backup destination path
- src_in_one: Set `true` to backup listed source paths in single snapshot (Optional, default: `false`)

```json
"local": [
    {
        "src": [
            "Syncing source path 1",
            "Syncing source path 2"
        ],
        "dest": "Syncing destination path",
        "src_in_one": true
    },
    {
        "src": [
            "Syncing source path 1",
            "Syncing source path 2"
        ],
        "dest": "Syncing destination path"
    }
]
```

##### sftp
Backup via SFTP with SSH. To use `sftp` type on backup, set `--type` option with `sftp` value

SFTP configuration block is consisted of an array with pairs of `host`, `src`, and `dest` set. 
- source: Backup source paths, list of paths to backup
- destination: Backup destination path
- host (`sftp` type only): host name set in ssh config file 
- src_in_one: Set `true` to backup listed source paths in single snapshot (Optional, default: `false`)

```json
"sftp": [
    {
        "host": "Host from ssh config file",
        "src": [
            "Syncing source path 1",
            "Syncing source path 2"
        ],
        "dest": "Syncing destination path",
        "src_in_one": true
    },
    {
        "host": "Host from ssh config file",
        "src": [
            "Syncing source path 1",
            "Syncing source path 2"
        ],
        "dest": "Syncing destination path"
    }
]

```
ssh config file format
```text
Host HOST_NAME
    Hostname HOST_ADDRESS/DOMAIN
    User USERNAME
    Port CONNECTION_PORT
    Identityfile /path/to/identity/file
    ServerAliveInterval 60
    ServerAliveCountMax 240
```

##### s3
Backup to s3 bucket as destination. To use `s3` type on backup, set `--type` option with `s3` value

s3 configuration block is consisted of an array with pairs of `aws_access_key_id`, `aws_secret_access_key`, `aws_region`, `src`, and `dest` set. 
- source: Backup source paths, list of paths to backup
- destination: Destination bucket path
- aws credential: Grant permission to perform action on s3 bucket
    ```bash
    # Permission on objects
    Service: S3
    Allow Actions: DeleteObject, GetObject, PutObject
    Resources: arn:aws:s3:::restic-demo/*

    # Permission on bucket
    Service: S3
    Allow Actions: ListBucket, GetBucketLocation
    Resource: arn:aws:s3:::restic-demo
    ```
- src_in_one: Set `true` to backup listed source paths in single snapshot (Optional, default: `false`)

```json
"s3": [
    {
        "aws_access_key_id": "access key id credential",
        "aws_secret_access_key": "secret access key credential",
        "aws_region": "aws region",
        "src": [
            "Syncing source path 1",
            "Syncing source path 2"
        ],
        "dest": "Syncing destination path (bucket/path/to/backup)",
        "src_in_one": true
    }
]
```

#### Mount
Mount points (MP value in help message) to indicate when using `mount` action

Read configuration from given argument and mount with according block settings

Supporting methods: `local` `sftp` `s3`

##### local
Snapshot is backup on local directory or disk
```json
"mount_point_name": {
    "type": "local",
    "src": "restic backup location",
    "dest": "Local mount point destination",
    "paths": [
        "/snapshot/matched/path",
        "/other/snapshot/matched/path"
    ]
}
```
- type: Specified `local` type is used
- src: Restic backup repository path
- dest: Mount point to serve the repository

##### sftp
Snapshot is backup using SFTP
```json
"mount_point_name": {
    "type": "sftp",
    "host": "Host from ssh config file",
    "src": "restic backup location",
    "dest": "Local mount point desitnation"
}
```
- type: Specified `sftp` type is used
- host: SSH host name set in ssh config file
- src: Restic backup repository path
- dest: Mount point to serve the repository

##### s3
Snapshot is backup using s3
```json
"mount_point_3": {
    "type": "s3",
    "aws_access_key_id": "access key id credential",
    "aws_secret_access_key": "secret access key credential",
    "aws_region": "aws region",
    "src": "restic backup location (bucket/path/to/backup)",
    "dest": "Local mount point destination"
}
```

##### Optional keys
- `default_password`: Set if use `password_file` as decrypt password. Interactively ask for password if set to `false` (default value `false`).
- `paths`: Mount snapshots with given paths only, default to show all snapshots if not set.

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

