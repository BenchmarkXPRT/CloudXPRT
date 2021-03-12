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

###############################################################################
# Script to setup Minio
###############################################################################

apt update
apt -y install python3-pip
pip3 install -U minio==7.0.1

mkdir -p /mnt/minio

kubectl apply -f single/storageclass.yaml
while [ ! "$(kubectl get storageclass | grep minio-data-storage | awk '{print $2}')" = "kubernetes.io/no-provisioner" ]; do
    sleep 2
done

kubectl apply -f single/pv.yaml

while [ ! "$(kubectl get pv/miniodb | grep miniodb | awk '{print $5}')" = "Available" ]; do
    echo "miniodb volume not available.. it may take a few minutes"
    sleep 2
done

kubectl apply -f single/pvc.yaml

sleep 30
#while [ ! "$(kubectl get pvc/minio-pv-claim | grep minio-pv-claim | awk '{print $2}')" = "Bound" ]; #do
#    echo "minio-pv-claim not available.. it may take a few minutes"
#    sleep 2
#done

kubectl apply -f single/singledeploy.yaml
while [ ! "$(kubectl get pods | grep minio | awk '{print $2}')" = "1/1" ]; do
    echo "waiting for minio pod.. it may take a few minutes"
    sleep 10
done

minio_pod_name=$(kubectl get pods | grep minio | awk '{print $1}')

kubectl apply -f single/singleservice.yaml
while [ ! "$(kubectl get services | grep minio-service | awk '{print $5}')" = "9000/TCP" ]; do
    echo "waiting for minio-service.. it may take a few minutes"
    sleep 2
done

minio_service_ip=$(kubectl get services | grep minio-service | awk '{print $3}')
minio_service_ip=$minio_service_ip:9000
minioEndpoint=$(kubectl logs $minio_pod_name | grep Endpoint: | awk '{print $2}')

echo "minio_pod_name:$minio_pod_name"
echo "minio_service_ip:$minio_service_ip"
echo "minioEndpoint:$minioEndpoint"

#download datasets
if [[ ! -d "data" ]]; then
    mkdir data
    cd data
    wget https://archive.ics.uci.edu/ml/machine-learning-databases/00280/HIGGS.csv.gz
    chmod 755 ./*
    cd ../
fi
if [[ ! -f data/HIGGS.csv.gz ]]; then
    echo 'File "data/HIGGS.csv.gz" is not there, aborting.'
    echo 'wget did not work ! if you are behind a proxy add proxy into /etc/wgetrc'
    echo 'wget https://archive.ics.uci.edu/ml/machine-learning-databases/00280/HIGGS.csv.gz'
    exit
fi

chmod -R 755 data/*

#setup a bucket and add content to it
#call the following script with the minio service ip. Edit "createbucketAddcontent.py" to create a bucket add required type of data
./createbucketAddcontent.py $minio_service_ip cnb-ml-bucket
