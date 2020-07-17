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
# This script runs the CNB OCR workload. cnbrun binary will call this script.
# It should only be used after the Kubernetes cluster is up.
###############################################################################

if [ "$EUID" -ne 0 ]; then
  echo ""
  echo "Please run as root"
  echo ""
  exit
fi

echo $@
CNB_DIRECTORY=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# Variable assignation
version=$1
num_cpus_per_pod=$2
num_kafka_messages=$3
loadgen_lambda=$4

#fixed parameters
num_cpu_to_harness=4
output_dir="./output/"
dataset="higgs1m"

if [ "$loadgen_lambda" = "0" ]
then
    echo "xgboost.sh: There is an Error: Lambda is equal to ZERO !!! Please specify a valid number for LAMBDA in the config json file"
    exit
fi

# Check that all nodes are ready
echo "xgboost.sh: Checking that all nodes are ready..."
nodes=$(kubectl get nodes | awk 'FNR > 1')
all_ready=true
num_cores=0
num_nodes=0

while read -r node; do
        node_name=$(echo $node | awk '{print $1}')
        node_status=$(echo $node | awk '{print $2}')
        if [ ! $node_status = "Ready" ]; then
                echo "Node $node_name is not ready!"
                $all_ready=false
        fi
        node_type=$(echo $node | awk '{print $3}')
        if [[ $node_type == "node" ]]; then
                (( num_nodes++ ))
                cores=$(kubectl describe node $node_name | grep cpu: | head -n 1 | awk '{print $2}')
                num_cores=$(($num_cores + $cores))
        fi
done <<< "$nodes"

if [ $all_ready = false ]; then
	exit 1
else
	echo "All nodes are ready!"
fi

echo "xgboost.sh: Number of nodes in cluster:" $num_nodes
echo "xgboost.sh: Total number of cores: "$num_cores

# recompute available_num_cores
echo "xgboost.sh: Number of CPUs per pod:" $num_cpus_per_pod
# Leave resources for webserver pod(s) and kube-system pods based on number of cores per MC pod
available_num_cores=$(( $num_cores - ( $num_cpu_to_harness * $num_nodes )))
echo "xgboost.sh: available_num_cores: " $available_num_cores "We reserve $num_cpu_to_harness CPUs per Node for webserver and kube-system pods"
max_xgb_pods=$(( $available_num_cores  / $num_cpus_per_pod  ))

# Make sure there is simetry in POD assigment per Node
if [ $(($max_xgb_pods%$num_nodes)) -eq 0 ]; then
        echo "There is node/pod simetry..."
else
        echo "There is no simetry, Lets remove one pod.."
        max_xgb_pods=$(($max_xgb_pods-1))
        echo $max_xgb_pods
fi

echo "xgboost.sh: Total number of cores in all nodes:" $num_cores
echo "xgboost.sh: Max number of xgboost PODs: " $max_xgb_pods
echo "xgboost.sh: dataset: " $dataset
echo "xgboost.sh: numKAFKAmessages: " $num_kafka_messages
echo "xgboost.sh: num_cpu_to_harness: " $num_cpu_to_harness

if [ "$max_xgb_pods" -le "0" ]
then
    echo "xgboost.sh: There is an Error: You are trying to oversubscribe vCPUs per Pods. We reserve $num_cpu_to_harness CPUs per Node for webserver and kube-system pods"
    exit
fi

# DEBUG
#max_xgb_pods=0
#num_cpus_per_pod=0
#num_kafka_messages=0
#dataset="higgs1m"
echo -e "\nxgboost.sh: Setting up $max_xgb_pods PODS of $num_cpus_per_pod vCPUs - Delivering $num_kafka_messages kafka_messages using $dataset dataset. per Node-CPUs devoted to harness: $num_cpu_to_harness. Lambda $loadgen_lambda\n"

cnb_results_foldername="output_cnb_xgb_"
datetimelabel=$(echo $(date +"%Y_%m_%d_%I_%M_%p"))
cnb_results_foldername="${cnb_results_foldername}${datetimelabel}"
if [ ! -d $output_dir ]; then
  mkdir $output_dir;
fi

#create security creds to use python k8s client
echo "xgboost.sh: creating security credentials..."
kubectl apply -f xgboost/xgboost_workload_yml/security.yml
kubectl apply -f xgboost/xgboost_workload_yml/security2.yml

#Creating persistent volumes and claims for the web server (which is the load gen)
echo "xgboost.sh: creating persistent volumes..."
kubectl apply -f xgboost/xgboost_workload_yml/pv-volume.yml
sleep 10
while [ ! "$(kubectl get pv/task-pv-volume -n kafka | grep "task-pv-volume" | awk '{print $5}')" = "Available"  ]
do
	if [ "$(kubectl get pv/task-pv-volume -n kafka | grep "task-pv-claim" | awk '{print $5}')" = "Bound"  ]; then
		break;
	else
		echo "xgboost.sh: task-pv-volume not available"
		sleep 2;
	fi
