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

INSTALLATION_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
CLUSTER_CONFIG="cluster_config.txt"

# setup Kafka
echo "Resetting Kafka"
cd $INSTALLATION_DIRECTORY/setup_kafka
./cleanup.sh
sleep 30

echo "Resetting prometheus"
cd $INSTALLATION_DIRECTORY/setup_prometheus
./cleanup.sh
sleep 30

echo "Resetting MinIO"
cd $INSTALLATION_DIRECTORY/setup_minio
./cleanup.sh
sleep 30

echo "Resetting Docker images"
cd $INSTALLATION_DIRECTORY/setup_docker
./cleanup.sh
sleep 30

