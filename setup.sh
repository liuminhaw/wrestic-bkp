
#!/bin/bash
#
# restic-bkp setup script
#
# Exit Code:
#   1 - Calling syntax error
#   3 - Destination directory does not exist
#
#   11 - Copy file failed
#   13 - Change file permission failed
#   15 - Make directory failed
#   17 - Download file failed
#   19 - Pre-requisite not met

# ----------------------------------------------------------------------------
# Show script usage
# Outputs:
# 	Write usage information to stdout
# ----------------------------------------------------------------------------
show_help() {
cat << EOF
Usage: ${0##*/} [--help] DESTINATION
    --help                      Display this help message and exit
EOF
}

# ------------------------------------------------
# Check command return code for script termination
# Globals:
#   ${?}
# Outputs:
#   Write excution error message to stderr
# ------------------------------------------------
checkCode() {
if [[ ${?} -ne 0 ]]; then
    echo ${2} 1>&2
    exit ${1}
fi
}

# ----------------------------------
# Loop check for 'y' or 'n' response
# Arguments:
#   None
# Returns:
#   0 if no
#   1 if yes
# ----------------------------------
yesNoCheck() {
    while true; do
        read _response_check
        if [[ "${_response_check,,}" == "y" ]]; then
            return 1
        elif [[ "${_response_check,,}" == "n" ]]; then
            return 0
        else
            printf "Please enter 'y' for yes and 'n' for no: "
            continue
        fi
    done
}


# ----------------------------------------------------------
# Install execution
# Arguments:
#   install destination path
# -----------------------------------------------------------
Installation() {
    if [[ "${#}" -ne 1 ]]; then
        echo "[ERROR] Function Installation usage error"
        exit 5
    fi

    local _dest_dir=${1}

    # Pre-requisite test
    which jq > /dev/null
    checkCode 19 "Pre-requisite jq command not found" > /dev/null

    # Version compatible check (0.0.X -> 0.1.Y)
    if [[ -f "${_dest_dir}/restic-backup.sh" ]]; then
        local _new_version=$(./restic-bkp.sh --version)
        local _cur_version=$(${_dest_dir}/restic-backup.sh --version) 
        echo ${_cur_version} | grep "Version: 0.0.*" > /dev/null
        if [[ "${?}" -eq 0 ]]; then
            echo "--------------------------------------------------------------------------"
            echo "Found current version of restic-bkp -> ${_cur_version},"
            echo "new version -> ${_new_version} of restic-bkp will be installed."
            echo "After installation, origin config file should be migrate to use new config json format (config.json) due to version incompatible."
            echo "JSON format content of config file can be referenced from config.template."
            echo "--------------------------------------------------------------------------"
            printf "Confirm to upgrade? (y/n): "
            yesNoCheck
            if [[ ${?} -eq 0 ]]; then # yes
                echo "Setup exit"
                exit 0
            fi

            echo "--------------------------------------------------------"
            echo "restic-backup.sh has renamed to restic-bkp.sh"
            echo "Do you want to overwrite current restic-backup.sh script?"
            echo "--------------------------------------------------------"
            printf "Overwrite current restic-backup.sh script? (y/n): "
            yesNoCheck
            if [[ ${?} -eq 1 ]]; then
                mv ${_dest_dir}/restic-backup.sh ${_dest_dir}/restic-bkp.sh
                checkCode 11 "Rename restic-backup.sh failed." > /dev/null
            fi
        fi
    fi

    # Setup process
    cp README.md ${_dest_dir}
    checkCode 11 "Copy README.md failed." > /dev/null

    cp restic-bkp.sh ${_dest_dir}
    checkCode 11 "Copy restic-bkp.sh failed." > /dev/null

    cp config.template ${_dest_dir}/config.template
    checkCode 11 "Copy config.template failed." > /dev/null

    if [[ ! -f "${_dest_dir}/config.json" ]]; then
        cp config.template ${_dest_dir}/config.json
        checkCode 11 "Copy config.template failed." > /dev/null
    fi

    if [[ ! -f "${_dest_dir}/.restic.pass" ]]; then
        cp .restic.pass ${_dest_dir}/.restic.pass
        checkCode 11 "Copy .restic.pass failed." > /dev/null
    fi

    if [[ ! -f "${_dest_dir}/excludes.txt" ]]; then
        cp excludes.txt ${_dest_dir}/excludes.txt
        checkCode 11 "Copy excludes.txt failed." > /dev/null
    fi
}


# Calling setup format check
if [[ ${#} -ne 1 ]];  then
    show_help
    exit 1
fi

while :; do
    case ${1} in
        --help)
            show_help
            exit 1
            ;;
        -?*)
            echo -e "[WARN] Unknown option (ignored): ${1}" 1>&2
            ;;
        *)  # Default case: no more options
            break
    esac

    shift
done

if [[ ! -d ${1} ]]; then
    echo "[ERROR] Destination directory does not exist"
    exit 3
fi


# System checking
SYSTEM_RELEASE=$(uname -a)
case ${SYSTEM_RELEASE} in
*Linux*)
    echo "Linux detected"
    echo ""
    Installation ${1}
    ;;
*)
    echo "OS Not supported."
    exit 1
esac

echo "restic-bkp setup success."
exit 0