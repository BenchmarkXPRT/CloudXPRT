package main
/*******************************************************************************
* Copyright 2020 BenchmarkXPRT Development Community
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************************/

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	//"reflect"
	"strconv"
	"strings"

	//"reflect"
	//"strconv"
	"time"
	//"timeutil"
	"github.com/levenlabs/golib/timeutil"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)


/*

https://www.robustperception.io/what-range-should-i-use-with-rate#more-4571
Prometheus:
default::
  scrape_interval: 5s
  scrape_timeout: 5s
  evaluation_interval: 5s

The general rule for choosing the range is that it should be at least 4x the scrape interval. This is to allow for various races, and to be resilient to a failed scrape.

Let's say you had a 10s scrape interval, and scrapes started at t=0. The rate function needs at least two samples to work, so for a query at t=10 you'd need 1x the scrape interval.
At t=20, the scrape at that time may not have been fully ingested yet so 2x will cover you for two samples back to t=0. At t=29 that scrape might still not have been ingested, so you'd need ~3x to be safe.
Finally you want to be resilient to a failed scrape. If the t=20 scrape fails and you're at t=39 but the t=30 scrape is still ongoing, then you'd need ~4x to see both the t=0 and t=10 samples.
So a 40s rate (i.e. rate(my_counter_total[40s]) would be the minimum safe range. Usually you would round this up to 60s for a 1m rate.

scrape interval (seconds)	~4x	round up to 60s
10	40	60s
60	240	300s
		5m
https://www.robustperception.io/step-and-query_range#more-4519:
One consequence of this is that you must take a little care when choosing the range for functions like rate or avg_over_time, as if it's smaller then the step then you'll undersample and skip over some data.

https://medium.com/@wbassler23/getting-started-with-prometheus-pt-1-8f95eef417ed

changing
global:
    evaluation_interval: 15s
    scrape_interval: 15s
    scrape_timeout: 10s
*/

/*
https://www.robustperception.io/what-is-a-job-label-for#more-4052
What is a job label for?
The job label is one of the labels your targets will always have. So how can you use it?

The job label is special as if after target relabelling if there is no job label present on a target, then the value of the job_name field  will be used. In this way your targets will always have a job label.

How best to use it then? What I would recommend is using the job label to organise applications which do the same thing, which almost always means processes running the same binary with the exact same configuration.
For example you might have web frontends with job="frontend" and a Redis used as a cache with job="redis". Using labels like this it's easy to aggregate across a job, for example CPU usage per job would be
sum by (job)(rate(process_cpu_seconds_total[5m])). As a counterexample a job="kubernetes" for all of your processes running on Kubernetes isn't particularly useful, you should aim for more meaningful job labels.

If you were running different sets of Redis servers for different purposes, it'd make sense to have different job labels for each set. Each set will likely have distinct performance characteristics after all, and aggregating them together would be unlikely to be of much use. If there are subdivisions within a set, such as sharding, it often makes sense to add a shard label or similar to subdivide the job.
*/
var(
	functionToRun string =""
	promServerUrl string = ""
	durationForMetrics int = 0
	podRegexStr string = ""
	cpuLimitPerPod float64 = 0.0
	outputDir = "output/"
	resultsFolder string=""
	dur = 0
)
type resultObject struct {
	Metric json.RawMessage `json:"metric"`
	Values json.RawMessage `json:"values"`
}
func main(){
	runArgs	:= os.Args[1:]

	if len(runArgs) == 6{
		functionToRun = runArgs[0]
		var ipAddressForProm string = runArgs[1]
		promServerUrl="http://" + ipAddressForProm + ":8080"
		fmt.Println("URL:",promServerUrl)

		dur, err := strconv.ParseFloat(runArgs[2], 64)
		if err != nil {
			fmt.Println("Provide duration for reporting : %s", err)
		}
		fmt.Println("Duration:",dur)
		durationForMetrics = int(dur)
		resultsFolder = runArgs[5]
	}else{
		var usageDesc string = "monitoring usage:\n"
		usageDesc= usageDesc + "Function To Call\n"
		usageDesc= usageDesc + "IP address of Prometheus service\n"
		usageDesc= usageDesc + "Duration in seconds - from (current time - duration) to current time\n"
		usageDesc= usageDesc + "Regex string to find pods of interest example'vod'\n"
		usageDesc= usageDesc + "CPU limit for pods of interest\n"
		usageDesc= usageDesc + "For example to get data for the past 400 seconds for a pod with a label containing 'vod' launched with a limit of 4 vcpus and to store results in resultsfolder use\n"
		usageDesc= usageDesc + "monitoring ipAddress_of_Promtheus_service 400 vod 4 resultsfolder\n"

		fmt.Println(usageDesc)
	}

	if len(functionToRun) <= 0 {
		os.Exit(1)
	}

	switch functionToRun {
	case "GetCpuUsageByPOD":
		podRegexStr = runArgs[3]
		cpuLimitPerPod, _= strconv.ParseFloat(runArgs[4],64)
		GetCpuUsageByPOD(durationForMetrics, 60, 12, podRegexStr, cpuLimitPerPod, resultsFolder)
		break
	case "GetMemoryUsageByPOD":
		podRegexStr = runArgs[3]
		GetMemoryUsageByPOD(durationForMetrics,60, 12, podRegexStr,resultsFolder)
		break
	case "GetCPUThrottleInfoByPOD":
		podRegexStr = runArgs[3]
		GetCPUThrottleInfoByPOD(durationForMetrics,60, 12, podRegexStr,resultsFolder)
		break
	case "GetCPUUsageByInstance":
		cpuLimitPerPod, _= strconv.ParseFloat(runArgs[2],64)
		GetCPUUsageByInstance(durationForMetrics, 60, 12,resultsFolder)
		break
	case "GetContextSwitchesByNode":
		GetContextSwitchesByNode(durationForMetrics,60,12,resultsFolder)
		break
	default:
		fmt.Println("Unknown function")
	}
}


