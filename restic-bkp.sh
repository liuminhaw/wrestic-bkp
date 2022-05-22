#!/bin/bash
#
# Backup files with custom configuration using restic


#  var
declare -r _VERSION=0.1.1

declare -r -A _FORGET_POLICY=(
    ["keep_last"]="keep-last"
    ["keep_hourly"]="keep-hourly"
    ["keep_daily"]="keep-daily"
    ["keep_weekly"]="keep-weekly"
    ["keep_monthly"]="keep-monthly"
    ["keep_yearly"]="keep-yearly"
    ["keep_within"]="keep-within"
    ["keep_within_hourly"]="keep-within-hourly"
    ["keep_within_daily"]="keep-within-daily"
    ["keep_within_monthly"]="keep-within-monthly"
    ["keep_within_yearly"]="keep-within-yearly"
)

# ----------------------------------------------------------------------------
# Show script usage
# Outputs:
#   Write usage information to stdout
# ----------------------------------------------------------------------------
show_help() {
cat << EOF
Usage: ${0##*/} [--help] [--version] [--config=CONFIG_FILE] [--type=local|sftp] backup|init|snapshots
    --help                      Display this help message and exit
    --config=CONFIG_FILE
    --config CONFIG_FILE        Specify which configuration file to use when running the script
                                Default config file: config.json
    --type=[local|sftp]         
    --type [local|sftp]         Specify backup destination type: (local, sftp)
                                Default type: local
    --version                   Show version information
    action                      Command to execute: (backup, init, snapshots)
                                backup: create new backup snapshot
                                init: prepare backup destination directory
                                snapshots: list previous snapshots
EOF
}

_config="config.json"
_backup_type="local"


# Command line options
while :; do
    case ${1} in
        --help)
            show_help
            exit
            ;;
        --version)
            echo "Version: ${_VERSION}"
            exit
            ;;
        --config)
            if [[ "${2}" ]]; then
                _config=${2}
                shift
            else
                echo -e "[ERROR] '--config' requires a non-empty option argument." 1>&2
                exit 1
            fi
            ;;
        --config=?*)
            _config=${1#*=} # Delete everything up to "=" and assign the remainder
            ;;
        --config=)
            echo -e "[ERROR] '--config' requires a non-empty option argument." 1>&2
            exit 1
            ;;
        --type)
            if [[ "${2}" ]]; then
                _backup_type=${2}
                shift
            else
                echo -e "[ERROR] '--type' requires a non-empty option argument." 1>&2
                exit 1
            fi
            ;;
        --type=?*)
            _backup_type=${1#*=} # Delete everything up to "=" and assign the remainder
            ;;
        --type=)
            echo -e "[ERROR] '--type' requires a non-empty option argument." 1>&2
            exit 1
            ;;
        -?*)
            echo -e "[WARN] Unknown option (ignored): ${1}" 1>&2
            exit 1
            ;;
        *)  # Default case: no more options
            break
    esac

    shift
done

if [[ ${#} != 1 ]]; then
    show_help
    exit 1
fi
_action=${1}

if [[ "${_action}" != "backup" && "${_action}" != "init" && "${_action}" != "snapshots" ]]; then
    show_help
    exit 1
fi

if [[ ! -z "${_backup_type}" && "${_backup_type}" -ne "local" && "${_backup_type}" -ne "sftp" ]]; then
    show_help
    exit 1
fi

if [[ ! -f "${_config}" ]]; then
    echo "[ERROR] config file: ${_config} not found"
    exit 1
fi

_password_file=$(jq -r .password_file ${_config})
_exclude_file=$(jq -r .exclude_file ${_config})

if [[ ! -f "${_password_file}" ]]; then
    echo "[ERROR] secret password file: ${_password_file} not found"
    exit 1
fi

if [[ ! -f "${_exclude_file}" ]]; then
    echo "[ERROR] exclude content file: ${_exclude_file} not found"
    exit 1
fi

if [[ "${_backup_type}" == "local" ]]; then
    _local=$(jq -r .local ${_config})
    _local_len=$(jq length <<< ${_local})
    for (( i=0; i<${_local_len}; i++ )); do
        _local_src=$(jq -r --arg i "${i}" '.local[$i|tonumber].src' ${_config})
        _local_dest=$(jq -r --arg i "${i}" '.local[$i|tonumber].dest' ${_config})
        _dest_repos[${i}]="${_local_dest}"
        _src_repos[${i}]="${_local_src}"
    done
elif [[ "${_backup_type}" == "sftp" ]]; then
    _sftp=$(jq -r .sftp ${_config})
    _sftp_len=$(jq length <<< ${_sftp})
    for (( i=0; i<${_sftp_len}; i++ )); do
        _sftp_host=$(jq -r --arg i "${i}" '.sftp[$i|tonumber].host' ${_config})
        _sftp_src=$(jq -r --arg i "${i}" '.sftp[$i|tonumber].src' ${_config})
        _sftp_dest=$(jq -r --arg i "${i}" '.sftp[$i|tonumber].dest' ${_config})
        _dest_repos[${i}]="sftp:${_sftp_host}:${_sftp_dest}"
        _src_repos[${i}]="${_sftp_src}"
    done
else
    echo "[ERROR] invalid type parameter: ${_backup_type}"
    exit 1
fi

echo "Backup type: ${_backup_type}"
echo "Secret password file: ${_password_file}"
echo "Exclude content file: ${_exclude_file}"
echo ""

case ${_action} in
    init)
        for (( i=0; i<${#_dest_repos[@]}; i++ )); do
            echo "[INFO] Destination: ${_dest_repos[${i}]}"
            restic init -r ${_dest_repos[${i}]} --password-file ${_password_file}
            echo ""
        done
        ;;
    snapshots)
        for (( i=0; i<${#_dest_repos[@]}; i++ )); do
            echo "[INFO] Destination: ${_dest_repos[${i}]}"
            restic snapshots -r ${_dest_repos[${i}]} --password-file ${_password_file}
            echo ""
        done
        ;;
    backup)
        _snapshots_policy=$(jq '.snapshots_policy' ${_config})
        _forget_options=""
        if [[ "${_snapshots_policy}" != "null" ]]; then
            _policies=($(echo ${_snapshots_policy} | jq -r 'keys[]'))
            for _policy in ${_policies[@]}; do
                if [[ "${_FORGET_POLICY[${_policy}]}" == "" ]]; then
                    echo "[ERROR] invalid snapshot policy: ${_policy}"
                    exit 1
                fi
                _policy_flag=${_FORGET_POLICY[${_policy}]}
                _value=$(echo ${_snapshots_policy} | jq -r --arg policy "${_policy}" '.[$policy]')
                _forget_options="${_forget_options} --${_policy_flag} ${_value}"
            done
        fi
        for (( i=0; i<${#_src_repos[@]}; i++ )); do
            echo "[INFO] Source: ${_src_repos[${i}]}, Destination: ${_dest_repos[${i}]}"
            restic backup -v -r ${_dest_repos[${i}]} --exclude-file="${_exclude_file}" --password-file ${_password_file} ${_src_repos[${i}]}
            if [[ "${_forget_options}" == "" ]]; then
                restic prune -r ${_dest_repos[${i}]} --password-file ${_password_file}
            else
                restic forget -v -r ${_dest_repos[${i}]} --password-file ${_password_file} ${_forget_options} --prune
            fi
            echo ""
        done
        ;;
esac

