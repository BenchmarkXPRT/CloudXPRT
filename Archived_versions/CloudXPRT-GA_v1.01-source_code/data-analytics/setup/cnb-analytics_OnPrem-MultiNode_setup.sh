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

#########################################################
# This script runs the setup for CNB-analytics multi-node
#########################################################

INSTALLATION_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
AWS_CONFIG_FOLDER="setup_csp-aws"

# asumption is that csp_aws contains required files with pertinent changes
# check directory exists
if [ ! ~/.ssh/id_rsa ]; then
  echo "the ~/.ssh/id_rsa file does not exist !"
  exit;
fi

# copy xgboost.sh
#cp $AWS_CONFIG_FOLDER/xgboost.sh $INSTALLATION_DIRECTORY/../cnbrun/

#remove content from last line
kafka_storage_file=$INSTALLATION_DIRECTORY/setup_kafka/storage/kafka-storage.yml
zookeeper_storage_file=$INSTALLATION_DIRECTORY/setup_kafka/storage/zookeeper-storage.yml
head -n -1 $kafka_storage_file > temp.txt ; mv temp.txt $kafka_storage_file
head -n -1 $zookeeper_storage_file > temp.txt ; mv temp.txt $zookeeper_storage_file

# ssh into working nodes
working_nodes=$(kubectl get nodes --selector='!node-role.kubernetes.io/master' --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')
for node_name in $working_nodes
do
  node_ip=$(kubectl get nodes --field-selector metadata.name=$node_name -o wide | grep $node_name |  awk '{print $6}')
  echo $node_name " " $node_ip
  echo "          - "$node_name >> $kafka_storage_file
  echo "          - "$node_name >> $zookeeper_storage_file
  kubectl label node $node_name node-role.kubernetes.io/node=node
  ssh root@${node_ip} << EOF
   sudo mkdir /mnt/cnb-pv
   sudo mkdir /mnt/cnb-pv/kafka
   sudo mkdir /mnt/cnb-pv/zookeeper
   sudo docker pull cloudxprt/xgboost:v1.00
   sudo docker pull cloudxprt/cdn-webserver:v1.00
   sudo docker images
EOF
done
