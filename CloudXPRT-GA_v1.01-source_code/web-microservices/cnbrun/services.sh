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
# This script can bring up and remove all CNB services during a normal run. For
# simplicity, the replicas for the MonteCarlo service is limited to one.
# This script should only be used after the Kubernetes cluster is up.
#
# usage ./services.sh up
# usage ./services.sh down
#
###############################################################################

CNB_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

if ! [ $(id -u) = 0 ]; then
   echo "You must run this script as root!"
   exit 1
fi

function bringUpServices {
	echo ""
	echo "***Deploying all services***"
	echo ""

	# Check that all nodes are ready
	echo "Checking that all nodes are ready..."
	nodes=$(kubectl get nodes | awk 'FNR > 1')
	all_ready=true
	num_nodes=0

	while read -r node; do
		node_name=$(echo $node | awk '{print $1}')
		node_status=$(echo $node | awk '{print $2}')
		(( num_nodes++ ))
		if [ ! $node_status = "Ready" ]; then
			echo "Node $node_name is not ready!"
			all_ready=false
		fi
		cores=$(kubectl describe node $node_name | grep cpu: | head -n 1 | awk '{print $2}')
	done <<< "$nodes"

	echo "Number of nodes in cluster:" $num_nodes

	if [ $all_ready = false ]; then
		exit 1
	else
		echo "All nodes are ready!"
	fi

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

	# wait for all Cassandra replicas to stabilize
	sleep 75

	echo "Create schema for Cassandra DB"
	kubectl cp ./cassandra/schema.cql cassandra-0:/root
	kubectl exec -i cassandra-0 -- cqlsh -f /root/schema.cql

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

	echo ""
	echo "Deploying mc-server pod"
	echo "-----------------------"
	kubectl create -f ./services/mc-service-ui.yml
	sleep 2

	echo "Waiting for mc-server pod to be available..."

	while [ ! "$(kubectl get pods | grep "mc-service" | grep 'Running' | wc -l)" -ge 1 ]
	do
		sleep 2;
	done

	echo ""
	web_service_ip=$(kubectl get services | grep "web-service" | awk '{print $3}')
	nodePort=$(kubectl get services | grep "web-service" | awk '{print $5}' | cut -d ":" -f 2 | cut -d / -f 1 )
	masterNodeIP=$(kubectl describe node $masterNode | grep "InternalIP" | cut -d ":" -f 2 | sed -e 's/^[[:space:]]*//')

	echo "You may access the web server UI by visiting one of the following addresses in your web browser:"
	echo -e "\thttp://$web_service_ip:8070 on any machine within the cluster, or"
	echo -e "\thttp://$masterNodeIP:$nodePort externally on any machine within the same network"
	echo ""
	echo "Ready to go!"
}

function bringDownServices {
	echo ""
	echo "***Removing all services***"
	echo ""
	echo "Cleaning up mc-server"
	echo "---------------------"
	kubectl delete service mc-service
	kubectl delete deployment mc-service

	echo ""
	echo "Cleaning up web-server"
	echo "----------------------"
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
	masterNode=$(kubectl get nodes | grep "master" | awk '{print $1}')
	kubectl label node $masterNode dbtype-
	
	echo ""
	echo "Done!"
}

num_arguments=$#

if [ $num_arguments -eq 0 ] || [ $num_arguments -gt 1 ]; then
	echo "This script requires one parameter! Acceptable parameters are:"
	echo -e "\tup\t - deploy all services, or"
	echo -e "\tdown\t - remove all services"
	exit
fi

action=$1

if [ $action == "up" ]; then
	bringUpServices
elif [ $action == "down" ]; then
	bringDownServices
else
	echo "Only acceptable parameters are 'up' or 'down'"
fi
