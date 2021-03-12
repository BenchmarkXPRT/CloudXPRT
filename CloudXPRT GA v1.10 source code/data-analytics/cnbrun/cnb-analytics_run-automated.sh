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

# requires apt install moreutils jq
apt install -y moreutils jq

# vcpus in the system
num_cores=0
num_nodes=$(kubectl get node --no-headers -o custom-columns=NAME:.metadata.name | wc -l)
nodes=$(kubectl get nodes | awk 'FNR > 1')
while read -r node; do
    node_name=$(echo $node | awk '{print $1}')
    node_role=$(echo $node | awk '{print $3}')
    cores=$(kubectl describe node $node_name | grep cpu: | head -n 1 | awk '{print $2}')
    # if control-plane node has more than 2 cores, include them.
    if [ "$node_role" = "master" ] || [ "$node_role" = "Master" ]; then
        if ((cores > 2)); then
            num_cores=$((num_cores + cores))
        fi
    else
        num_cores=$((num_cores + cores))
    fi
done <<<"$nodes"

echo "Number of nodes in cluster: $num_nodes"
echo "Total number of cores: $num_cores"

case "$num_cores" in
  16) Lambda=(87 88 89); vCPU_per_POD=(12);
      # vCPU_per_POD=(12);
      ;;
  24) Lambda=(70 80 90); vCPU_per_POD=(20);
      # vCPU_per_POD=(20);
      ;;
  32) Lambda=(36 37 38); vCPU_per_POD=(14);
      # vCPU_per_POD=(28 14);
      ;;
  40) Lambda=(33 34 35); vCPU_per_POD=(18);
      # vCPU_per_POD=(36 18);
      ;;
  48) Lambda=(31 32 33); vCPU_per_POD=(22);
      # vCPU_per_POD=(11 14 22);
      ;;
  56) Lambda=(28 29 30); vCPU_per_POD=(26);
      # vCPU_per_POD=(52 26 13);
      ;;
  64) Lambda=(26 27 28); vCPU_per_POD=(20);
      # vCPU_per_POD=(30 20 15 10);
      ;;
  72) Lambda=(31 32 33); vCPU_per_POD=(34);
      # vCPU_per_POD=(34 22 17 11);
      ;;
  80) Lambda=(24 25 26); vCPU_per_POD=(19);
      # vCPU_per_POD=(38 25 19 12)
      ;;
  88) Lambda=(24 25 26); vCPU_per_POD=(14);
      # vCPU_per_POD=(42 28 21 14)
      ;;
  96) Lambda=(17 18 19); vCPU_per_POD=(15);
      # vCPU_per_POD=(46 30 23 15)
      ;;
  104) Lambda=(17 18 19); vCPU_per_POD=(16);
      # vCPU_per_POD=(25 16 12)
      ;;
  112) Lambda=(16 17 18); vCPU_per_POD=(13);
       # vCPU_per_POD=(54 36 27 18 13)
       ;;
  128) Lambda=(12 13 14); vCPU_per_POD=(12);
       # vCPU_per_POD=(62 41 31 20 15 12)
       ;;
  144) Lambda=(11 12 13); vCPU_per_POD=(14);
       # vCPU_per_POD=(70 46 35 23 17 14)
       ;;
  152) Lambda=(11 12 13); vCPU_per_POD=(14);
       # vCPU_per_POD=(70 46 35 23 17 14)
       ;;
  160) Lambda=(10 11 12); vCPU_per_POD=(19);
       # vCPU_per_POD=(78 52 39 26 19 15 13)
       ;;
  192) Lambda=(15 16 17); vCPU_per_POD=(14)
       # vCPU_per_POD=(94 62 47 31 23 18 15 13)
       ;;
  224) Lambda=(6 7 8); vCPU_per_POD=(15);
       # vCPU_per_POD=(110 73 55 36 27 22 18 15)
       ;;
  256) Lambda=(12 13 14); vCPU_per_POD=(31);
       # vCPU_per_POD=(126 84 63 42 31 25 21 18 15)
       ;;
  *) echo "No DataBase Records for $num_cores vCPUs"
     echo "Please edit this line with desired values..."
     Lambda=(20); vCPU_per_POD=(12);
     exit
     ;;
esac

config_file='cnb-analytics_config.json'
MESSAGES=100
arr_dataset=(higgs1m)

# DEBUG
# Lambda=(12 13);
# vCPU_per_POD=(31);

## now loop through the above array
for dataset in "${arr_dataset[@]}"; do
    for cpus in ${vCPU_per_POD[@]}; do
        if ((cpus == 0)); then
            continue
        fi

        for lambda in "${Lambda[@]}"; do
            FILENAME="output_${cpus}vcpus_${MESSAGES}mges_${lambda}lamb_${dataset}"
            echo $FILENAME

            jq '.workload[].cpus_per_pod="'${cpus}'"' $config_file | sponge $config_file
            jq '.workload[].numKAFKAmessages="'${MESSAGES}'"' $config_file | sponge $config_file
            jq '.workload[].loadgen_lambda="'${lambda}'"' $config_file | sponge $config_file

            mkdir output
            cp $config_file output
            ./cnbrun
            mv output $FILENAME
        done
    done
done

chmod -R 755 $PWD
#chown -R t:t $PWD

#grep NumberOfPods output_*/output_result.txt | sort -n -k1.16
#grep vCPUsperPod output_*/output_result.txt | sort -n -k1.16
#grep TotalDuration output_*/output_result.txt | sort -n -k1.16
#grep 90th_Percentile output_*/output_result.txt | sort -n -k1.16
#grep "Throughput_tnx/min" output_*/output_result.txt | sort -n -k1.16
