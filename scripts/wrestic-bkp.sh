#!/bin/bash
#
# Backup files with custom configuration using restic


# Global variables
declare -r _VERSION=0.3.0

declare -r _SCRIPT=$(readlink -f "${0}")
declare -r _SCRIPT_DIR=$(dirname ${_SCRIPT})

declare -r _VALID_ACTIONS=("backup" "init" "mount" "snapshots")
declare -r _BACKUP_TYPES=("local" "sftp" "s3") 

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

declare -a _SRC_IN_ONE
declare -a _DEST_REPOS
declare -a _SRC_REPOS
declare -a _REPO_CREDS
declare -a _TAGS

# ----------------------------------------------------------------------------
# Show script usage
# Outputs:
#   Write usage information to stdout
# ----------------------------------------------------------------------------
show_help() {
cat << EOF
Usage:  ${0##*/} [--help] [--version] [--config=CONFIG_FILE] [--type=local|sftp|s3] backup|init|snapshots
        ${0##*/} [--help] [--version] [--config=CONFIG_FILE] mount MP

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
#   _REPO_CREDS
# Outputs:
#   Write error messages to stderr
# Returns:
#   1 on execution error
#   2 on function usage error
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
    if [[ ! " ${_BACKUP_TYPES[*]} " =~ " ${_type} "  ]]; then
        echo "[ERROR] .mount.${_mount_point}.type value not set or invalid" >&2
        return 1
    fi

    case ${_type} in
        sftp)
            local _host=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].host' ${_config})
            if [[ "${_host}" == "null" ]]; then
                echo "[ERROR] .mount.${_mount_point}.host value not set with type ${_type}" >&2
                return 1
            fi
        ;;
        s3)
            local _aws_profile=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].aws_profile_name' ${_config})
            local _aws_access_key_id=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].aws_access_key_id' ${_config})
            local _aws_secret_access_key=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].aws_secret_access_key' ${_config})
            local _aws_region=$(jq -r --arg _mp "${_mount_point}" '.mount[$_mp].aws_region' ${_config})
            aws_creds_check "${_aws_profile}" "${_aws_access_key_id}" "${_aws_secret_access_key}" "${_aws_region}"
            (( ${?} == 0 )) || return 1
        ;;
    esac

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
        s3)
            _SRC_REPOS[0]="s3:s3.amazonaws.com/${_src}"
            if [[ -n ${_aws_profile} && "${_aws_profile}" != "null" ]]; then
                _REPO_CREDS[${i}]="aws-profile:${_aws_profile}"
            else
                _REPO_CREDS[0]="aws-key:${_aws_access_key_id}:${_aws_secret_access_key}:${_aws_region}"
            fi
        ;;
        *)
            echo "[ERROR] invalid type value: ${_type}"
            return 1
    esac

    repo_permission_init ${_type} "${_REPO_CREDS[0]}"
    (( ${?} == 0 )) || return 1
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
        local _src_in_one=$(jq -r --arg i "${i}" '.local[$i|tonumber].src_in_one' ${_config})
        local _tags=$(jq -r --arg i "${i}" '.local[$i|tonumber].tags' ${_config})
        _DEST_REPOS[${i}]="${_dest}"
        _SRC_REPOS[${i}]="${_src}"
        _SRC_IN_ONE[${i}]="${_src_in_one}"
        if [[ "${_tags}" == "null" ]]; then
            _TAGS[${i}]=""
        else
            _TAGS[${i}]=$(jq -r '.[]' <<< ${_tags})
        fi
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
        local _src_in_one=$(jq -r --arg i "${i}" '.sftp[$i|tonumber].src_in_one' ${_config})
        local _tags=$(jq -r --arg i "${i}" '.sftp[$i|tonumber].tags' ${_config})
        _DEST_REPOS[${i}]="sftp:${_host}:${_dest}"
        _SRC_REPOS[${i}]="${_src}"
        _SRC_IN_ONE[${i}]="${_src_in_one}"
        if [[ "${_tags}" == "null" ]]; then
            _TAGS[${i}]=""
        else
            _TAGS[${i}]=$(jq -r '.[]' <<< ${_tags})
        fi
    done
}


