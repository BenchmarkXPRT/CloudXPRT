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

package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/olekukonko/tablewriter"
)

var (
	clients      int
	urlPath      string
	urlsFilePath string
	clientStep   int
	clientEnd    int
	sla          int
	timeInterval int
	expResult    string
)

type Result struct {
	clients         int
	requests        int
	success         int
	networkFailed   int
	badFailed       int
	mismatched      int
	rate            float64
	readThroughput  int
	writeThroughput int
	cpu             int
	serviceName     []string
	serviceResp     []string
	elapsed         []string
	apdexScore      []string
}

const (
	maxRetry     = 5
	showLocalCPU = false
	onCloud      = false
	masterN      = "master"
)

var OCRSLA = []int{1500, 2000, 2500}
var OCRTHROUGHPUT = []int{0, 0, 0}

var (
	results     []*Result
	prevIdle    uint64
	prevNonIdle uint64
	cpuUsage    []int
	masterNodes []string
	nodeCPU     map[string]int
	title       string
	command     string
	logFile     string
	maxReq      = 0
	retry       = 0
	DEBUG       = true
	aveCPU      = 0
)

func init() {
	flag.IntVar(&clients, "c", 100, "Number of start concurrent clients")
	flag.StringVar(&urlPath, "u", "", "URL")
	flag.StringVar(&urlsFilePath, "f", "", "URL's file path (line seperated)")
	flag.IntVar(&clientStep, "ci", 100, "Client number increase step")
	flag.IntVar(&clientEnd, "cl", -1, "Last client number to run tests")
	flag.IntVar(&sla, "s", -1, "Service level agreement (in milliseconds)")
	flag.IntVar(&timeInterval, "ti", 120, "Time interval between tests (in seconds)")
	flag.StringVar(&expResult, "e", "", "Expected string pattern from response")
}