//func GetCpuUsageByPOD: This uses the rate function
//example:
//rate(http_requests_total[5m])[30m:1m]  :: this is: Return the 5-minute rate of the http_requests_total metric for the past 30 minutes, with a resolution of 1 minute

// Let's say you had a 10s scrape interval, and scrapes started at t=0. The rate function needs at least two samples to work, so for a query at t=10 you'd need 1x the scrape interval.
//At t=20, the scrape at that time may not have been fully ingested yet so 2x will cover you for two samples back to t=0. At t=29 that scrape might still not have been ingested, so you'd need ~3x to be safe.
//Finally you want to be resilient to a failed scrape. If the t=20 scrape fails and you're at t=39 but the t=30 scrape is still ongoing, then you'd need ~4x to see both the t=0 and t=10 samples.
//So a 40s rate (i.e. rate(my_counter_total[40s]) would be the minimum safe range. Usually you would round this up to 60s for a 1m rate.
//
//scrape interval (seconds)	~4x	round up to 60s
//10s	40s	60s
//60s	240s 300s
//		5m

//
// We have the following scrape settings:
//  scrape_interval: 5s
//  scrape_timeout: 5s
//  evaluation_interval: 5s

// 5s  20s  30s


//The rate(v range-vector)[time]function takes a range vector. it calculates the average over time

