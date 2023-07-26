#!/bin/bash
#
# Backup files with custom configuration using restic


# Global variables
declare -r _VERSION=0.2.0

declare -r _SCRIPT=$(readlink -f "${0}")
declare -r _SCRIPT_DIR=$(dirname ${_SCRIPT})

declare -r _VALID_ACTIONS=("backup" "init" "mount" "snapshots")
declare -r _BACKUP_TYPES=("local" "sftp")

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

declare -a _DEST_REPOS
declare -a _SRC_REPOS

# ----------------------------------------------------------------------------
# Show script usage
# Outputs:
#   Write usage information to stdout
# ----------------------------------------------------------------------------
show_help() {
cat << EOF
Usage:  ${0##*/} [--help] [--version] [--config=CONFIG_FILE] [--type=local|sftp] backup|init|mount|snapshots
        ${0##*/} [--help] [--version] [--config=CONFIG_FILE] mount MP

    --help                      Display this help message and exit
    --config=CONFIG_FILE
    --config CONFIG_FILE        Specify which configuration file to use when running the script
                                Default config file: config.json
    --type=[local|sftp]         
    --type [local|sftp]         Specify backup destination type: (local, sftp)
                                Default type: local
    --version                   Show version information

    backup                      Create new backup snapshot
    init                        Prepare backup destination repository
    mount MP                    Mounting backup repository for browsing or restoring
                                MP: Mount point name in configuration .restore block
    snapshots                   List history snapshots
EOF
}

# ----------------------------------------------------------------------
# Check required option file setting in config
# Arguments:
#   key name in config
#   config filename / filepath
# Globals:
#   _SCRIPT_DIR
# Outputs:
#   Write password filename / filepath to stdout
#   Write error message to stderr
# Returns:
#   0 on success, non-zero on error
# -----------------------------------------------------------------------
check_config_required_file() {
    if [[ "${#}" -ne 2 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _key=${1}
    local _config=${2}
    local _file

    case ${_key} in
        password_file)
            _file=$(jq -r .password_file ${_config})
            ;;
        exclude_file)
            _file=$(jq -r .exclude_file ${_config})
            ;;
        *)
            echo "[ERROR] Not supported key - ${_key}" >&2
            return 1
    esac
    
    # Using absolute filepath if not already set
    if [[ "${_file}" != /* ]]; then
        _file=${_SCRIPT_DIR}/${_file}
    fi

    if [[ ! -f "${_file}" ]]; then
        echo "[ERROR] ${_key} file: ${_file} not found" >&2
        return 1
    fi
    echo "${_file}"
}

# -------------------------------------------------------------------------------------------------
# Read mount point configuration
# Arguments:
#   config filename / filepath
#   key name of mount point in configuration
# Globals:
#   _DEST_REPOS
#   _SRC_REPOS
# Outputs:
#   Write error messages to stderr
# Returns:
#   non-zero on error
# -------------------------------------------------------------------------------------------------
read_mount_point() {
    if [[ "${#}" -ne 2 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _config=${1}
    local _mount_point=${2}

    # Check mount point key existense
    if [[ $(jq --arg _mp "${_mount_point}" '.mount | has($_mp)' ${_config}) != "true" ]]; then
        echo "[ERROR] .mount.${_mount_point} not set" >&2
        return 1
    fi

    local _type=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].type' ${_config})
    if [[ "${_type}" != "local" && "${_type}" != "sftp" ]]; then
        echo "[ERROR] .mount.${_mount_point}.type value not set or invalid" >&2
        return 1
    fi

    local _host=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].host' ${_config})
    if [[ "${_type}" == "sftp" && "${_host}" == "null" ]]; then
        echo "[ERROR] .mount.${_mount_point}.host value not set with type ${_type}" >&2
        return 1
    fi

    local _src=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].src' ${_config})
    if [[ "${_src}" == "null" ]]; then 
        echo "[ERROR] .mount.${_mount_point}.src value not set" >&2
        return 1
    fi

    local _dest=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].dest' ${_config})
    if [[ "${_dest}" == "null" ]]; then 
        echo "[ERROR] .mount.${_mount_point}.dest value not set" >&2
        return 1
    elif [[ ! -d "${_dest}" ]]; then
        echo "[ERROR] destination repository ${_dest} not exist" >&2
        return 1
    fi

    _DEST_REPOS[0]="${_dest}"
    case ${_type} in 
        local)
            _SRC_REPOS[0]="${_src}"
        ;;
        sftp)
            _SRC_REPOS[0]="sftp:${_host}:${_src}"
        ;;
        *)
            echo "[ERROR] invalid type value: ${_type}"
            return 1
    esac
}

# -------------------------------------------------------------------------------
# Read local block configuration 
# Arguments:
#   config filename / filepath
# Globals
#   _DEST_REPOS
#   _SRC_REPOS
# Returns:
#   2 if function usage error
# -------------------------------------------------------------------------------
read_local() {
    if [[ "${#}" -ne 1 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _config=${1}
    local _block=$(jq -r .local ${_config})
    local _block_len=$(jq length <<< ${_block})

    for (( i=0; i<${_block_len}; i++ )); do
        local _src=$(jq -r --arg i "${i}" '.local[$i|tonumber].src[]' ${_config})
        local _dest=$(jq -r --arg i "${i}" '.local[$i|tonumber].dest' ${_config})
        _DEST_REPOS[${i}]="${_dest}"
        _SRC_REPOS[${i}]="${_src}"
    done
}

# --------------------------------------------------------------------------------
# Read sftp block configuration
# Arguments:
#   config filename / filepath
# Globals:
#   _DEST_REPOS
#   _SRC_REPOS
# Returns:
#   2 if function usage error
# --------------------------------------------------------------------------------
read_sftp() {
    if [[ "${#}" -ne 1 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _config=${1}
    local _block=$(jq -r .sftp ${_config})
    local _block_len=$(jq length <<< ${_block})

    for (( i=0; i<${_block_len}; i++ )); do
        local _host=$(jq -r --arg i "${i}" '.sftp[$i|tonumber].host' ${_config})
        local _src=$(jq -r --arg i "${i}" '.sftp[$i|tonumber].src[]' ${_config})
        local _dest=$(jq -r --arg i "${i}" '.sftp[$i|tonumber].dest' ${_config})
        _DEST_REPOS[${i}]="sftp:${_host}:${_dest}"
        _SRC_REPOS[${i}]="${_src}"
    done
}

# -----------------------------------------------------------------------------
# Get backup usage conifuration setting according to given type of destination
# (local, sftp)
# Arguments:
#   backup type, string
#   config filename / filepath
# Returns:
#   non-zero on error
# -----------------------------------------------------------------------------
summarize_backup_config() {
    if [[ "${#}" -ne 2 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _type=${1}
    local _config=${2}

    case ${_type} in
        local)
            read_local ${_config}
            (( ${?} == 0 )) || return 1
        ;;
        sftp)
            read_sftp ${_config}
            (( ${?} == 0 )) || return 1
        ;;
        *)
            return 1
    esac
}

# ------------------------------------------------------------------------------
# Init restic repositories in configuration setting
# Arguments:
#   backup type, string
#   config filename / filepath
#   restic password filepath
# Globals:
#   _DEST_REPOS
# Returns:
#   non-zero on error
# ------------------------------------------------------------------------------
restic_init() {
    if [[ "${#}" -ne 3 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _type=${1}
    local _config=${2}
    local _password_file=${3}

    summarize_backup_config ${_type} ${_config} 
    (( ${?} == 0 )) || return 1

    for (( i=0; i<${#_DEST_REPOS[@]}; i++ )); do
        echo "[INFO] Destination: ${_DEST_REPOS[${i}]}"
        restic init -r ${_DEST_REPOS[${i}]} --password-file ${_password_file}
        echo ""
    done
}

# ---------------------------------------------------------------------------
# List restic snapshots records
# Arguments:
#   backup type, string
#   config filename / filepath
#   restic password filepath
# Globals:
#   _DEST_REPOS
# Returns:
#   non-zero on error
# ---------------------------------------------------------------------------
restic_snapshots() {
    if [[ "${#}" -ne 3 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _type=${1}
    local _config=${2}
    local _password_file=${3}

    summarize_backup_config ${_type} ${_config} 
    (( ${?} == 0 )) || return 1

    for (( i=0; i<${#_DEST_REPOS[@]}; i++ )); do
        echo "[INFO] Destination: ${_DEST_REPOS[${i}]}"
        restic snapshots -r ${_DEST_REPOS[${i}]} --password-file ${_password_file}
        echo ""
    done
}

# ---------------------------------------------------------------------------------
# Create new restic backup snapshot
# Arguments:
#   backup type, string
#   config filename / filepath
#   restic password filepath
#   exclude files filepath
# Globals:
#   _FORGET_POLICY
#   _SRC_REPOS
#   _DEST_REPOS
# Returns:
#   non-zero on error
# ---------------------------------------------------------------------------------
restic_backup() {
    if [[ "${#}" -ne 4 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _type=${1}
    local _config=${2}
    local _password_file=${3}
    local _exclude_file=${4}

    summarize_backup_config ${_type} ${_config} 
    (( ${?} == 0 )) || return 1

    local _snapshots_policy=$(jq '.snapshots_policy' ${_config})
    local _forget_options=""
    if [[ "${_snapshots_policy}" != "null" ]]; then
        local _policies=($(echo ${_snapshots_policy} | jq -r 'keys[]'))
        local _policy_flag
        local _value
        for _policy in ${_policies[@]}; do
            if [[ "${_FORGET_POLICY[${_policy}]}" == "" ]]; then
                echo "[ERROR] invalid snapshot policy: ${_policy}" >&2
                return 1
            fi
            _policy_flag=${_FORGET_POLICY[${_policy}]}
            _value=$(echo ${_snapshots_policy} | jq -r --arg policy "${_policy}" '.[$policy]')
            _forget_options="${_forget_options} --${_policy_flag} ${_value}"
        done
    fi
    local _src_paths
    for (( i=0; i<${#_SRC_REPOS[@]}; i++ )); do
        readarray -t _src_paths <<< "${_SRC_REPOS[${i}]}"
        for (( j=0; j<${#_src_paths[@]}; j++ )); do
            echo "[INFO] Source: ${_src_paths[${j}]}, Destination: ${_DEST_REPOS[${i}]}"
            restic backup -v -r ${_DEST_REPOS[${i}]} --exclude-file="${_exclude_file}" --password-file ${_password_file} ${_src_paths[${j}]}
        done
        echo "Backup check"
        restic check -v -r ${_DEST_REPOS[${i}]} --password-file ${_password_file}
        echo "Backup clean"
        if [[ "${_forget_options}" == "" ]]; then
            restic prune -r ${_DEST_REPOS[${i}]} --password-file ${_password_file}
        else
            restic forget -v -r ${_DEST_REPOS[${i}]} --password-file ${_password_file} ${_forget_options} --prune
        fi
        echo ""
    done
}

# ---------------------------------------------------------------------------
# Mount restic backup as regular file system for browsing / restore
# Arguments:
#   config filename / filepath
#   key name of mount point in configuration
# Returns:
#   non-zero on error
# ---------------------------------------------------------------------------
restic_mount() {
    if [[ "${#}" -ne 2 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _config=${1}
    local _mount_point=${2}

    read_mount_point ${_config} ${_mount_point}
    (( ${?} == 0 )) || return 1

    restic -r ${_SRC_REPOS[0]} mount ${_DEST_REPOS[0]}
}

main() {
    local _backup_type="local"
    local _config="${_SCRIPT_DIR}/config.json"

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

    if [[ ${#} -eq 1 && ${1} != "mount" ]]; then
        local _action=${1}
    elif [[ ${#} -eq 2 && ${1} == "mount" ]]; then
        local _action=${1}
        local _mount_point=${2}
    else
        show_help
        exit 1
    fi

    if [[ ! " ${_VALID_ACTIONS[*]} " =~ " ${_action} " ]]; then
        show_help
        exit
    fi

    if [[ ! " ${_BACKUP_TYPES[*]} " =~ " ${_backup_type} "  ]]; then
        show_help
        exit 1
    fi

    if [[ ! -f "${_config}" ]]; then
        echo "[ERROR] config file: ${_config} not found"
        exit 1
    fi

    local _password_file
    _password_file=$(check_config_required_file password_file ${_config})
    (( ${?} == 0 )) || exit 1

    local _exclude_file=$(check_config_required_file exclude_file ${_config})
    (( ${?} == 0 )) || exit 1

    echo "Backup type: ${_backup_type}"
    echo "Secret password file: ${_password_file}"
    echo "Exclude content file: ${_exclude_file}"
    echo ""

    case ${_action} in
        backup)
            restic_backup ${_backup_type} ${_config} ${_password_file} ${_exclude_file}
            (( ${?} == 0 )) || exit 1
        ;;
        init)
            restic_init ${_backup_type} ${_config} ${_password_file}
            (( ${?} == 0 )) || exit 1
        ;;
        mount)
            restic_mount ${_config} ${_mount_point}
            (( ${?} == 0 )) || exit 1
        ;;
        snapshots)
            restic_snapshots ${_backup_type} ${_config} ${_password_file}
            (( ${?} == 0 )) || exit 1
        ;;
        *)
        echo "[WARNING] Unsupported action - ${_action}"
        exit 1
    esac
}

main "$@"
