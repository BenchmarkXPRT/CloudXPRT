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

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

# Create storage location for persistent volumes
[ -d /mnt/cnb-pv ] || mkdir /mnt/cnb-pv
[ -d /mnt/cnb-pv/kafka ] || mkdir /mnt/cnb-pv/kafka
[ -d /mnt/cnb-pv/zookeeper ] || mkdir /mnt/cnb-pv/zookeeper

# Create namespace
kubectl create namespace kafka

# Create storage classes and persistent volumes
kubectl apply -f ./storage

# Create kafka namespace
#kubectl apply -f 00-namespace.yml

# Deploy zookeeper
kubectl apply -f ./zookeeper

# Deploy kafka
kubectl apply -f ./kafka

# Setup kafka topics / partitions