//returnRate - return rate in seconds
//duration in seconds - from (current time - duration) to current time
//resStep - from (current time - duration) to current time.// The range in query that is, returnRate should not be smaller than this
//currently need to pass cpu limit
func GetCpuUsageByPOD (duration int,returnRate int, resStep int, podNameRegEx string,cpuLimit float64, resFolder string ){
	//create results folder if it does not exist and create resource usage file
	s := strings.Split(resFolder, "/")
	//fmt.Printf( "resFolder: '%s'\n", resFolder);
	var fileName string = s[1]
	//fmt.Printf( "fileName: '%s'\n", fileName);
	if _, create_dir_err := os.Stat(resFolder); os.IsNotExist(create_dir_err) {
		os.Mkdir(resFolder, os.FileMode(0755))
	}
	resf, err := os.Create(resFolder + "/" + "resources_cpu" + fileName + ".csv")
	if err != nil {
		log.Printf("LOG err creating file resource usage file%s\n", err )
		return
	}
	defer resf.Close()



	var podName string ="~'" + podNameRegEx + ".+'"

	var queryStr = "sum(rate(container_cpu_usage_seconds_total{image!='', pod=" + podName + "}[" + strconv.Itoa(returnRate) + "s])) by (pod)"
	//fmt.Printf( "query string: '%s'\n", queryStr);
	var resultJson = runQueryRange(duration,resStep,queryStr)
	var ro []resultObject


	err2 := json.Unmarshal(resultJson, &ro)
	if err2 != nil {
		log.Printf("GetCpuUsageByPOD %s", err2)
	}
	resf.WriteString("timestamp,%cpu usage\n")
	for k := range ro{
		//fmt.Printf( "POD: '%s'\n", ro[k].Metric);
		metricStr := fmt.Sprintf("POD: '%s'\n", ro[k].Metric)
		resf.WriteString(metricStr)
		//fmt.Printf( "The cpu usages were'%s'\n", ro[k].Values);
		//fmt.Println(reflect.TypeOf(ro[k].Values))
		bytes, err := ro[k].Values.MarshalJSON()
		if err != nil {
			panic(err)
		}
		var f interface{}
		err = json.Unmarshal(bytes, &f)

		m := f.([]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				//fmt.Println(k, "is an array:")
				//range on arrays and slices provides both the index and value for each entry.
				//for i, u := range vv {
				//	fmt.Print(i, u, "\n")
				//}
				for i, u := range vv {
					//fmt.Println(i, u)
					if i == 0 {
						var ts timeutil.Timestamp
						fTimeStamp, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}

						ts = timeutil.TimestampFromFloat64(fTimeStamp)
						tm := time.Unix(ts.Unix(), 0)
						resf.WriteString(tm.String() + ",")
						//fmt.Print(tm, ":")
					}
					if i == 1 {
						fCpu, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}
						var cpuUsed = ((math.Floor(fCpu*100)/100) *100)/float64(cpuLimit)
						//var cpuUsed = fCpu
						cpuUsedStr := fmt.Sprintf("%.2f", cpuUsed)
						//cpuUsedStr := fmt.Sprintf("%.2f", fCpu)
						resf.WriteString(cpuUsedStr + "\n")
						//fmt.Print(((math.Floor(fCpu*100)/100) *100)/float64(cpuLimit),"%", "\n")
					}
				}

			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
	}

	resf.Sync()
}
func GetMemoryUsageByPOD (duration int,returnRate int, resStep int, podNameRegEx string,resFolder string ){
	//(container_memory_usage_bytes{image="", pod="vod06dd2p"})[30m:1m]
	//create results folder if it does not exist and create resource usage file
	s := strings.Split(resFolder, "/")
	var fileName string = s[1]
	if _, create_dir_err := os.Stat(resFolder); os.IsNotExist(create_dir_err) {
		os.Mkdir(resFolder, os.FileMode(0755))
	}


	resf, err := os.OpenFile(resFolder + "/" + "resources_memory" + fileName + ".csv",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Printf("LOG err creating file resource usaae file%s\n", err )
		return
	}

	defer resf.Close()

	//var queryStr = "(container_memory_usage_bytes{image='', pod=~'" + podNameRegEx +".+'})"
	//since we are calling this with the exact name of the pod
	var queryStr = "(container_memory_usage_bytes{image='', pod=~'" + podNameRegEx +"'})"
	//fmt.Println("queryStr:",queryStr)
	var resultJson = runQueryRange(duration,resStep,queryStr)
	var ro []resultObject


	err2 := json.Unmarshal(resultJson, &ro)
	if err2 != nil {
		log.Printf("GetMemoryUsageByPOD %s", err2)
	}
	resf.WriteString("timestamp,memory usage(bytes)\n")
	for k := range ro{

		testbytes, err := ro[k].Metric.MarshalJSON()
		if err != nil {
			panic(err)
		}

		var metricIf map[string]interface {}
		err = json.Unmarshal(testbytes, &metricIf)
		//fmt.Println(reflect.TypeOf(metricIf))

		var testStr string = metricIf["pod"].(string)
		resf.WriteString(testStr + "\n")


		bytes, err := ro[k].Values.MarshalJSON()
		if err != nil {
			panic(err)
		}
		//fmt.Println(reflect.TypeOf(bytes))

		var f interface{}
		err = json.Unmarshal(bytes, &f)
		//fmt.Println(reflect.TypeOf(f))
		m := f.([]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				//range on arrays and slices provides both the index and value for each entry.
				for i, u := range vv {
					//fmt.Println(i, u)
					if(i == 0){
						var ts timeutil.Timestamp
						fTimeStamp, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}

						ts = timeutil.TimestampFromFloat64(fTimeStamp)
						tm := time.Unix(ts.Unix(), 0)
						resf.WriteString(tm.String() + ",")
						//fmt.Print(tm, ":")
					}
					if(i == 1){
						//memGauge, err := strconv.ParseInt(fmt.Sprintf("%v", u),10, 64)
						memGauge, err := strconv.ParseInt(fmt.Sprintf("%v", u),10, 64)
						if err != nil {
							panic(err)
						}
						memUsedStr := fmt.Sprintf("%d", memGauge)
						resf.WriteString(memUsedStr + "\n")
					}
				}

			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
	}

	resf.Sync()
}

