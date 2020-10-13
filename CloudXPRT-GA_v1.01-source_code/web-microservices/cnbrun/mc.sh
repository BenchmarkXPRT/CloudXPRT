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


###############################################################################
# This script runs the CNB MONTE CARLO workload. cnbrun binary will call this
# script. It should only be used after the Kubernetes cluster is up.
###############################################################################

CNB_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# Check that all nodes are ready
echo "Checking that all nodes are ready..."
nodes=$(kubectl get nodes | awk 'FNR > 1')
all_ready=true
num_cores=0
num_nodes=0
https=true

while read -r node; do
        node_name=$(echo $node | awk '{print $1}')
        node_status=$(echo $node | awk '{print $2}')
        (( num_nodes++ ))
        if [ ! $node_status = "Ready" ]; then
                echo "Node $node_name is not ready!"
                all_ready=false
        fi
        cores=$(kubectl describe node $node_name | grep cpu: | head -n 1 | awk '{print $2}')
        num_cores=$(($num_cores + $cores))
done <<< "$nodes"

if [ $all_ready = false ]; then
	exit 1
else
	echo "All nodes are ready!"
fi

echo "Number of nodes in cluster:" $num_nodes
echo "Total number of cores: $num_cores"

# Deploy metrics-server if not running
if [[ ! "$(kubectl get pods --all-namespaces | grep "metrics-server" | awk '{print $3}')" = "1/1" ]]; then
	echo "Deploying metrics-server"
	kubectl create -f $CNB_DIRECTORY/metrics-server/deploy/1.8+/
fi

echo ""
echo "***Running CNB***"
echo ""
echo "Deploying redis-server"
echo "----------------------"
kubectl create -f ./services/redis/

echo "Waiting for redis-service pod to be available..."

while [ ! "$(kubectl get pods | grep "redis-service" | awk '{print $2}')" = "1/1" ]
do
    sleep 2;
done

echo ""
echo "Deploying Cassandra DB"
echo "----------------------"
masterNode=$(kubectl get nodes | grep "master" | awk '{print $1}')
kubectl label nodes $masterNode dbtype=cassandra
echo "Label master node with dbtype=cassandra"
kubectl create -f ./cassandra/onprem/cassandra-storage.yaml
echo "Creating persistent volume for cassandra..."
sleep 3

kubectl create -f ./cassandra/onprem/cassandra-deploy.yaml
echo "Waiting for Cassandra replicas to be available..."

while [ ! "$(kubectl get pods | grep "cassandra-" | grep 'Running' | wc -l)" -ge 3 ]
do
  sleep 2;
done

# Wait for Cassandra replicas to stabilize
sleep 20
echo "Create schema for Cassandra DB"
kubectl cp ./cassandra/schema.cql cassandra-0:/root
kubectl exec -i cassandra-0 -- cqlsh -f /root/schema.cql
sleep 1

echo ""
echo "Deploying db-server"
echo "-------------------"
kubectl create -f ./services/db-service.yml

echo "Waiting for db-service pod to be available..."

while [ ! "$(kubectl get pods | grep "db-service" | awk '{print $2}')" = "1/1" ]
do
    sleep 2;
done

echo ""
echo "Deploying user-server"
echo "---------------------"
kubectl create -f ./services/user-service.yml

echo "Waiting for user-service pod to be available..."

while [ ! "$(kubectl get pods | grep "^user-service" | awk '{print $2}')" = "1/1" ]
do
    sleep 2;
done

echo ""
echo "Deploying crypt-server"
echo "----------------------"
kubectl create -f ./services/crypt-service.yml

echo "Waiting for crypt-service pod to be available..."

while [ ! "$(kubectl get pods | grep "^crypt-service" | awk '{print $2}')" = "1/1" ]
do
    sleep 2;
done

echo ""
echo "Deploying web-server"
echo "--------------------"
kubectl create -f ./services/web-service.yml

echo "Waiting for web-server pod to be available..."

while [ ! "$(kubectl get pods | grep "web-service" | grep 'Running' | wc -l)" -ge 1 ]
do
	sleep 2;