# --------------------------------------------------------------------------------
# Read s3 block configuration
# Arguments:
#   config filename / filepath
# Globals:
#   _DEST_REPOS
#   _SRC_REPOS
#   _REPO_CREDS
# Returns:
#   1 on function execution error
#   2 on function usage error
# --------------------------------------------------------------------------------
read_s3() {
    if [[ "${#}" -ne 1 ]]; then
        echo "[ERROR] Function ${FUNCNAME[0]} usage error" >&2
        return 2
    fi

    local _config=${1}
    local _block
    local _block_len

    _block=$(jq -r .s3 "${_config}")
    _block_len=$(jq length <<< "${_block}")

    local _aws_profile
    local _aws_access_key_id
    local _aws_secret_access_key
    local _aws_region
    local _src
    local _dest
    local _src_in_one
    local _tags
    for (( i=0; i<_block_len; i++ )); do
        _aws_profile=$(jq -r --arg i "${i}" '.s3[$i|tonumber].aws_profile_name' "${_config}")
        _aws_access_key_id=$(jq -r --arg i "${i}" '.s3[$i|tonumber].aws_access_key_id' "${_config}")
        _aws_secret_access_key=$(jq -r --arg i "${i}" '.s3[$i|tonumber].aws_secret_access_key' "${_config}")
        _aws_region=$(jq -r --arg i "${i}" '.s3[$i|tonumber].aws_region' "${_config}")
        _src=$(jq -r --arg i "${i}" '.s3[$i|tonumber].src[]' "${_config}")
        _dest=$(jq -r --arg i "${i}" '.s3[$i|tonumber].dest' "${_config}")
        _src_in_one=$(jq -r --arg i "${i}" '.s3[$i|tonumber].src_in_one' "${_config}")
        _tags=$(jq -r --arg i "${i}" '.s3[$i|tonumber].tags' "${_config}")

        _DEST_REPOS[${i}]="s3:s3.amazonaws.com/${_dest}"
        _SRC_REPOS[${i}]="${_src}"
        _SRC_IN_ONE[${i}]="${_src_in_one}"

        aws_creds_check "${_aws_profile}" "${_aws_access_key_id}" "${_aws_secret_access_key}" "${_aws_region}"
        (( ${?} == 0 )) || return 1

        if [[ -n ${_aws_profile} && ${_aws_profile} != "null" ]]; then 
            _REPO_CREDS[${i}]="aws-profile:${_aws_profile}"
        else
            _REPO_CREDS[${i}]="aws-key:${_aws_access_key_id}:${_aws_secret_access_key}:${_aws_region}"
        fi
        if [[ "${_tags}" == "null" ]]; then
            _TAGS[${i}]=""
        else
            _TAGS[${i}]=$(jq -r '.[]' <<< ${_tags})
        fi
    done
}

# --------------------------------------------------------------------------------
# Set mount options reading from config file
# Arguments:
#   config filename / filepath
#   restic password filepath
# Outputs:
#   write mount options to stdout on success
# --------------------------------------------------------------------------------
read_mount_options() {
    if [[ "${#}" -ne 2 ]]; then
        echo "[ERROR] Function ${FUNCNAME[0]} usage error" >&2
        return 2
    fi

    local _config=${1}
    local _password_file=${2}
    local _options

    local _use_default_password=$(jq -r --arg mp "${_mount_point}" '.mount[$mp].default_password' ${_config})
    if [[ "${_use_default_password}" == "true" ]]; then
        _options+="--password-file ${_password_file} "
    fi

    local _paths
    _paths=($(jq -r --arg mp "${_mount_point}" '.mount[$mp].paths[]' ${_config} 2> /dev/null))
    if [[ ${?} -eq 0 ]]; then
        for _path in "${_paths[@]}"; do
            _options+="--path ${_path} "
        done
    fi
    local _tags
    _tags=($(jq -r --arg mp "${_mount_point}" '.mount[$mp].tags[]' ${_config} 2> /dev/null))
    if [[ ${?} -eq 0 ]]; then
        for _tag in "${_tags[@]}"; do
            _options+="--tag ${_tag} "
        done
    fi

    echo "${_options}"
}

