## Running Cloud Native Benchmark

#### Configure parameters for a test run
Open cnb-analytics_config.json file to set the parameters for CNB.
   ```
   nano cnb-analytics_config.json
      cpus_per_pod: Number of vCPUs per Pod. default 12
      numKAFKAmessages: Number of transactions to be delivered and executed. default 1
      loadgen_lambda: Inter-arrival time between transactions following Poisson distribution. default 0.33
   ```

#### Run CNB-analytics
Once parameters are configured, run the cnbrun executable.
   ```
   su
   ./cnbrun 
   ./cnb-analytics_parse-all-results.sh
   ```
**NOTE:** use cnb-analytics_clear.sh to reset kubernetes in case you have an invalid run. then re-issue ./cnbrun

#### Deep dive analysis to determine best system configuration
A script is provided to create a swept analysis in order to find the best throughput under a particular SLA.
   ```
   su
   ./cnb-analytics_run-automated.sh
   ```

Make sure you set the desired parameters
   ```
   nano cnb-analytics_run-automated.sh
      Lambda: sets the desired Inter-arrival time for the Poisson distribution. default Lambda=(0.33 0.66 0.85 1)
      vCPU_per_POD: sets the desired swept for different number of vCPUs per pod. default vCPU_per_POD=(46 23 15 11)
   ```

In case of errors please clear the temp PODs using:
   ``` 
   su
   ./cnb-analytics_clear.sh
   ```

**Results**:
A script is provided to create a table from output folders
   ```
    ./cnb-analytics_parse-all-results.sh
   ```
You can easily create a csv file using these command: ./cnb-analytics_parse-all-results.sh | sed -e 's/\s\+/,/g' > results.csv

Some of the metrics listed in the output are listed below:
- NumberOfPods: number of working Pods
- vCPUsperPod: number of vCPUs used per Pod
- DeliveredKAFKAmessages: number of Kafka messages that were processed among Pods.
- 90th_Percentile: Tail latency for the 90th percentile
- Throughput_jobs/min: Throughput in transactions per minute
User has the freedom to define a throughput that comply with 90th_Percentile latency.

**Notes**:
- cnbrun, xgboost.sh, cnb-analytics_config.json must all be in the same directory
- xgbooost.sh need to have executable permissions

**Q&A**:

Q1. The benchmark is looping for long time while setting up any of the "pod/services"
- Please open a new console and use the script to visually verify there are no errors and/or pods in underfined status.
   ```
   su
    ./cnb-analytics_status.sh
   ```



