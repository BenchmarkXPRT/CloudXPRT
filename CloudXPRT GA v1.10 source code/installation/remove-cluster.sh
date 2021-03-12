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

INSTALLATION_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
CLUSTER_CONFIG="cluster_config.json"
ANSIBLE_PATH=/usr/local/bin

if [ -z "$SUDO_USER" ]; then
    echo "You must run this script as sudo user!"
    exit 1
fi

printf "Are you sure that you want to remove the cluster? [y/n]: "
read response
case "$response" in
[yY][eE][sS] | [yY])
    # remove cluster
    cd $INSTALLATION_DIRECTORY/kubespray
    sudo -u "$SUDO_USER" $ANSIBLE_PATH/ansible-playbook -i inventory/cnb-cluster/inventory.yaml --become --become-user=root reset.yml --extra-vars "reset_confirmation=yes"

    # reset control-plane node environment
    cd $INSTALLATION_DIRECTORY
    ./bin/reset-environment

    # ssh into each worker node and reset environment
    while read -r node; do
        echo "Resetting $node"
        echo "Removing Docker and its images..."
        if hash apt-get &>/dev/null; then
            sudo -u "$SUDO_USER" ssh -n $node 'sudo apt-get purge -y --allow-change-held-packages docker-ce docker-ce-cli'
            sudo -u "$SUDO_USER" ssh -n $node 'sudo apt autoremove -y'
        elif hash yum &>/dev/null; then
            sudo -u "$SUDO_USER" ssh -n $node 'sudo yum remove -y docker-ce docker-ce-cli'
        fi
        sudo -u "$SUDO_USER" ssh -n $node 'sudo rm -rf /var/lib/docker /etc/docker'
        echo "Resetting hostname"
        sudo -u "$SUDO_USER" ssh -n $node 'sudo cp cnb/hostname.back /etc/hostname'
        sudo -u "$SUDO_USER" ssh -n $node 'sudo hostname $(cat cnb/hostname.back)'
    done < <(jq -r '.nodes | .[1:] | .[].ip_address' $CLUSTER_CONFIG)

    echo "Kubernetes cluster has been removed"
    ;;
*)
    echo "Kubernetes cluster will not be removed"
    ;;
esac