# -----------------------------------------------------------------------------
# Get backup usage configuration setting according to given type of destination
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
        s3)
            read_s3 ${_config}
            (( ${?} == 0 )) || return 1
        ;;
        *)
            return 1
    esac
}

# ---------------------------------------------------------------------------------
# Verify AWS credential config, aws profile or aws key / secret pair should be set
# Arguments:
#   aws profile
#   aws access key id
#   aws secret access key
#   aws region
# Returns:
#   1 on invalid setting
#   2 on usage error
# ---------------------------------------------------------------------------------
aws_creds_check() {
    if [[  "${#}" -ne 4 ]]; then
        echo "[ERROR] Function ${FUNCNAME[0]} usage error" >&2
        return 2
    fi

    local _aws_profile=${1}
    local _aws_access_key_id=${2}
    local _aws_secret_access_key=${3}
    local _aws_region=${4}

    if [[ -n ${_aws_profile} && "${_aws_profile}" != "null" ]]; then
        echo "Using aws profile: ${_aws_profile}"
        return 0
    fi

    echo "AWS profile not set, use config key and secret"
    if [[ -z ${_aws_access_key_id} || "${_aws_access_key_id}" == "null" ]]; then
        echo "No aws access key, please check config setting"
        return 1
    fi
    if [[ -z ${_aws_secret_access_key} || "${_aws_secret_access_key}" == "null" ]]; then
        echo "No aws secret access key, please check config setting"
        return 1
    fi
    if [[ -z ${_aws_region} || "${_aws_region}" == "null" ]]; then
        echo "No aws region, please check config setting"
        return 1
    fi
}

# ------------------------------------------------------------------------------
# Init aws credentials
# Arguments:
#   aws credentials in format accessKeyId:secretAccessKey:region, string
# Globals:
#   AWS_ACCESS_KEY_ID
#   AWS_SECRET_ACCESS_KEY
#   AWS_DEFAULT_REGION
# Returns:
#   non-zero on error
# ------------------------------------------------------------------------------
aws_init() {
    if [[ "${#}" -ne 1 ]]; then
        echo "[ERROR] Function ${FUNCNAME[0]} usage error" >&2
        return 2
    fi

    local _aws_creds=${1}
    local _aws_cred_type
    local _aws_profile
    local _aws_access_key_id
    local _aws_secret_access_key
    local _aws_region

    _aws_cred_type=$(cut -d: -f1 <<< "${_aws_creds}")

    case "${_aws_cred_type}" in
    aws-profile)
        _aws_profile=$(cut -d: -f2 <<< "${_aws_creds}")
        export AWS_PROFILE=${_aws_profile}
    ;;
    aws-key)
        _aws_access_key_id=$(cut -d: -f2 <<< "${_aws_creds}")
        _aws_secret_access_key=$(cut -d: -f3 <<< "${_aws_creds}")
        _aws_region=$(cut -d: -f4 <<< "${_aws_creds}")
        export AWS_ACCESS_KEY_ID=${_aws_access_key_id}
        export AWS_SECRET_ACCESS_KEY=${_aws_secret_access_key}
        export AWS_DEFAULT_REGION=${_aws_region}
    ;;
    *)
        echo "[ERROR] Not supported cred type: ${_aws_cred_type}"
        return 1
    esac
}