func GetCPUThrottleInfoByPOD (duration int,returnRate int, resStep int, podNameRegEx string,resFolder string ){

	//create results folder if it does not exist and create resource usage file
	s := strings.Split(resFolder, "/")
	var fileName string = s[1]
	if _, create_dir_err := os.Stat(resFolder); os.IsNotExist(create_dir_err) {
		os.Mkdir(resFolder, os.FileMode(0755))
	}
	resf, err := os.Create(resFolder + "/" + "cpu_throttle" + fileName + ".csv")
	if err != nil {
		log.Printf("LOG err creating file resource usaae file%s\n", err )
		return
	}
	defer resf.Close()



	var podName string ="~'" + podNameRegEx + ".+'"

	var queryStr = "sum(rate(container_cpu_cfs_throttled_seconds_total{image!='', pod=" + podName + "}[" + strconv.Itoa(returnRate) + "s])) by (pod)"
	//fmt.Println("queryStr:",queryStr)

	var resultJson = runQueryRange(duration,resStep,queryStr)
	var ro []resultObject


	err2 := json.Unmarshal(resultJson, &ro)
	if err2 != nil {
		log.Printf("GetCPUThrottleInfoByPOD %s", err2)
	}
	resf.WriteString("timestamp,%num throtled\n")
	for k := range ro{

		metricStr := fmt.Sprintf("POD: '%s'\n", ro[k].Metric)
		resf.WriteString(metricStr)

		bytes, err := ro[k].Values.MarshalJSON()
		if err != nil {
			panic(err)
		}
		var f interface{}
		err = json.Unmarshal(bytes, &f)

		m := f.([]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				for i, u := range vv {
					//fmt.Println(i, u)
					if(i == 0){
						var ts timeutil.Timestamp
						fTimeStamp, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}

						ts = timeutil.TimestampFromFloat64(fTimeStamp)
						tm := time.Unix(ts.Unix(), 0)
						resf.WriteString(tm.String() + ",")
						//fmt.Print(tm, ":")
					}
					if(i == 1){
						fCpu, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}

						cpuUsedStr := fmt.Sprintf("%.2f", fCpu)

						resf.WriteString(cpuUsedStr + "\n")
					}
				}
			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
	}

	resf.Sync()
}