func printResults(startTime time.Time) {
	var buf bytes.Buffer
	buf.WriteString("\n\nResults Summary:\n")

	elapsed := int64(time.Since(startTime).Seconds())

	if elapsed == 0 {
		elapsed = 1
	}

	for _, result := range results {
		buf.WriteString(fmt.Sprintf("Clients:                        %10d clts\n", result.clients))
		buf.WriteString(fmt.Sprintf("Requests:                       %10d hits\n", result.requests))
		buf.WriteString(fmt.Sprintf("Successful requests:            %10d hits\n", result.success))
		buf.WriteString(fmt.Sprintf("Network failed:                 %10d hits\n", result.networkFailed))
		buf.WriteString(fmt.Sprintf("Bad requests failed (!2xx):     %10d hits\n", result.badFailed))
		buf.WriteString(fmt.Sprintf("Pattern mismatch:               %10d hits\n", result.mismatched))
		buf.WriteString(fmt.Sprintf("Successful requests rate:       %10.2f hits/sec\n", result.rate))
		buf.WriteString(fmt.Sprintf("Read throughput:                %10d bytes/sec\n", result.readThroughput))
		buf.WriteString(fmt.Sprintf("Write throughput:               %10d bytes/sec\n", result.writeThroughput))
		buf.WriteString(fmt.Sprintf("Average CPU usage:              %10d %%\n", result.cpu))
		buf.WriteString(fmt.Sprintf("===========================================================\n"))
	}

	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("Total Test Time:                %10d sec\n\n", elapsed))
	outputToStdout(buf.String())

	// create log file
	logFile = getRandomFileName(".log")
	f, err := os.Create(logFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	// create csv file
	csvFile := getRandomFileName(".csv")
	cf, err := os.Create(csvFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer cf.Close()
	csvWriter := csv.NewWriter(cf)

	table := tablewriter.NewWriter(f)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	// change titles NET_FAILED to FAIL_REQS, BAD_REQS to RESP_MISMATCH
	header := []string{"CONCURRENCY", "REQUESTS", "SUCC_REQS", "FAIL_REQS", "RESP_MISMATCH",
		"SUCC_REQS_RATE(REQ/S)", "READ_TP(B/S)", "WRITE_TP(B/S)", "AVE_CPU_USAGE(%)", "TIME(S)"}
	//"SUCC_REQS_RATE(REQ/S)", "READ_TP(B/S)", "WRITE_TP(B/S)", "TIME(S)"}
	csvHeader := []string{"CONCURRENCY", "REQUESTS", "SUCC_REQS", "FAIL_REQS", "RESP_MISMATCH",
		"SUCC_REQS_RATE(REQ/S)", "READ_TP(B/S)", "WRITE_TP(B/S)", "AVE_CPU_USAGE(%)", "TIME(S)"}
	if len(results) > 0 {
		for i := 0; i < len(results[0].elapsed); i++ {
			header = append(header, strings.ToUpper(results[0].serviceName[i])+"_RESP_TIME(95%ile)(MS)")
			if len(results[0].elapsed) > 1 {
				csvHeader = append(csvHeader, strings.ToUpper(results[0].serviceName[i])+"_REQS")
			}
			csvHeader = append(csvHeader, strings.ToUpper(results[0].serviceName[i])+"_RESP_TIME(95%ile)(MS)")
			if len(results[0].apdexScore) > 0 {
				header = append(header, strings.ToUpper(results[0].serviceName[i])+"_APDEX")
				csvHeader = append(csvHeader, strings.ToUpper(results[0].serviceName[i])+"_APDEX")
			}
		}
	}
	csvWriter.Write(csvHeader)
	table.SetHeader(header)
	table.SetAutoFormatHeaders(false)
	var maxResult = &Result{}
	for _, result := range results {
		contents := []string{fmt.Sprintf("%d", result.clients),
			fmt.Sprintf("%d", result.requests),
			fmt.Sprintf("%d", result.success),
			fmt.Sprintf("%d", result.networkFailed+result.badFailed),
			fmt.Sprintf("%d", result.mismatched),
			fmt.Sprintf("%.2f", result.rate),
			fmt.Sprintf("%d", result.readThroughput),
			fmt.Sprintf("%d", result.writeThroughput),
			fmt.Sprintf("%d", result.cpu),
			fmt.Sprintf("%d", timeInterval)}

		if maxReq == result.success {
			maxResult = result
		}

		csvCont := make([]string, len(contents))
		copy(csvCont, contents)

		for idx, time := range result.elapsed {
			var resptime string
			if len(result.elapsed) == 1 {
				//resptime = fmt.Sprintf("%s (%.5f)", time,
				//	float32(atoi(result.serviceResp[idx]))/float32(atoi(time)))
				resptime = fmt.Sprintf("%s", time)
				csvCont = append(csvCont, time)
			} else {
				//resptime = fmt.Sprintf("%s (%.5f) %s", time,
				//	float32(atoi(result.serviceResp[idx]))/float32(atoi(time)), result.serviceResp[idx])
				resptime = fmt.Sprintf("%s %s", time, result.serviceResp[idx])
				csvCont = append(csvCont, result.serviceResp[idx])
				csvCont = append(csvCont, time)
			}
			contents = append(contents, resptime)

			if len(result.apdexScore) > 0 {
				contents = append(contents, result.apdexScore[idx])
				csvCont = append(csvCont, result.apdexScore[idx])
			}
		}
		csvWriter.Write(csvCont)
		table.Append(contents)
	}

	// Add the test summary part
	if maxResult.rate > 0 {
		table.SetCaption(true, fmt.Sprintf("Best throughput found at %.2f requests per second with 95th percentile latency of %s ms",
			maxResult.rate, maxResult.elapsed[0]))
	}

	csvWriter.Flush()
	table.Render()
}

func atoi(input string) int {
	temp, err := strconv.Atoi(input)
	if err != nil {
		log.Fatal(err.Error())
	}
	if temp <= 0 {
		log.Fatalf("Invalid response time found %s", input)
	}
	return temp
}

func inputCheck() {
	if urlsFilePath == "" && urlPath == "" {
		outputToStdout("URL or URL file must be provided")
		flag.Usage()
		os.Exit(1)
	}

	if timeInterval <= 0 {
		outputToStdout(" period must be provided")
		flag.Usage()
		os.Exit(1)
	}

	if clientEnd != -1 && clientEnd <= clients {
		outputToStdout("Last client number must be bigger than start client number")
		flag.Usage()
		os.Exit(1)
	}
}

func parseOutput(out string, clients int, aveCPU int) ([]string, bool) {
	result := &Result{}
	result.clients = clients
	result.cpu = aveCPU
	if DEBUG {
		outputToStdout(out)
	}
	outArray := strings.Split(out, "\n")
	for _, line := range outArray {
		if strings.Contains(line, "Requests:") {
			tokens := strings.Fields(line)
			temp, err := strconv.Atoi(tokens[1])
			if err != nil {
				log.Fatal(err.Error())
			}
			result.requests = temp
		}
		if strings.Contains(line, "Successful requests:") {
			tokens := strings.Fields(line)
			temp, err := strconv.Atoi(tokens[2])
			if err != nil {
				log.Fatal(err.Error())
			}
			result.success = temp
			if temp > maxReq {
				maxReq = temp
				retry = 0
			} else {
				retry++
			}
		}
		if strings.Contains(line, "Network failed:") {
			tokens := strings.Fields(line)
			temp, err := strconv.Atoi(tokens[2])
			if err != nil {
				log.Fatal(err.Error())
			}
			result.networkFailed = temp
		}
		if strings.Contains(line, "Bad requests failed") {
			tokens := strings.Fields(line)
			temp, err := strconv.Atoi(tokens[4])
			if err != nil {
				log.Fatal(err.Error())
			}
			result.badFailed = temp
		}
		if strings.Contains(line, "Pattern mismatch") {
			tokens := strings.Fields(line)
			temp, err := strconv.Atoi(tokens[2])
			if err != nil {
				log.Fatal(err.Error())
			}
			result.mismatched = temp
		}
		if strings.Contains(line, "Successful requests rate:") {
			tokens := strings.Fields(line)
			temp, err := strconv.ParseFloat(tokens[3], 32)
			if err != nil {
				log.Fatal(err.Error())
			}
			result.rate = temp
		}
		if strings.Contains(line, "Read throughput:") {
			tokens := strings.Fields(line)
			temp, err := strconv.Atoi(tokens[2])
			if err != nil {
				log.Fatal(err.Error())
			}
			result.readThroughput = temp
		}
		if strings.Contains(line, "Write throughput:") {
			tokens := strings.Fields(line)
			temp, err := strconv.Atoi(tokens[2])
			if err != nil {
				log.Fatal(err.Error())
			}
			result.writeThroughput = temp
		}
		if strings.Contains(line, "For URL:") {
			result.serviceName = append(result.serviceName, getServiceName(line))
		}
		if strings.Contains(line, "Total") {
			tokens := strings.Fields(line)
			result.serviceResp = append(result.serviceResp, tokens[1])
		}
		if strings.Contains(line, "95%") {
			tokens := strings.Fields(line)
			result.elapsed = append(result.elapsed, tokens[1])
		}
		if strings.Contains(line, "Apdex") {
			tokens := strings.Fields(line)
			result.apdexScore = append(result.apdexScore, tokens[3])
		}
	}
	results = append(results, result)

	// If network failed is more than 10% of total requests, stop the tests
	return result.elapsed, float32(result.networkFailed+result.badFailed+result.mismatched)/float32(result.requests) > 0.1
}

func cpuProfile() {
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		log.Fatal("stat read fail")
	}

	idle := stat.CPUStatAll.Idle + stat.CPUStatAll.IOWait
	nonIdle := stat.CPUStatAll.User + stat.CPUStatAll.Nice + stat.CPUStatAll.System + stat.CPUStatAll.IRQ +
		stat.CPUStatAll.SoftIRQ + stat.CPUStatAll.Steal

	if prevIdle > 0 && prevNonIdle > 0 {
		prevTotal := prevNonIdle + prevIdle
		total := idle + nonIdle

		// differentiate: actual value minus the previous one
		totalld := total - prevTotal
		idled := idle - prevIdle

		outputToStdout(fmt.Sprintf("CPU Usage: %d", (totalld-idled)*100/totalld))
		cpuUsage = append(cpuUsage, int((totalld-idled)*100/totalld))
	}

	prevIdle = idle
	prevNonIdle = nonIdle
}

