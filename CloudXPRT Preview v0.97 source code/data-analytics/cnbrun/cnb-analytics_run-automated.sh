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

# requires apt install moreutils jq
apt install -y moreutils jq

config_file='cnb-analytics_config.json'
MESSAGES=100
arr_dataset=(higgs1m)

#ALL_THREADS=`grep -c ^processor /proc/cpuinfo`
#START=104
#INCREMENT=4

#vCPU_per_POD=(125 62 31 15) # 256vCPU
#vCPU_per_POD=(110 55 36 27 22 18 15 13) # 224vCPU
#vCPU_per_POD=(94 47 23 18 15) # 192vCPU
#vCPU_per_POD=(62 31 20 15 10) # 128vCPU
#vCPU_per_POD=(54 27 13 10) # 112vCPU
vCPU_per_POD=(46 23 15 11) # 96vCPU
#vCPU_per_POD=(30 20 15 12) # 64vCPU

Lambda=(10 20 30 40)

## now loop through the above array
for dataset in "${arr_dataset[@]}"
do
   #echo "$dataset"

   #for ((cpus=$START;cpus<ALL_THREADS;cpus+=$INCREMENT)); do
   for cpus in ${vCPU_per_POD[@]}; do
      if [ $cpus -eq 0 ]; then
         continue
      fi

      #MESSAGES=$((ALL_THREADS / cpus))
      for lambda in "${Lambda[@]}"
      do
        #FILENAME="output_"$dataset"_"$cpus"vcpus_"$MESSAGES"mges"_$lambda"lambda"
        FILENAME="output_"$cpus"vcpus_"$MESSAGES"mges"_$lambda"lamb_"$dataset
        echo $FILENAME

        jq '.workload[].cpus_per_pod="'${cpus}'"' $config_file | sponge $config_file
        jq '.workload[].numKAFKAmessages="'${MESSAGES}'"' $config_file | sponge $config_file
        jq '.workload[].loadgen_lambda="'${lambda}'"' $config_file | sponge $config_file
        #cat xgboost_config.json

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