func GetCPUUsageByInstance(duration int,returnRate int, resStep int,resFolder string ){

	//create results folder if it does not exist and create resource usage file
	s := strings.Split(resFolder, "/")
	var fileName string = s[1]
	if _, create_dir_err := os.Stat(resFolder); os.IsNotExist(create_dir_err) {
		os.Mkdir(resFolder, os.FileMode(0755))
	}
	resf, err := os.Create(resFolder + "/" + "cpu_instance" + fileName + ".csv")
	if err != nil {
		log.Printf("LOG err creating file resource usage file%s\n", err )
		return
	}
	defer resf.Close()

	//100 - (avg by (instance) (irate(node_cpu_seconds_total{job="node-exporter",mode="idle"}[5m])) * 100)
	//var queryStr = "100 - (avg by (instance) (irate(node_cpu_seconds_total{job='node-exporter',mode='idle'}[5m])) * 100)"
	var queryStr = "100 - (avg by (instance) (irate(node_cpu_seconds_total{job='node-exporter',mode='idle'}[" + strconv.Itoa(returnRate) + "m])) * 100)"

	var resultJson = runQueryRange(duration,resStep,queryStr)
	var ro []resultObject


	err2 := json.Unmarshal(resultJson, &ro)
	if err2 != nil {
		log.Printf("GetCPUThrottleInfoByPOD %s", err2)
	}
	resf.WriteString("timestamp,%cpu usage\n")
	for k := range ro{
		//fmt.Printf( "POD: '%s'\n", ro[k].Metric);
		metricStr := fmt.Sprintf("POD: '%s'\n", ro[k].Metric)
		resf.WriteString(metricStr)
		//fmt.Printf( "The cpu usages were'%s'\n", ro[k].Values);
		//fmt.Println(reflect.TypeOf(ro[k].Values))
		bytes, err := ro[k].Values.MarshalJSON()
		if err != nil {
			panic(err)
		}
		var f interface{}
		err = json.Unmarshal(bytes, &f)

		m := f.([]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				//fmt.Println(k, "is an array:")
				//range on arrays and slices provides both the index and value for each entry.
				//for i, u := range vv {
				//	fmt.Print(i, u, "\n")
				//}
				for i, u := range vv {
					//fmt.Println(i, u)
					if(i == 0){
						var ts timeutil.Timestamp
						fTimeStamp, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}

						ts = timeutil.TimestampFromFloat64(fTimeStamp)
						tm := time.Unix(ts.Unix(), 0)
						resf.WriteString(tm.String() + ",")
						//fmt.Print(tm, ":")
					}
					if(i == 1){
						fCpu, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}
						//var cpuUsed = ((math.Floor(fCpu*100)/100) *100)/float64(cpuLimit)
						//var cpuUsed = fCpu
						cpuUsedStr := fmt.Sprintf("%.2f", fCpu)
						//cpuUsedStr := fmt.Sprintf("%.2f", fCpu)
						resf.WriteString(cpuUsedStr + "\n")
					}
				}

			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
	}

	resf.Sync()
}
func TimestampFromFloat64(ts float64) timeutil.Timestamp {
	secs := int64(ts)
	nsecs := int64((ts - float64(secs)) * 1e9)
	return timeutil.Timestamp{time.Unix(secs, nsecs)}
}
func ExampleAPI_query() {
	client, err := api.NewClient(api.Config{
		Address: promServerUrl,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//expression 0.1 is equivalent to the expression 100m, which can be read as “one hundred millicpu” OR "one hundred millicores”
	//1000 milliCPU is 1 CPU
	// You can use the suffix m to mean milli. For example 100m CPU, 100 milliCPU, and 0.1 CPU are all the same. Precision finer than 1m is not allowed.
	//CPU is always requested as an absolute quantity, never as a relative quantity; 0.1 is the same amount of CPU on a single-core, dual-core, or 48-core machine.
	result, warnings, err := api.Query(ctx, "rate(node_context_switches_total{job='node-exporter'}[5m])[30m:1m]", time.Now().Add(-(time.Minute)*50))
	//result, warnings, err := api.Query(ctx, "rate(node_context_switches_total{job='node-exporter'})", time.Now().Add(-(time.Minute)*5))
	//result, warnings, err := api.Query(ctx, "100 - (avg by (instance) (irate(node_cpu_seconds_total{job='node',mode='idle'}[5m])) * 100)", time.Now())
	//result, warnings, err := api.Query(ctx, "sum(kube_pod_container_resource_limits_cpu_cores{}) by (pod, namespace)", time.Now())
	//result, warnings, err := api.Query(ctx, "100 - (avg by (instance) (irate(node_cpu_seconds_total{job='node',mode='idle'}[5m])) * 100)", time.Now())

	//sum by (mode, instance) (irate(node_cpu_seconds_total{job="node"}[5m]))

	//sum (rate (container_cpu_usage_seconds_total{image!=""}[90m])) by (pod)
	//.Add(-(time.Minute)* 2))
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
}

/*type resultObject struct {
	Metric json.RawMessage `json:"metric"`
	Values []Value `json:"values"`
}

type Value struct {
	Timestamp float64
	CpuUsage  float64
}
func (tp *Value) UnmarshalJSON(data []byte) error {
	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Printf("Error while decoding %v\n", err)
		return err
	}
	tp.Timestamp = v[0].(float64)
	tp.CpuUsage, _ = strconv.ParseFloat(v[1].(string), 64)

	return nil
}*/
func runQueryRange(duration int, resolutionStep int,queryStr string) []byte {
	//rate(http_requests_total[5m])[30m:1m]  :: this is: Return the 5-minute rate of the http_requests_total metric for the past 30 minutes, with a resolution of 1 minute.

	client, err := api.NewClient(api.Config{
		Address: promServerUrl,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := v1.Range{
		Start: time.Now().Add(-(time.Second) * time.Duration(duration)),
		End:   time.Now(),
		Step:  (time.Second)* time.Duration(resolutionStep),//range in query should not be smaller than this
	}
	//a Counter is a single, monotonically increasing, cumulative metric.
	//A single metric means, that a Counter represents a single value, e.g. the number of orders created in a shop system.
	//It’s monotonically increasing, so it can only increase, usually one-by-one. It’s a cumulative metric, so it always contains the overall value.
	//_total is the conventional postfix for counters in Prometheus
	//vector of values over time (called instant vector)
	//A range vector can be seen as a continuous subset of the instant vector
	//The range is defined in square brackets and appended to the instant vector selector
	//container_cpu_usage_seconds_total(for an accumulating count with unit)
	//The rate(v range-vector)[time]function takes a range vector. it calculates the average over time
	/*We're using the container_cpu_usage_seconds_total metric to calculate Pod CPU usage. This metrics contains the total amount of CPU seconds consumed by container
	by core (this is important, as a Pod may consist of multiple containers, each of which can be scheduled across multiple cores; however, the metric has a pod annotation
	that we can use for aggregation). Of special interest is the change rate of that metric (which can be calculated with PromQL's rate() function).
	If it increases by 1 within one second, the Pod consumes 1 CPU core (or 1000 milli-cores) in that second.
	*/
	//pod=~'vod.+'
	//result, warnings, err := api.QueryRange(ctx, "sum(rate(container_cpu_usage_seconds_total{image!='',pod='vod0jdw4n'}[15m])) by (pod)", r)
	result, warnings, err := api.QueryRange(ctx, queryStr, r)//the range is set depending on scrape interval
	//result, warnings, err := api.QueryRange(ctx, "sum(rate(container_cpu_usage_seconds_total{image!='',pod=~'vod.+'}[1m])) by (pod)", r)//the range is set depending on scrape interval
	//result, warnings, err := api.QueryRange(ctx, "rate(container_cpu_usage_seconds_total{image!='',pod=~'vod.+'}[5m])", r)//the range is set depending on scrape interval
	//result, warnings, err := api.QueryRange(ctx, "sum(sum by (container_name)( rate(container_cpu_usage_seconds_total{image!=''}[5m] ) )) / count(node_cpu_seconds_total{mode='system'}) * 100", r)

	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	//fmt.Printf("Result:\n%v\n", result)
	//fmt.Printf("Test:\n%v\n", result.String())


	resultJson, err := json.Marshal(result)
	//fmt.Println(string(resultJson))
	if err != nil {
		fmt.Println("error:", err)
	}
	return resultJson
}

func ExampleAPI_queryRange() {
	//rate(http_requests_total[5m])[30m:1m]  :: this is: Return the 5-minute rate of the http_requests_total metric for the past 30 minutes, with a resolution of 1 minute.
	type resultObject struct {
		Metric json.RawMessage `json:"metric"`
		Values json.RawMessage `json:"values"`
	}


	client, err := api.NewClient(api.Config{
		Address: promServerUrl,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := v1.Range{
		Start: time.Now().Add(-(time.Minute)*15),
		End:   time.Now(),
		Step:  time.Second,//range should not be smaller than this
	}
	//a Counter is a single, monotonically increasing, cumulative metric.
	//A single metric means, that a Counter represents a single value, e.g. the number of orders created in a shop system.
	//It’s monotonically increasing, so it can only increase, usually one-by-one. It’s a cumulative metric, so it always contains the overall value.
	//_total is the conventional postfix for counters in Prometheus
	//vector of values over time (called instant vector)
	//A range vector can be seen as a continuous subset of the instant vector
	//The range is defined in square brackets and appended to the instant vector selector
	//container_cpu_usage_seconds_total(for an accumulating count with unit)
	//The rate(v range-vector)[time]function takes a range vector. it calculates the average over time
	/*We're using the container_cpu_usage_seconds_total metric to calculate Pod CPU usage. This metrics contains the total amount of CPU seconds consumed by container
	by core (this is important, as a Pod may consist of multiple containers, each of which can be scheduled across multiple cores; however, the metric has a pod annotation
	that we can use for aggregation). Of special interest is the change rate of that metric (which can be calculated with PromQL's rate() function).
	If it increases by 1 within one second, the Pod consumes 1 CPU core (or 1000 milli-cores) in that second.
	*/
	//pod=~'vod.+'
	//result, warnings, err := api.QueryRange(ctx, "sum(rate(container_cpu_usage_seconds_total{image!='',pod='vod0jdw4n'}[15m])) by (pod)", r)
	//result, warnings, err := api.QueryRange(ctx, "sum(rate(container_cpu_usage_seconds_total{image!='',pod=~'vod.+'}[1m])) by (pod)", r)//the range is set depending on scrape interval
	//result, warnings, err := api.QueryRange(ctx, "rate(container_cpu_usage_seconds_total{image!='',pod=~'vod.+'}[5m])", r)//the range is set depending on scrape interval
	//result, warnings, err := api.QueryRange(ctx, "sum(sum by (container_name)( rate(container_cpu_usage_seconds_total{image!=''}[5m] ) )) / count(node_cpu_seconds_total{mode='system'}) * 100", r)
	result, warnings, err := api.QueryRange(ctx, "rate(node_context_switches_total{job='node-exporter'}[5m])", r)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
	//fmt.Printf("Test:\n%v\n", result.String())


	resultJson, _ := json.Marshal(result)
	fmt.Println(string(resultJson))

	var ro []resultObject

	err2 := json.Unmarshal(resultJson, &ro)
	if err2 != nil {
		fmt.Println("error:", err2)
		os.Exit(1)
	}

	for k := range ro{
		fmt.Printf( "POD: '%s'\n", ro[k].Metric);
		//fmt.Printf( "The cpu usages were'%s'\n", ro[k].Values);
		//fmt.Printf(reflect.TypeOf(ro[k].Values))
		//fmt.Println(reflect.TypeOf(ro[k].Values))
		bytes, err := ro[k].Values.MarshalJSON()
		if err != nil {
			panic(err)
		}
		var f interface{}
		err = json.Unmarshal(bytes, &f)

		m := f.([]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				//fmt.Println(k, "is an array:")
				//range on arrays and slices provides both the index and value for each entry.
				//for i, u := range vv {
				//	fmt.Print(i, u, "\n")
				//}
				for i, u := range vv {
					//fmt.Println(i, u)
					if(i == 0){
						var ts timeutil.Timestamp
						fTimeStamp, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}

						ts = timeutil.TimestampFromFloat64(fTimeStamp)
						tm := time.Unix(ts.Unix(), 0)
						fmt.Print(tm, ":")
					}
					if(i == 1){
						fCpu, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}
						fmt.Print(((math.Floor(fCpu*100)/100) *100)/4,"%", "\n")
					}
				}

			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
	}
}

func GetContextSwitchesByNode (duration int,returnRate int, resStep int, resFolder string ){

	s := strings.Split(resFolder, "/")
	var fileName string = s[1]
	if _, create_dir_err := os.Stat(resFolder); os.IsNotExist(create_dir_err) {
		os.Mkdir(resFolder, os.FileMode(0755))
	}
	resf, err := os.Create(resFolder + "/" + "context_switches" + fileName + ".csv")
	if err != nil {
		log.Printf("LOG err creating file context_switches file%s\n", err )
		return
	}
	defer resf.Close()

	var queryStr = "rate(node_context_switches_total{job='node-exporter'}[5m])"

	var resultJson = runQueryRange(duration,resStep,queryStr)
	var ro []resultObject

	err2 := json.Unmarshal(resultJson, &ro)
	if err2 != nil {
		log.Printf("GetContextSwitchesByNode %s", err2)
	}

	resf.WriteString("timestamp,context switches\n")
	for k := range ro{

		//testbytes, err := ro[k].Metric.MarshalJSON()
		if err != nil {
			panic(err)
		}

		bytes, err := ro[k].Values.MarshalJSON()
		if err != nil {
			panic(err)
		}
		//fmt.Println(reflect.TypeOf(bytes))

		var f interface{}
		err = json.Unmarshal(bytes, &f)

		m := f.([]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				//range on arrays and slices provides both the index and value for each entry.
				for i, u := range vv {
					if(i == 0){
						var ts timeutil.Timestamp
						fTimeStamp, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}

						ts = timeutil.TimestampFromFloat64(fTimeStamp)
						tm := time.Unix(ts.Unix(), 0)
						resf.WriteString(tm.String() + ",")
					}
					if(i == 1){
						memGauge, err := strconv.ParseFloat(fmt.Sprintf("%v", u),64)

						if err != nil {
							panic(err)
						}
						memUsedStr := fmt.Sprintf("%f", memGauge)
						resf.WriteString(memUsedStr + "\n")
						//fmt.Println(memUsedStr)
					}
				}

			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
	}

	resf.Sync()
}
func GetCPUUsageByNode(duration int) {
	//rate(http_requests_total[5m])[30m:1m]  :: this is: Return the 5-minute rate of the http_requests_total metric for the past 30 minutes, with a resolution of 1 minute.
	type resultObject struct {
		Metric json.RawMessage `json:"metric"`
		Values json.RawMessage `json:"values"`
	}


	client, err := api.NewClient(api.Config{
		Address: promServerUrl,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := v1.Range{
		Start: time.Now().Add(-(time.Minute)*15),
		End:   time.Now(),
		Step:  time.Second*30,//range should not be smaller than this
	}
	//a Counter is a single, monotonically increasing, cumulative metric.
	//A single metric means, that a Counter represents a single value, e.g. the number of orders created in a shop system.
	//It’s monotonically increasing, so it can only increase, usually one-by-one. It’s a cumulative metric, so it always contains the overall value.
	//_total is the conventional postfix for counters in Prometheus
	//vector of values over time (called instant vector)
	//A range vector can be seen as a continuous subset of the instant vector
	//The range is defined in square brackets and appended to the instant vector selector
	//container_cpu_usage_seconds_total(for an accumulating count with unit)
	//The rate(v range-vector)[time]function takes a range vector. it calculates the average over time
	/*We're using the container_cpu_usage_seconds_total metric to calculate Pod CPU usage. This metrics contains the total amount of CPU seconds consumed by container
	by core (this is important, as a Pod may consist of multiple containers, each of which can be scheduled across multiple cores; however, the metric has a pod annotation
	that we can use for aggregation). Of special interest is the change rate of that metric (which can be calculated with PromQL's rate() function).
	If it increases by 1 within one second, the Pod consumes 1 CPU core (or 1000 milli-cores) in that second.
	*/
	//pod=~'vod.+'
	//result, warnings, err := api.QueryRange(ctx, "sum(rate(container_cpu_usage_seconds_total{image!='',pod='vod0jdw4n'}[15m])) by (pod)", r)
	var queryString string = "100 - (avg by (instance) (irate(node_cpu_seconds_total{job='node-exporter',mode='idle'}[" + strconv.Itoa(duration) + "s])) * 100)"
	//result, warnings, err := api.QueryRange(ctx, "100 - (avg by (instance) (irate(node_cpu_seconds_total{job='node-exporter',mode='idle'}[15m])) * 100)", r)
	result, warnings, err := api.QueryRange(ctx, queryString, r)//the range is set depending on scrape interval
	//result, warnings, err := api.QueryRange(ctx, "rate(container_cpu_usage_seconds_total{image!='',pod=~'vod.+'}[5m])", r)//the range is set depending on scrape interval
	//result, warnings, err := api.QueryRange(ctx, "sum(sum by (container_name)( rate(container_cpu_usage_seconds_total{image!=''}[5m] ) )) / count(node_cpu_seconds_total{mode='system'}) * 100", r)

	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
	//fmt.Printf("Test:\n%v\n", result.String())


	resultJson, _ := json.Marshal(result)
	fmt.Println(string(resultJson))

	var ro []resultObject

	err2 := json.Unmarshal(resultJson, &ro)
	if err2 != nil {
		fmt.Println("error:", err2)
		os.Exit(1)
	}

	for k := range ro{
		fmt.Printf( "POD: '%s'\n", ro[k].Metric);
		//fmt.Printf( "The cpu usages were'%s'\n", ro[k].Values);
		//fmt.Printf(reflect.TypeOf(ro[k].Values))
		//fmt.Println(reflect.TypeOf(ro[k].Values))
		bytes, err := ro[k].Values.MarshalJSON()
		if err != nil {
			panic(err)
		}
		var f interface{}
		err = json.Unmarshal(bytes, &f)

		m := f.([]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				//fmt.Println(k, "is an array:")
				//range on arrays and slices provides both the index and value for each entry.
				//for i, u := range vv {
				//	fmt.Print(i, u, "\n")
				//}
				for i, u := range vv {
					//fmt.Println(i, u)
					if(i == 0){
						var ts timeutil.Timestamp
						fTimeStamp, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}

						ts = timeutil.TimestampFromFloat64(fTimeStamp)
						tm := time.Unix(ts.Unix(), 0)
						fmt.Print(tm, ":")
					}
					if(i == 1){
						fCpu, err := strconv.ParseFloat(fmt.Sprintf("%v", u), 64)
						if err != nil {
							panic(err)
						}
						fmt.Print(((math.Floor(fCpu*100)/100)),"%", "\n")
					}
				}

			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
	}
}
func ExampleAPI_series() {
	client, err := api.NewClient(api.Config{
		Address: promServerUrl,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	lbls, warnings, err := api.Series(ctx, []string{
		"{__name__=~\"scrape_.+\",job=\"node\"}",
		"{__name__=~\"scrape_.+\",job=\"prometheus\"}",
	}, time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Println("Result:")
	for _, lbl := range lbls {
		fmt.Println(lbl)
	}
}
