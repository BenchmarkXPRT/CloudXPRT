#!/bin/bash
DIRECTORY=$(pwd)"/output*"

declare -a RESULTS=(
    "NumberOfPods"
    "vCPUsperPod"
    "DeliveredKAFKAmessages"
    "Lambda"
    "TotalDuration"
    "min_duration"
    "max_duration"
    "stdev_duration"
    "mean_duration"
    "90th_Percentile"
    "95th_Percentile"
    "Throughput_tnx/min"
)

declare -a RESULTS_NAME=(
    "NumberOfPods"
    "vCPUsperPod"
    "#Jobs"
    "Lambda"
    "jobs_duration"
    "min_duration"
    "max_duration"
    "stdev_duration"
    "mean_duration"
    "90th_Percentile"
    "95th_Percentile"
    "Throughput_jobs/min"
)

# print header
echo -n "FILE"
for RESULT in "${RESULTS_NAME[@]}"
do
    echo -n ","$RESULT
done
echo ""

#grep every file
for filename in $DIRECTORY/output_result.txt; do
    echo -n $filename
    for RESULT in "${RESULTS[@]}"
    do
        grep $RESULT $filename | awk '{ printf ","$NF }'
    done
    echo ""
done

exit