done

echo "xgboost.sh: creating claims for the web server..."
kubectl apply -f xgboost/xgboost_workload_yml/pv-claim.yml
sleep 10
while [ "$(kubectl get pvc/task-pv-claim -n kafka | grep "task-pv-claim" | awk '{print $5}')" = "Pending"  ]
do
	if [ "$(kubectl get pvc/task-pv-claim -n kafka | grep "task-pv-claim" | awk '{print $5}')" = "Bound"  ]; then
		break;
	else
		echo "xgboost.sh: task-pv-claim not available"
		sleep 2;
	fi
done

echo "xgboost.sh: Deploying cdn-webserver..."
echo "xgboost.sh: creating claims for the web server..."
kubectl apply -f xgboost/xgboost_workload_yml/cdn-webserver.yml
sleep 10
while [ ! "$(kubectl get pvc/task-pv-claim -n kafka | grep "task-pv-claim" | awk '{print $2}')" = "Bound"  ]
do
	echo "xgboost.sh: task-pv-claim not bound"
	sleep 2;
done

echo "xgboost.sh: waiting for cdn-webserver pod to be available..."
while [ ! "$(kubectl get pods -n kafka | grep "cdn-webserver" | awk '{print $2}')" = "1/1" ]
do
	sleep 2;
done

echo "xgboost.sh: exposing cdn-webserver to kafka namespace..."
kubectl expose pod cdn-webserver -n kafka --name=cdn-webservice
while [ ! "$(kubectl get svc -n kafka | grep "cdn-webservice" | awk '{print $5}')" = "7079/TCP" ]
do
	echo "xgboost.sh: waiting for cdn-webservice"
	sleep 2;
done
web_service_ip=$(kubectl get services -n kafka | grep "cdn-webservice" | awk '{print $3}')
echo "xgboost.sh: IP address for cdn-webservice $web_service_ip"

#echo ""
#echo "Deploying HPA for web-server"
#echo "----------------------------"
#kubectl autoscale deployment web-service --cpu-percent=75 --min=1 --max=5

#start prometheus service:
echo "xgboost.sh: start prometheus service to monitoring namespace..."
kubectl apply -f  ../setup/setup_prometheus/prometheus-service.yaml
sleep 10
while [ ! "$(kubectl get svc -n monitoring | grep "prometheus-service" | awk '{print $5}')" = "8080:30000/TCP" ]
do
	echo "waiting for prometheus-service"
	sleep 2;
done
prom_service_ip=$(kubectl get services -n monitoring | grep "prometheus-service" | awk '{print $3}')
#prom_deploy_pod_name=$(kubectl get pods -n monitoring | grep "prometheus-deployment-[a-z 0-9 -]"  | awk '{print $1}')
#echo "Pod name for prometheus:" $prom_deploy_pod_name
#kubectl port-forward $prom_deploy_pod_name 8080:9090 -n monitoring
#Run monitoring to get CPU usage and other metrics
#echo "Results folder: /mnt/data/"$cnb_results_foldername

#Gather IP from minio service
minio_service_ip=$(kubectl get services | grep "minio-service" | awk '{print $3}')
minio_service_ip=$minio_service_ip:9000
echo "minioserviceip: $minio_service_ip"

#########################
# xgboost Workload
#########################
sleep 6
echo ""
echo "xgboost.sh: Testing xgboost server using curl..."
echo "------------------------------------------------"
#curl --silent --noproxy "*" http://$web_service_ip:7079/xgb
curl --silent --noproxy "*" http://$web_service_ip:7079/xgb?resfolder=$cnb_results_foldername #RESULTS
echo "xgboost.sh: Ready to go!"
echo ""
echo "###############################"
echo "Running XGBoost workload"
echo "###############################"

retcode=$(curl -silent --noproxy "*" "http://$web_service_ip:7079/xgb_StartTest?numberOfPods=$max_xgb_pods&cpusPerPod=$num_cpus_per_pod&numKAFKAmessages=$num_kafka_messages&dataset=$dataset&minioSERVICEip=$minio_service_ip&loadgen_lambda=$loadgen_lambda"| grep HTTP/ | awk {'print $2'})

# delay time make sure all messages are produced before Starting
sleep 60