// Send output to both stdout
func outputToStdout(output string) {
	fmt.Println(output)
}

func getTitle() {
	if len(title) == 0 {
		if len(urlsFilePath) == 0 {
			index1 := strings.LastIndex(urlPath, "/")
			index2 := strings.LastIndex(urlPath, ":")
			if index2 > index1 {
				title = urlPath[index2+1:]
			} else {
				title = urlPath[index1+1:]
			}
		} else {
			index := strings.LastIndex(urlsFilePath, ".")
			title = urlsFilePath[:index]
		}
	}
}

func getServiceName(url string) string {
	// Trim leading and trailing white spaces
	urlTmp := strings.TrimSpace(url)
	// Trim extra "/" if any
	urlTmp = strings.TrimSuffix(urlTmp, "/")

	// URL could have 2 formats, need consider both case
	// 1. http://webserviceIP:8070/ocr
	// 2. http://ocrserviceIP:8073
	index1 := strings.LastIndex(urlTmp, "/")
	index2 := strings.LastIndex(urlTmp, ":")
	if index2 > index1 {
		return urlTmp[index2+1:]
	}
	return urlTmp[index1+1:]
}

func getRandomFileName(ext string) string {
	timeNow := time.Now().Format("20060102150405")
	exePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	index := 8
	newTime := timeNow[:index] + "_" + timeNow[index:]
	return exePath + "/output/autoloader_" + title + "_" + newTime + ext
}