done

web_service_ip=$(kubectl get services | grep "web-service" | awk '{print $3}')
echo "IP address for web-service $web_service_ip"
echo ""
echo "Deploying HPA for web-server"
echo "----------------------------"
kubectl autoscale deployment web-service --cpu-percent=75 --min=1 --max=5

#########################
# MC Workload
#########################
echo ""
echo "Deploying mc-server pod"
echo "-----------------------"
cp ./services/mc-service-template.yml ./services/mc-service.yml
sed -i "s/{{IMAGE_VERSION}}/$1/" ./services/mc-service.yml
sed -i "s/{{CPU_REQUESTS}}/$2/" ./services/mc-service.yml
sed -i "s/{{THREADS_NUM}}/$3/" ./services/mc-service.yml
kubectl create -f ./services/mc-service.yml
sleep 2

echo "Waiting for mc-server pod to be available..."

while [ ! "$(kubectl get pods | grep "mc-service" | grep 'Running' | wc -l)" -ge 1 ]
do
	sleep 2;
done

sleep 6
echo ""
echo "Testing mc-server"
echo "-----------------"
if [ $https = false ]; then
        curl --silent --noproxy "*" http://$web_service_ip:8070/mc
else
        curl --silent --noproxy "*" -k https://$web_service_ip:8443/mc
fi
echo ""

# Leave resources for webserver pod(s) and kube-system pods based on number of cores per MC pod
if [ "$3" -eq 1 ]
then
	max_mc_pods=$((( $num_cores  / $3  ) - ( 4 * $num_nodes ) ))
elif [ "$3" -ge 3 ]
then
	max_mc_pods=$((( $num_cores  / $3  ) - $num_nodes ))
else
	max_mc_pods=$((( $num_cores  / $3  ) - ( 2 * $num_nodes ) ))
fi

echo "Max number of MC Pods:" $max_mc_pods

min_mc_pods=$(( $max_mc_pods / 3 ))
if [ "${@: -2:1}" = "enablehpa" ]
then
	echo "Deploying HPA for mc-server"
	echo "---------------------------"
	kubectl autoscale deployment mc-service --cpu-percent=75 --min=$min_mc_pods --max=$max_mc_pods
else
	echo "Deploying max pods for mc-server"
	echo "--------------------------------"
	kubectl scale deployment/mc-service --replicas=$max_mc_pods
fi

echo "Waiting until all mc-server pods are up and running..."

if [ "${@: -2:1}" = "enablehpa" ]
then
	while [ ! "$(kubectl get deploy mc-service | awk 'FNR > 1 {print $4}')" = "$min_mc_pods" ]
	do
		sleep 5;
	done
else
	while [ ! "$(kubectl get deploy mc-service | awk 'FNR > 1 {print $4}')" = "$max_mc_pods" ]
	do
		sleep 5;
	done
fi

echo "mc-service: $(kubectl get pods | grep "mc-service" | grep "Running" | wc -l) pods are created!"

echo ""
echo "Ready to go!"

echo ""
echo "############################"
echo "Running Monte Carlo workload"
echo "############################"
if [ $https = false ]; then
        ./autoloader -u http://$web_service_ip:8070/mc -c $4 -ci $5 -cl $6 -s $7 -ti $8 -e Monte
else
        ./autoloader -u https://$web_service_ip:8443/mc -c $4 -ci $5 -cl $6 -s $7 -ti $8 -e Monte
fi

echo ""
echo "Total number of pods created during the run"
echo "-------------------------------------------"
echo "web-service pods: $(kubectl get pods | grep "web-service" | grep "Running" | wc -l)"
echo "mc-service pods: $(kubectl get pods | grep "mc-service" | grep "Running" | wc -l)"

if [ "${@: -1}" = "needclean" ]
then
	echo ""
	echo "Cleaning up mc-server"
	echo "---------------------"
	if [ "${@: -2:1}" = "enablehpa" ]
	then
		kubectl delete hpa mc-service
	fi
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
fi
