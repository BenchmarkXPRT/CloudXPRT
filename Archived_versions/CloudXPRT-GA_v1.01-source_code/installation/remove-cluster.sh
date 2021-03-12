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

if ! [ $(id -u) = 0 ]; then
   echo "You must run this script as root!"
   exit 1
fi

printf "Are you sure that you want to remove the cluster? [y/n]: "
read response
case "$response" in
	[yY][eE][sS]|[yY])
		# reset master node environment
		./bin/reset-environment

		# ssh into each worker node and reset environment
		while read -r node; do
			echo "Resetting $node"
			echo "Removing Docker and its images..."
			ssh -n $node 'apt-get purge -y --allow-change-held-packages docker-ce docker-ce-cli'
			ssh -n $node 'rm -rf /var/lib/docker'
			ssh -n $node 'rm -rf /etc/docker'
			ssh -n $node 'rm -rf /var/lib/dockershim'
			ssh -n $node 'apt autoremove -y'
			echo "Resetting hostname"
			ssh -n $node 'cp /root/cnb/hostname.back /etc/hostname'
			ssh -n $node 'hostname $(cat /root/cnb/hostname.back)'
		done < <(jq -r '.nodes | .[1:] | .[].ip_address' $CLUSTER_CONFIG)

		echo "Kubernetes cluster has been removed"
		;;
	*)
		echo "Kubernetes cluster will not be removed"
		;;
esac
