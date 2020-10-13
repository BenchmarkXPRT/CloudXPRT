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

echo ""
echo "Cleaning up pods and services"
echo "-----------------------------"

xgbpods=$(kubectl get pods --selector=app=xgb -n kafka | awk 'FNR >= 2 {print $1}')
while read -r podname; do
	echo "deleting "$podname
	kubectl delete pod/$podname -n kafka
done <<< "$xgbpods"

kubectl delete service cdn-webservice -n kafka
kubectl delete pod/cdn-webserver -n kafka
kubectl delete service prometheus-service -n monitoring

kubectl delete pvc/task-pv-claim -n kafka
kubectl delete pv/task-pv-volume -n kafka
	


