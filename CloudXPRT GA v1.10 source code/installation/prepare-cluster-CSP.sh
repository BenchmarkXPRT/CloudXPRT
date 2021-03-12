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

################################################################################
# This script gathers cluster IP addresses and sets up the cluster environment
# for CloudXPRT
################################################################################

INSTALLATION_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
CLUSTER_CONFIG="cluster_config.json"

if [ -z "$SUDO_USER" ]; then
    echo "You must run this script as sudo user!"
    exit 1
fi

# Regex for IP validity
IP_REGEX='([1-9]?[0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])'

# Validate that the given IP address is a valid IPV4 address format and can be pinged
function validateIP {
    IP_ADDRESS=$1
    if [[ $IP_ADDRESS =~ ^$IP_REGEX\.$IP_REGEX\.$IP_REGEX\.$IP_REGEX$ ]]; then
        if ping -c 1 -w 1 -q $IP_ADDRESS >/dev/null 2>&1; then
            return 0
        else
            echo "Warning: Can't ping $IP_ADDRESS..."
            printf "Are you sure the IP address is correct? [y/n]: "
            read response
            case "$response" in
            [yY][eE][sS] | [yY])
                return 0
                ;;
            *)
                return 1
                ;;
            esac
        fi
    else
        echo "$IP_ADDRESS is not a valid IP address. Please try again."
        return 1
    fi
}

function validateClusterConfig {
    # Simple (not strict) JSON validation
    jq empty $CLUSTER_CONFIG
    rc="$?"

    if [ "$rc" -ne 0 ]; then
        exit
    fi

    # Get IP addresses and hostnames from config file
    NUM_NODES=$(jq '.nodes | length' $CLUSTER_CONFIG)

    while read -r value; do
        if [ -n "$value" ]; then
            IPs+=("$value")
        fi
    done < <(jq -r '.nodes | .[].ip_address' $CLUSTER_CONFIG)

    while read -r value; do
        if [ -n "$value" ]; then
            HOSTNAMES+=("$value")
        fi
    done < <(jq -r '.nodes | .[].hostname' $CLUSTER_CONFIG)

    if [ "${#IPs[@]}" -eq 0 ]; then
        echo "Please edit cluster_config.json with node IPs, hostnames, and proxy info (if applicable)"
        exit
    fi

    # Validate IP addresses
    for ip in ${IPs[@]}; do
        if (! validateIP $ip); then
            echo "Invalid IP address $ip in $CLUSTER_CONFIG!"
            exit
        fi
        sudo -u "$SUDO_USER" ssh -o StrictHostKeyChecking=no "$SUDO_USER"@"$ip" 'exit'
    done

    if [ "${#HOSTNAMES[@]}" -ne 0 ]; then
        if [ "${#IPs[@]}" -ne "${#HOSTNAMES[@]}" ]; then
            echo "Number of hostnames provided is not equal to the number of IPs provided"
            exit
        fi

        # Set inventory to have both IP addresses and hostnames
        INVENTORY=()
        for ((i = 0; i < "$NUM_NODES"; i++)); do
            INVENTORY+="${HOSTNAMES[i]},${IPs[i]} "
        done
    else
        # Set inventory to only have IP addresses
        INVENTORY=${IPs[@]}
    fi

    # Get proxy information
    SET_PROXY=$(jq -r '.proxy.set_proxy' $CLUSTER_CONFIG)

    if [ $SET_PROXY == "yes" ]; then
        read -r HTTP_PROXY <<<$(jq -r '.proxy.http_proxy' $CLUSTER_CONFIG)
        read -r HTTPS_PROXY <<<$(jq -r '.proxy.https_proxy' $CLUSTER_CONFIG)
        read -r REBOOT <<<$(jq -r '.proxy.reboot' $CLUSTER_CONFIG)
    fi
}

# Ensure dependencies are installed
if hash apt-get &>/dev/null; then
    apt-get update
    apt-get install python3 python3-pip python3-testresources openssh-server jq -y
elif hash yum &>/dev/null; then
    yum install epel-release -y
    yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm -y
    yum install python3-pip python3-testresources openssh-server jq -y
fi

# Get cluster configuration
echo "Checking configuration in cluster_config.json"
validateClusterConfig

# setup environments for each node (date, proxies, etc.)
cd $INSTALLATION_DIRECTORY

# SSH into each worker node to setup proxies and reboot if needed
while read -r node; do
    echo "Working on node with IP address: $node"
    sudo -u "$SUDO_USER" ssh -n $node "mkdir cnb"
    sudo -u "$SUDO_USER" scp bin/setup-environment $node:~/cnb/setup-environment

    if [ "$SET_PROXY" = "yes" ]; then
        if [ "$REBOOT" = "yes" ]; then
            sudo -u "$SUDO_USER" ssh -n $node "cd cnb && sudo ./setup-environment -nodetype worker -http_proxy $HTTP_PROXY -https_proxy $HTTPS_PROXY -reboot"
        else
            sudo -u "$SUDO_USER" ssh -n $node "cd cnb && sudo ./setup-environment -nodetype worker -http_proxy $HTTP_PROXY -https_proxy $HTTPS_PROXY"
        fi
    else
        sudo -u "$SUDO_USER" ssh -n $node "cd cnb && sudo ./setup-environment -nodetype worker -noproxy"
    fi
done < <(jq -r '.nodes | .[1:] | .[].ip_address' $CLUSTER_CONFIG)

# Control Plane node setup
if [ "$SET_PROXY" = "yes" ]; then
    ./bin/setup-environment -nodetype master -http_proxy $HTTP_PROXY -https_proxy $HTTPS_PROXY

    if [ "$REBOOT" = "no" ]; then
        echo "Please reboot all nodes manually for proxy changes to take effect before running CNB create-cluster.sh!"
        echo "Otherwise creating the cluster will fail!"
    fi
else
    ./bin/setup-environment -nodetype master -noproxy
fi

# Setup ansible/kubespray requirements
cd $INSTALLATION_DIRECTORY/kubespray
pip3 install -r requirements.txt
pip3 install --upgrade setuptools

# Setup cluster hosts file
CONFIG_FILE=inventory/cnb-cluster/inventory.yaml python3 contrib/inventory_builder/inventory.py ${INVENTORY[@]}

if [ "$REBOOT" = "yes" ]; then
    for i in {10..1..1}; do
        echo "System will be rebooted in $i seconds"
        sleep 1
    done
    reboot now
fi