echo "xgboost.sh: xgbStartTest retcode:" $retcode
if [ ${retcode} == "202" ]
then
	echo "xgboost.sh: Accepted"
	outputFull=$(curl --silent --noproxy "*" http://$web_service_ip:7079/xgb_Status)
	IFS=':' read -r output string <<< "$outputFull"
	#output=$outputFull | awk 'NR == 1 {print $0}'
	echo "xgboost.sh: First xgb_Status " $output

	while [ "$output" = "Test in Progress..." ]
	do
        outputFull=$(curl --silent --noproxy "*" http://$web_service_ip:7079/xgb_Status)
	    IFS=':' read -r output string <<< "$outputFull"
 		echo "xgboost.sh: xgb_Status $output"
		if [ "$output" != "Test is Done" ]
		then
			sleep 10
		else
			echo "xgboost.sh: Creating Report... "	$output
		fi
	done
else
    echo "xgboost.sh: Unexpected retcode $retcode from:"
    echo "curl -silent --noproxy "*" "http://$web_service_ip:7079/xgb_StartTest?numberOfPods=$max_xgb_pods&cpusPerPod=$num_cpus_per_pod&numKAFKAmessages=$num_kafka_messages&dataset=$dataset&minioSERVICEip=$minio_service_ip&loadgen_lambda=$loadgen_lambda""
    exit 1
fi


output=$(curl --silent --noproxy "*" http://$web_service_ip:7079/xgb_Status)
echo "xgboost.sh:" $output
echo ""

output_file="$output_dir/output_result.txt"
if [ -f $output_file ] ; then
    rm $output_file
fi
if [[ $output == "Test is Done"* ]];then
    preffix="Test is Done:"
    out_short=${output#"$preffix"}
    #echo $out_short
    lines=$(echo $out_short | tr "," "\n")
    #echo $lines
    IFS=',' ;for i in $lines ; do echo $i | sed -e "s/:/ /" | sed -e "s/^ //" >> $output_file ;done
else
    echo "xgboost.sh: ERROR... Return Message is incomplete !!!"
fi

echo "Lambda " $loadgen_lambda >> $output_file

cat $output_file

duration=`cat $output_dir/output_result.txt | grep TotalTimeWithSetup | awk '{print $2}'`
#duration=50
#Run monitoring to get CPU usage and other metrics
echo ""
echo "xgboost.sh: Parsing from Prometheus... "
echo "    prom_service_ip: "$prom_service_ip
echo "    duration: "$duration
echo "    num_cpus: "$num_cpus_per_pod
echo "    cnbresfoldername: "$cnb_results_foldername
#echo "    Results folder: /mnt/data/"$cnb_results_foldername

echo "./monitoring GetCpuUsageByPOD $prom_service_ip $duration xgb $num_cpus_per_pod $output_dir/$cnb_results_foldername"
./monitoring GetCpuUsageByPOD $prom_service_ip $duration xgb $num_cpus_per_pod $output_dir"/"$cnb_results_foldername

xgbPODS=$(kubectl get pods --selector=app=xgb -n kafka | awk 'FNR >= 2 {print $1}')
echo ""$xgbPODS
while read -r podname; do
    ./monitoring GetMemoryUsageByPOD $prom_service_ip $duration $podname $num_cpus_per_pod $output_dir"/"$cnb_results_foldername
done <<< "$xgbPODS"

./monitoring GetCPUThrottleInfoByPOD $prom_service_ip $duration xgb $num_cpus_per_pod $output_dir"/"$cnb_results_foldername
./monitoring GetCPUUsageByInstance $prom_service_ip $duration xgb $num_cpus_per_pod $output_dir"/"$cnb_results_foldername
./monitoring GetContextSwitchesByNode $prom_service_ip $duration xgb $num_cpus_per_pod $output_dir"/"$cnb_results_foldername

echo ""
echo "xgboost.sh: Total number of pods created during the run"
echo "-------------------------------------------------------"
kubectl get pods --all-namespaces
kubectl get pod -o=custom-columns=NAME:.metadata.name,STATUS:.status.phase,NODE:.spec.nodeName --all-namespaces
kubectl logs -n kafka cdn-webserver > $output_dir/cdn-webserver.log
#echo "webservercdn pods: $(kubectl get pods | grep "cdn-webserver" | grep "Running" | wc -l)"
#echo "xgb service pods: $(kubectl get pods | grep "xgb" | grep "Running" | wc -l)"
echo ""

echo "###############################"
echo "DONE XGBoost workload"
echo "###############################"

echo "xgboost.sh: Cleaning up pods and services"
echo "-----------------------------------------"
time_start_cleaning=$SECONDS
echo cleaning started at: $time_start_cleaning
xgbpods=$(kubectl get pods --selector=app=xgb -n kafka | awk 'FNR >= 2 {print $1}')
while read -r podname; do
	echo "deleting "$podname
	#kubectl logs pod/$podname -n kafka >> ./$output_dir/$podname.log
	#sleep 10
	kubectl delete pod/$podname -n kafka
done <<< "$xgbpods"
kubectl delete service cdn-webservice -n kafka
kubectl delete pod/cdn-webserver -n kafka
kubectl delete service prometheus-service -n monitoring

kubectl delete pvc/task-pv-claim -n kafka
kubectl delete pv/task-pv-volume -n kafka

duration_cleaning=$(( SECONDS - time_start_cleaning ))
echo cleaning ended at: $SECONDS
echo "DurationCleaning $duration_cleaning" >> $output_file

#cd cnb_plot_hdr; python3 hdr_cnb-analytics.py; cd ../

lscpu > $output_file.system.txt
lshw -short | grep -i "system memory" >> $output_file.system.txt
