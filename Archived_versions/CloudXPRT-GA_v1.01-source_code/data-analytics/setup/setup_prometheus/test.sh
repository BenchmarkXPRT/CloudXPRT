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

strToParse="Total_duration_(secs):180.45,1574365628_1_tearsofsteel_1080p_1_min.mp4:171.89, 1574365627_0_tearsofsteel_1080p_1_min.mp4:174.05,  Number of pods:2, vCpusperPod:4"

echo $strToParse

#echo $strToParse | awk '{split($0,numbers,",")} END {for(n in numbers){ print numbers[n] }}'
echo "next"

IFS=',' # hyphen (-) is set as delimiter
read -ra ADDR <<< "$strToParse" # str is read into an array as tokens separated by IFS
for i in "${ADDR[@]}"; do # access each element of array
    #echo "$i"
    echo "$i" | awk -F':' '{print $2}'
done
IFS=' '