func makeOutputDirectory() {
	exePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	// ignore error if the output directory is already there
	_ = os.Mkdir(exePath+"/output", 0777)
}

func checkSLA(elapsed []string) bool {
	temp, err := strconv.Atoi(elapsed[0])
	if err != nil {
		log.Fatal(err.Error())
	}

	return temp > sla
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func getNodesList() {
	out, err := exec.Command("kubectl", "get", "nodes").Output()
	if err != nil || len(out) == 0 {
		return
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "NAME") {
			continue
		}
		tokens := strings.Fields(line)
		if strings.TrimSpace(tokens[2]) == masterN {
			masterNodes = append(masterNodes, tokens[0])
		}
		nodeCPU[tokens[0]] = getNodeCPU(tokens[0])
	}
	// fmt.Println(nodeCPU)
}

func getNodeCPU(nodename string) int {
	var cmd = "kubectl describe node " + nodename + " | grep 'cpu:'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err.Error())
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		tokens := strings.Fields(line)
		temp, err := strconv.Atoi(tokens[1])
		if err != nil {
			log.Fatal(err.Error())
		}
		if temp > 0 {
			// get the first CPU value and return
			return temp
		}
	}
	log.Fatalf("No CPU value found for node %s", nodename)
	return 0
}

func getAverageCPU(input string) int {
	lines := strings.Split(input, "\n")
	var count = 0
	var total = 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "NAME") {
			continue
		}
		tokens := strings.Fields(line)
		if onCloud {
			if find(masterNodes, tokens[0]) {
				continue
			}
		}
		// The formula is (perc1*cpu1+perc2*cpu2)/(cpu1+cpu2)
		temp, err := strconv.Atoi(strings.TrimSuffix(tokens[2], "%"))
		if err == nil {
			total += temp * nodeCPU[tokens[0]]
			count += nodeCPU[tokens[0]]
		}
	}

	var aveCPU = 0
	if count > 0 {
		aveCPU = (int)(total / count)
	}
	return aveCPU
}

