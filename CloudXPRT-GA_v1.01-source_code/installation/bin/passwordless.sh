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
  else
    echo "This is an unsupported OS. We cannot install the prerequites for this OS."
    exit
  fi
}


#############################################
# Get IP addresses from config file
#############################################
function getConfig {
  read -r -d ' ' -a IPs <<< $(jq -r '.nodes | .[].ip_address' $CLUSTER_CONFIG)

  for i in "${IPs[@]}"; do
    SERVERS+=("$i")
    server_count=$((server_count + 1))
  done
}


#############################################
# Create ssh keys on servers
#############################################
function createKeys {
  if [ ! -d '/root/.ssh' ]; then
    mkdir /root/.ssh
  fi
  if [ ! -f '/root/.ssh/id_rsa.pub' ]; then
    ssh-keygen -t rsa -N "" <<< ""
  fi
  if [ ! -f '/root/.ssh/authorized_keys' ]; then
    touch /root/.ssh/authorized_keys
  fi
  chmod 700 -R /root/.ssh
  cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys
}


#############################################
# Distribute keys and authorized_keys file
# to each server
#############################################
function distributeKeys {
  for server in "${SERVERS[@]}"; do
    sshpass -p "$PASSWORD" ssh -o "StrictHostKeyChecking no" root@"$server" '[[ -d "/root/.ssh" ]] && echo "Directory /root/.ssh/ exists." || mkdir /root/.ssh/; exit'
    sshpass -p "$PASSWORD" scp /root/.ssh/id_rsa root@"$server":/root/.ssh/id_rsa
    sshpass -p "$PASSWORD" scp /root/.ssh/id_rsa.pub root@"$server":/root/.ssh/id_rsa.pub
    sshpass -p "$PASSWORD" scp /root/.ssh/authorized_keys root@"$server":/root/.ssh/authorized_keys
    ssh root@"$server" 'chmod 700 -R /root/.ssh/; exit'
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
