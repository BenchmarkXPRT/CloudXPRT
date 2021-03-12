## Running CloudXPRT Benchmark

#### Configure parameters for a test run
Edit the `cnb-analytics_config.json` file to set the parameters for CloudXPRT data-analytics. It is sugested you run it once with default parameters to make sure correct functionallity.

The following parameters are the most important for determining perfomnance.

- `cpus_per_pod`: Number of vCPUs per Pod. default 12
- `numKAFKAmessages`: Number of transactions to be delivered and executed. default 1. Users should normally use at least 100 messages for statistical purposes.
- `loadgen_lambda`: Inter-arrival time in seconds between transactions following Poisson distribution. default 12

#### Run CloudXPRTS analytics workload
Once parameters are configured, run the `cnbrun` executable.
   ```
   sudo ./cnbrun 
   sudo ./cnb-analytics_parse-all-results.sh
   ```
**Note**: use `cnb-analytics_clear.sh` to reset Kubernetes in case you have an invalid run. Then rerun `./cnbrun`

#### Deep dive analysis to determine the cluster configuration for best throughput.
A script is provided to create a swept analysis in order to find the best throughput under a particular SLA.
   ```
   sudo ./cnb-analytics_run-automated.sh
   ```
Make sure you set the desired parameters. Identify the number of vCPUs in your cluster and Look through the case scenarios, starting at line 48.

In the following example, we will be editing the parameters for a cluster with 48 vCPUs (line 60). In this case, we willi run the workload using three different Lambdas and three different vCPUs-per-Pod for a total of 9 different runs. The modified line is `48) Lambda=(31 32 33); vCPU_per_POD=(11 14 22);`.
```
   sudo ./cnb-analytics_run-automated.sh
   sudo ./cnb-analytics_parse-all-results.sh
```

In case of errors please clear the temp PODs using:
   ``` 
   sudo ./cnb-analytics_clear.sh
   ```

## Benchmark results
A script is provided to create a CSV table from the data in the output folders
   ```
    ./cnb-analytics_parse-all-results.sh
   ```

Some of the meThe main metrics in the results table are --
- `NumberOfPods`: number of working Pods
- `vCPUsperPod`: number of vCPUs used per Pod
- `DeliveredKAFKAmessages`: number of Kafka messages that were processed among Pods.
- `90th_Percentile`: Tail latency for the 90th percentile
- `Throughput_jobs/min`: Throughput in transactions per minute

The user has the freedom to define a throughput that comply with 90th_Percentile latency.

The complete format of the results table is as follows --
- `FILE`: location of output_result.txt file containing the results from a particular simulation
- `NumberOfPods`: number of working Pods executing XGBoost
- `vCPUsperPod`: number of vCPUs used per Pod (executing XGBoost training)
- `number of Jobs`: Total number of jobs executed during the simulation
- `Lambda`: Interarrival time in seconds between transactions. The lower the lambda the more traffic sent to the CUT.
- `jobs_duration`: Includes creation of Pods and the eleapsed time between the creation of the first transaction by the Load generator until arrival of the last transaction.
- `min_duration`: minimum transaction time
- `max_duration`: maximum transaction time
- `stdev_duration`: Standard deviation of transaction time
- `mean_duration`: maximum transaction time
- `90th_Percentile`: 90th percentile for transaction times
- `95th_Percentile`: 95th percentile for transaction times
- `Throughput_jobs/min`: Throughput in transactions per minute

## Known Issues

### FAQ

Q1. The benchmark is looping for long time while setting up any of the "pod/services"

Please open a new console and use the script to visually verify there are no errors and/or pods in underfined status.
   ```
   sudo ./cnb-analytics_status.sh
   ```

 Q2. How to clean the current execution and start again?

1. Identify any running scripts with the following command and then kill them.
   ```
   ps -aux | grep cnbrun | grep automated
   ```
2. Remove all pods associated with the workload
    ```
    sudo ./cnb-analytics_clear.sh
   ```
3. Reset the `kafka` subsystem
   ```
    cd ../setup/setup_kafka
    sudo ./cleanup.sh
    sudo ./setup.sh
   ```
4. Restart the run
   ```
    cd ../../cnbrun
    sudo ./cnbrun
    ```
    
