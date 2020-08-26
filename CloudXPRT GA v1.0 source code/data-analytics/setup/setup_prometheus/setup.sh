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

# Create namespace
kubectl create namespace monitoring

# Create ClusterRole and binding
kubectl apply -f clusterRole.yaml

# Create Config map
kubectl create -f config-map.yaml

# Deploy prometheus
kubectl create  -f prometheus-deployment.yaml


# Wait until prometheus pod is available
while [ ! "$(kubectl get pods -n monitoring| grep "prometheus-deployment-[a-z 0-9 -]" | awk '{print $2}')" = "1/1" ]
do
    sleep 5;
done

# Deploy node-exporter
#kubectl create  -f node-exporter-deployment.yaml


#Install helm to add node-exporter. Till we can find a better way of adding node-exporter
tar -zxvf helm-v2.16.0-linux-amd64.tar.gz

mv linux-amd64/helm /usr/local/bin/helm
rm -rf linux-amd64

#Create service account and clusterRoleBinding for Tiller
kubectl apply -f helm-rbac.yaml

echo "waiting for Tiller Pod in monitoring namespace... waiting for 30 secs"
sleep 30

#Initialize
helm init --service-account=tiller --history-max 300

echo "waiting for Tiller Pod in monitoring namespace... waiting for 30 secs"
sleep 30

#Install node-exporter
helm install --name node-exporter --namespace monitoring stable/prometheus-node-exporter

#install kube-state-metrics
helm install --name kube-state-metrics --namespace kube-system stable/kube-state-metrics

#need to check

#setup port forwarding to connect and get sytem data after workloads are run
#prom_deploy_pod_name=$(kubectl get pods -n monitoring | grep "prometheus-deployment-[a-z 0-9 -]"  | awk '{print $1}')
#echo "Pod name for prometheus prom_deploy_pod_name"
#kubectl port-forward $prom_deploy_pod_name 8080:9090 -n monitoring