# ------------------------------------------------------------------------------
# Init permissiong for restic repository if needed for type of destination
# Arguments:
#   backup type
#   type required credential
#     s3 format: accessKeyId:secretAccessKey:region
# Returns:
#   non-zero on error
# ------------------------------------------------------------------------------
repo_permission_init() {
    if [[ "${#}" -ne 2 ]]; then
        echo "[ERROR] Function ${FUNCNAME[0]} usage error" >&2
        return 2
    fi

    local _type=${1}
    local _cred=${2}

    case ${_type} in
        s3)
            if ! aws_init "${_cred}"; then
                echo "[ERROR] repo_permission_init: aws_init failed"
                return 1
            fi
        ;;
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
        repo_permission_init ${_type} "${_REPO_CREDS[${i}]}"
        (( ${?} == 0 )) || return 1
        echo "[INFO] Destination: ${_DEST_REPOS[${i}]}"
        restic init -r ${_DEST_REPOS[${i}]} --password-file ${_password_file}
        (( ${?} == 0 )) || return 1
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
        repo_permission_init ${_type} "${_REPO_CREDS[${i}]}"
        (( ${?} == 0 )) || return 1
        echo "[INFO] Destination: ${_DEST_REPOS[${i}]}"
        restic snapshots -r ${_DEST_REPOS[${i}]} --password-file ${_password_file}
        (( ${?} == 0 )) || return 1
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
        for _policy in "${_policies[@]}"; do
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
    local _tags
    local _tag_options=""
    for (( i=0; i<${#_SRC_REPOS[@]}; i++ )); do
        repo_permission_init ${_type} "${_REPO_CREDS[${i}]}"
        (( ${?} == 0 )) || return 1
        echo "Read tags"
        readarray -t _tags <<< "${_TAGS[${i}]}"
        if [[ -z ${_tags[*]} ]]; then
            _tag_options=""
        else
            for (( j=0; j<${#_tags[@]}; j++ )); do
                _tag_options="--tag ${_tags[j]} ${_tag_options}"
            done
        fi

        readarray -t _src_paths <<< "${_SRC_REPOS[${i}]}"
        if [[ "${_SRC_IN_ONE[${i}]}" == "true" ]]; then
            echo "[INFO] Source: ${_src_paths[*]}, Destination: ${_DEST_REPOS[${i}]}"
            restic backup -v -r ${_DEST_REPOS[${i}]} \
                --exclude-file="${_exclude_file}" \
                --password-file ${_password_file} \
                ${_tag_options} \
                "${_src_paths[@]}"
            (( ${?} == 0 )) || return 1
        else
            for (( j=0; j<${#_src_paths[@]}; j++ )); do
                echo "[INFO] Source: ${_src_paths[${j}]}, Destination: ${_DEST_REPOS[${i}]}"
                restic backup -v -r ${_DEST_REPOS[${i}]} \
                    --exclude-file="${_exclude_file}" \
                    --password-file ${_password_file} \
                    ${_tag_options} \
                    ${_src_paths[${j}]}
                (( ${?} == 0 )) || return 1
            done
        fi
        echo "Backup check"
        restic check -v -r ${_DEST_REPOS[${i}]} --password-file ${_password_file}
        echo "Backup clean"
        if [[ "${_forget_options}" == "" ]]; then
            restic prune -r ${_DEST_REPOS[${i}]} --password-file ${_password_file}
            (( ${?} == 0 )) || return 1
        else
            restic forget -v -r ${_DEST_REPOS[${i}]} --password-file ${_password_file} ${_forget_options} --prune
            (( ${?} == 0 )) || return 1
        fi
        echo ""
    done
}

# ---------------------------------------------------------------------------
# Mount restic backup as regular file system for browsing / restore
# Arguments:
#   config filename / filepath
#   key name of mount point in configuration
#   restic password filepath
# Returns:
#   non-zero on error
# ---------------------------------------------------------------------------
restic_mount() {
    if [[ "${#}" -ne 3 ]]; then
        echo "[ERROR] Function ${FUNCNAME} usage error" >&2
        return 2
    fi

    local _config=${1}
    local _mount_point=${2}
    local _password_file=${3}

    read_mount_point ${_config} ${_mount_point}
    (( ${?} == 0 )) || return 1

    local _options
    _options=$(read_mount_options ${_config} ${_password_file})
    (( ${?} == 0 )) || return 1

    # echo "[DEBUG] Mount options: ${_options}"
    # echo "[DEBUG] restic -r ${_SRC_REPOS[0]} mount ${_DEST_REPOS[0]} ${_options}"
    restic -r ${_SRC_REPOS[0]} mount ${_DEST_REPOS[0]} ${_options}
    (( ${?} == 0 )) || return 1
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
            restic_mount ${_config} ${_mount_point} ${_password_file}
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
