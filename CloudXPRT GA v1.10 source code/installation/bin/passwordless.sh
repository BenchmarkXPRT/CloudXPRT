#!/bin/bash
#===============================================================================
# Copyright 2020 BenchmarkXPRT Development Community
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#===============================================================================

##############################################################################
#  passwordless.sh - This script handles passwordless setup for new servers
##############################################################################

server_count=0
IP_REGEX='([1-9]?[0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])'
cd $INSTALLATION_DIRECTORY

#############################################
# Allow user to predefine the server(s)
# password and config file
#############################################
if [[ "$PASSWORD" == "" ]]; then
    echo
    echo -n "What is the root password for all machines? "
    read -s PASSWORD
    echo
fi

#############################################
# Install required sshpass
#############################################
function prerequisites {
    if hash apt-get &>/dev/null; then
        sudo -E apt-get update
        sudo -E apt-get upgrade -y
        sudo -E apt-get install build-essential -y
        sudo -E apt-get install sshpass -y
    elif hash yum &>/dev/null; then
        sudo -E yum check-update
        sudo -E yum update -y
        sudo yum -y install sshpass
    else
        echo "This is an unsupported OS. We cannot install the prerequites for this OS."
        exit
    fi
}

#############################################
# Get IP addresses from config file
#############################################
function getConfig {
    # Only work on Ubuntu, not working on CentOS!
    # read -r -d ' ' -a IPs <<< $(jq -r '.nodes | .[].ip_address' $CLUSTER_CONFIG)

    # Parse config file only if IPs is empty
    if [ "${#IPs[@]}" -eq 0 ]; then
        while read -r value; do
            IPs+=("$value")
        done < <(jq -r '.nodes | .[].ip_address' $CLUSTER_CONFIG)
    fi

    for i in "${IPs[@]}"; do
        SERVERS+=("$i")
        server_count=$((server_count + 1))
    done
}

#############################################
# Create ssh keys on servers
#############################################
function createKeys {
    sudo -u "$SUDO_USER" ssh-keygen -t rsa -N "" <<<""
}

#############################################
# Distribute keys and authorized_keys file
# to each server
#############################################
function distributeKeys {
    for server in "${SERVERS[@]}"; do
        sudo -u "$SUDO_USER" sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$SUDO_USER"@"$server" 'exit'
        sudo -u "$SUDO_USER" sshpass -p "$PASSWORD" ssh-copy-id "$SUDO_USER"@"$server"
    done
}

#############################################
# Setup Passwordless on all machines
#############################################
function setupPasswordless {
    if hash sshpass; then
        echo "'sshpass' is installed. Continuing..."
    else
        prerequisites
    fi
    if [[ $SERVERS == "" ]]; then
        SERVERS=()
        getConfig
    fi
    createKeys
    distributeKeys
}

setupPasswordless