func main() {
	var curIndex = 0
	var err error
	nodeCPU = make(map[string]int)
	makeOutputDirectory()
	getNodesList()

	flag.Parse()
	inputCheck()

	ticker := time.NewTicker(10 * time.Second)
	tickerCPU := time.NewTicker(30 * time.Second)
	startTime := time.Now()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		_ = <-signalChannel
		ticker.Stop()
		outputToStdout("########## Tests are interrupted in the middle! ##########")
		printResults(startTime)
		os.Exit(0)
	}()

	go func() {
		for _ = range ticker.C {
			curIndex++
			// For the case load generator run on the same machine of k8s working node
			if showLocalCPU {
				cpuProfile()
			} else {
				// Do not display CPU info during run time
				if false { // aveCPU > 0 && curIndex > (timeInterval/10 - 1) {
					outputToStdout(fmt.Sprintf("CPU Usage: %d", aveCPU))
					curIndex = 0
				} else {
					outputToStdout("Time Passed: 10 seconds...")
				}
			}
		}
	}()

	// Set requests to web server every 30 secs to get CPU info
	go func() {
		time.Sleep(5 * time.Second)
		for _ = range tickerCPU.C {
			out, err := exec.Command("kubectl", "top", "nodes").Output()
			if err != nil || len(out) == 0 {
				continue
			}
			aveCPU = getAverageCPU(string(out))
		}
	}()

	// Parse title before getLogFileName
	getTitle()

	goMaxProcs := os.Getenv("GOMAXPROCS")

	if goMaxProcs == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	outputToStdout("Start auto loader......")

	if urlsFilePath == "" {
		command = fmt.Sprintf("autoloader -u %s -c %d -ci %d -ti %d\n", urlPath, clients, clientStep, timeInterval)
	} else {
		command = fmt.Sprintf("autoloader -f %s -c %d -ci %d -ti %d\n", urlsFilePath, clients, clientStep, timeInterval)
	}

	currentClient := clients
	var out []byte
	for true {
		if urlsFilePath == "" {
			gobenchCmd := fmt.Sprintf("./gobench -u %s -c %d -t %d -e %s\n", urlPath,
				currentClient, timeInterval, expResult)
			if DEBUG {
				outputToStdout(gobenchCmd)
			}
			out, err = exec.Command("./gobench", "-u", urlPath, "-c",
				fmt.Sprintf("%d", currentClient), "-t",
				fmt.Sprintf("%d", timeInterval), "-e",
				expResult).Output()
			if err != nil {
				log.Fatal(err.Error())
			}
		} else {
			gobenchCmd := fmt.Sprintf("./gobench -f %s -c %d -t %d\n", urlsFilePath, currentClient, timeInterval)
			if DEBUG {
				outputToStdout(gobenchCmd)
			}
			out, err = exec.Command("./gobench", "-f", urlsFilePath, "-c",
				fmt.Sprintf("%d", currentClient), "-t",
				fmt.Sprintf("%d", timeInterval), "-e", expResult).Output()
			if err != nil {
				log.Fatal(err.Error())
			}
		}

		// Get the average of local CPU usage, not used any more!
		if len(cpuUsage) > 0 {
			if showLocalCPU {
				for _, cpu := range cpuUsage {
					aveCPU += cpu
				}
				aveCPU = aveCPU / len(cpuUsage)
			}
			cpuUsage = nil
		}
		//////////

		elapsed, networkFailed := parseOutput(string(out), currentClient, aveCPU)
		// If response time longer than SLA, stop. Do not support multi case
		if sla != -1 && checkSLA(elapsed) {
			break
		}

		// If a lot of network failed happens, something really wrong, system could
		// run out of resources, so do not continue the tests
		// If lower than maxReq for more than maxRetry times, get out
		if networkFailed || retry >= maxRetry {
			break // get out of the loop
		}
		currentClient += clientStep
		// if user set last client number, stop the tests there
		if clientEnd != -1 && currentClient > clientEnd {
			break
		}
	}

	ticker.Stop()
	printResults(startTime)
}
