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

masterNode=$(kubectl get nodes | grep "master" | awk '{print $1}')

echo ""
echo "Cleaning up mc-server"
echo "---------------------"
kubectl delete hpa mc-service
kubectl delete service mc-service
kubectl delete deployment mc-service

echo ""
echo "Cleaning up web-server"
echo "----------------------"
kubectl delete hpa web-service
kubectl delete service web-service
kubectl delete deployment web-service

echo ""
echo "Cleaning up db-server"
echo "---------------------"
kubectl delete deploy,svc db-service

echo ""
echo "Cleaning up user-server"
echo "-----------------------"
kubectl delete deploy,svc user-service

echo ""
echo "Cleaning up redis-server"
echo "------------------------"
kubectl delete deploy,svc redis-service
kubectl delete deploy,svc redis-crypt-service
kubectl delete deploy,svc redis-user-service

echo ""
echo "Cleaning up crypt-server"
echo "------------------------"
kubectl delete deploy,svc crypt-service

echo ""
echo "Cleaning up cassandra-server"
echo "----------------------------"
kubectl delete svc cassandra
kubectl delete statefulset cassandra
kubectl delete pvc cassandra-data-cassandra-0
kubectl delete pvc cassandra-data-cassandra-1
kubectl delete pvc cassandra-data-cassandra-2
kubectl delete pv cassandra-data-1
kubectl delete pv cassandra-data-2
kubectl delete pv cassandra-data-3
kubectl delete pod recycler-for-cassandra-data-1
kubectl delete pod recycler-for-cassandra-data-2
kubectl delete pod recycler-for-cassandra-data-3
kubectl label nodes $masterNode dbtype-
