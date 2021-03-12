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

lscpu
lshw -short | grep -i "system memory"
lshw -html >system_info.html
lsblk
df -h

echo "==> Docker Images:"
docker images
echo ""

echo "==> Nodes:"
kubectl get nodes -o wide
echo ""

echo "==> Reources system wide:"
kubectl get all --all-namespaces -o wide
echo ""

echo "==> Resources on kafka namespace:"
kubectl get all -n kafka -o wide
echo ""

echo "==> All pods:"
kubectl get pod -o=custom-columns=NAME:.metadata.name,STATUS:.status.phase,NODE:.spec.nodeName --all-namespaces
echo ""

echo "==> PV:"
kubectl get pv
kubectl get pvc -n kafka
echo ""

echo "==> Events:"
kubectl get events -n kafka
echo ""

echo "==> Description cdn-webserver:"
kubectl describe pod -n kafka cdn-webserver
kubectl describe nodes node1
kubectl describe pod -n kafka xgb*
echo ""

echo "==> LOG cdn-webserver"
kubectl logs -n kafka cdn-webserver
echo ""

echo "==> Transactions interim counters"
kubectl logs -n kafka cdn-webserver | grep -c Received
kubectl logs -n kafka cdn-webserver | grep -c transaction
kubectl logs -n kafka cdn-webserver | grep -c Input
kubectl logs -n kafka cdn-webserver | grep -c Fail
echo ""

echo "==> CPU and netstat"
netstat -ant
lscpu
echo ""
