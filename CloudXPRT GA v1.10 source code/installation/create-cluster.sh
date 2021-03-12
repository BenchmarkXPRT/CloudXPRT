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
# This script creates the cluster for CloudXPRT
################################################################################

INSTALLATION_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
CLUSTER_CONFIG="$INSTALLATION_DIRECTORY/cluster_config.json"
ANSIBLE_PATH=/usr/local/bin

if hash yum &>/dev/null; then
    PATH=$PATH:/usr/local/bin
fi

if [ -z "$SUDO_USER" ]; then
    echo "You must run this script as sudo user!"
    exit 1
fi

# Ensure date and time are current
date -s "$(wget -qSO- --max-redirect=0 google.com 2>&1 | grep Date: | cut -d' ' -f5-8)Z"

DIR=$HOME
if [ -n "$SUDO_USER" ]; then
    DIR=/home/$SUDO_USER
fi

# Create Cluster
cd $INSTALLATION_DIRECTORY/kubespray
sudo -u "$SUDO_USER" $ANSIBLE_PATH/ansible-playbook -i inventory/cnb-cluster/inventory.yaml --become --become-user=root cluster.yml

NUM_NODES=$(jq '.nodes | length' $CLUSTER_CONFIG)

mkdir -p $DIR/.kube
cp /etc/kubernetes/admin.conf $DIR/.kube/config
if [ -n "$SUDO_USER" ]; then
    chown -R $SUDO_USER $DIR/.kube
fi

cd $INSTALLATION_DIRECTORY

if [ "$(kubectl get nodes | awk 'FNR > 1' | awk '{print $2}' | grep -c Ready)" == $NUM_NODES ]; then
    echo "Installation completed successfully!"
else
    echo "Installation failed!"
fi
