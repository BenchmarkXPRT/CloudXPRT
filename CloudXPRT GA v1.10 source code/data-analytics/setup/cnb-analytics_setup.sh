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

if [ "$EUID" -ne 0 ]; then
    echo "Please run as sudo user"
    exit
fi

apt install -y bc moreutils jq

INSTALLATION_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# setup docker
cd $INSTALLATION_DIRECTORY/setup_docker
./setup.sh
sleep 30
if [ "$(docker images | awk 'FNR > 1 {print $1}' | grep -c xgboost)" == '1' ]; then
    echo -e "xgboost image loaded successfully !\n\n"
fi

# setup Kafka
# If single node cluster
if [ "$(kubectl get node --no-headers -o custom-columns=NAME:.metadata.name | wc -l)" == '1' ]; then
    echo -e "Single node setup detected... start Kafka configuration\n\n"

    #remove content from last line
    kafka_storage_file=$INSTALLATION_DIRECTORY/setup_kafka/storage/kafka-storage.yml
    zookeeper_storage_file=$INSTALLATION_DIRECTORY/setup_kafka/storage/zookeeper-storage.yml
    if [ ! -f $kafka_storage_file'.bkp' ]; then
        echo -e "creating bkp...\n"
        cp $kafka_storage_file $kafka_storage_file'.bkp'
        cp $zookeeper_storage_file $zookeeper_storage_file'.bkp'
    fi
    cp $kafka_storage_file'.bkp' $kafka_storage_file
    cp $zookeeper_storage_file'.bkp' $zookeeper_storage_file
    head -n -1 $kafka_storage_file >temp.txt ; mv temp.txt $kafka_storage_file
    head -n -1 $zookeeper_storage_file >temp.txt ; mv temp.txt $zookeeper_storage_file

    # configure kafka storage file
    master_node=$(kubectl get nodes --selector='node-role.kubernetes.io/master' --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')
    echo "          - $master_node" >>$kafka_storage_file
    echo "          - $master_node" >>$zookeeper_storage_file
fi

cd $INSTALLATION_DIRECTORY/setup_kafka
./setup.sh
sleep 30
if [ "$(kubectl get service -n kafka | awk 'FNR > 1 {print $1}' | grep -c bootstrap)" == '1' ]; then
    echo -e "kafka bootstrap loaded successfully !\n\n"
fi

# setup prometheus
cd $INSTALLATION_DIRECTORY/setup_prometheus
./setup.sh
sleep 30
if [ "$(kubectl get service -n monitoring | awk 'FNR > 1 {print $1}' | grep -c prometheus)" == '1' ]; then
    echo -e "prometheus loaded successfully !\n\n"
fi

# setup minio
cd $INSTALLATION_DIRECTORY/setup_minio
./setup.sh
sleep 30
if [ "$(kubectl get pods | awk 'FNR > 1 {print $1}' | grep -c minio)" == '1' ]; then
    echo -e "minio loaded successfully !\n\n"
fi

# final check
sleep 30
if [ "$(kubectl get pod/kafka-0 -n kafka | grep kafka-0 | awk '{print $3}')" = "Running" ]; then
    echo -e "\n\n cnb-ml loaded successfully ! \n\n"
fi
