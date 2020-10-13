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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Result struct {
	clients   string
	requests  []int
	responses []int
	reqMid    int
	respMid   int
	factor    float32
}

const (
	maxRetry   = 5
	configFile = "config.json"
	CPUSTR     = "000m"

	roundComplete = `
#########################################
%s workload iteration %d completed
#########################################
`
	allDone = `
***CNB completed. Please check results under output directory***
`
)

func main() {
	// or use SetConfigType("json") and SetConfigName("config")
	viper.SetConfigFile(configFile)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	option := strings.ToLower(viper.GetString("runoption"))
	if option == "all" {
		runAll()
	} else {
		runByTitle(option)
	}
}

func even(num int) bool {
	return num%2 == 0
}

func runPostProcess(title string) {
	if !viper.GetBool("postprocess") {
		return
	}

	runtime := viper.GetInt("iterations")
	if runtime > 1 && runtime <= 9 && !even(runtime) {
		time.Sleep(2 * time.Second)
		postprocess(runtime, title)
	} else {
		fmt.Println("Only when runtime is an odd number between 3 and 9 could run postprocess!")
	}
}

func runWorkload(title string) {
	titleLow := strings.ToLower(title)
	titleUpp := strings.ToUpper(title)

	var args []string
	args = []string{viper.GetString("workload.version"),
		viper.GetString("workload.cpurequests") + CPUSTR,
		viper.GetString("workload.cpurequests"),
		viper.GetString("autoloader.initialclients"),
		viper.GetString("autoloader.clientstep"),
		viper.GetString("autoloader.lastclients"),
		viper.GetString("autoloader.SLA"),
		viper.GetString("autoloader.timeinterval")}

	if viper.GetBool("hpamode") {
		args = append(args, "enablehpa")
	} else {
		args = append(args, "disablehpa")
	}

	var cmd *exec.Cmd
	var outputBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &outputBuffer)
	defer writeToLog(titleLow, &outputBuffer)

	runtime := viper.GetInt("iterations")
	for i := 0; i < runtime; i++ {
		if i == runtime-1 {
			// Only the last run needs clean up kubernetes resources
			args = append(args, "needclean")
			cmd = exec.Command("./"+titleLow+".sh", args...)
		} else {
			cmd = exec.Command("./"+titleLow+".sh", args...)
		}
		cmd.Stdout = mw
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Run "+titleUpp+" scripts failed with %s\n", err)
		}

		outputToBoth(&outputBuffer, fmt.Sprintf(roundComplete, titleUpp, i+1))
		if i == runtime-1 {
			outputToBoth(&outputBuffer, allDone)
		}
	}
}

func writeToLog(title string, outputBuffer *bytes.Buffer) {
	logFile, confFile := getLogConfigfileName(title)
	err := ioutil.WriteFile(logFile, outputBuffer.Bytes(), 0666)
	if err != nil {
		log.Fatalf("Error saving output to log file: %s\n", err.Error())
	}

	// copy config.json file to output directory
	input, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error read config file: %s\n", err.Error())
	}

	err = ioutil.WriteFile(confFile, input, 0666)
	if err != nil {
		log.Fatalf("Error saving config file to output directory: %s\n", err.Error())
	}
}

func outputToBoth(mybuffer *bytes.Buffer, input string) {
	fmt.Println(input)
	mybuffer.WriteString(input + "\n")
}

func getLogConfigfileName(title string) (string, string) {
	timeNow := time.Now().Format("20060102150405")
	exePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	index := 8
	newTime := timeNow[:index] + "_" + timeNow[index:]
	return exePath + "/output/autoloader_" + title + "_all_" + newTime + ".log",
		exePath + "/output/config_" + title + "_" + newTime + ".json"
}

func runByTitle(title string) {
	runWorkload(title)
	runPostProcess(title)
}

// For release0.5, only ocr workload is enabled
func runAll() {
	runWorkload("ocr")
	//runWorkload("kmeans")
	runPostProcess("ocr")
	//runPostProcess("kmeans")
}

func readOneFile(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading input file, %s", err.Error())
	}
	lines := strings.Split(string(content), "\n")
	return lines
}

func getReqResp(line string) (string, int, int) {
	tokens := strings.Fields(line)

	requests, err := strconv.Atoi(tokens[2])
	if err != nil {
		log.Fatalf("Invalid data format found %s, error: %s", tokens[2], err.Error())
	}

	response, err := strconv.Atoi(tokens[10])
	if err != nil {
		log.Fatalf("Invalid data format found %s, error: %s", tokens[10], err.Error())
	}
	if response == 0 {
		log.Fatalf("Invalid response time, should not be 0")
	}
	return tokens[0], requests, response
}

func findFileNamesByDate(number int, title string, directory string) []string {
	var results []string
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	// Sort files by modification time
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	for _, file := range files {
		if (!file.Mode().IsRegular()) || file.Mode().IsDir() ||
			strings.Contains(file.Name(), ".json") || strings.Contains(file.Name(), ".csv") {
			continue
		}
		if strings.Contains(file.Name(), "_"+title+"_") && (!strings.Contains(file.Name(), "_all_")) {
			results = append(results, directory+"/"+file.Name())
		}
		if len(results) == number {
			break
		}
	}
	if len(results) != number {
		fmt.Println(results)
		log.Fatal("More or less log files are found to do post process")
	}
	return results
}

func fileSize(path string) int {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return int(fi.Size())
}

func postprocess(number int, title string) {
	directory := "./output"
	var filesizes []int
	var buf bytes.Buffer

	fileNames := findFileNamesByDate(number, title, directory)

	// make sure files to be processed are around the same size
	for _, fname := range fileNames {
		filesizes = append(filesizes, fileSize(fname))
	}

	sort.Ints(filesizes)

	if filesizes[len(filesizes)-1]-filesizes[0] > 50 {
		fmt.Println(filesizes)
		log.Fatal("Files to be processed are not in the same size!")
	}

	var allContents [][]string
	for _, fname := range fileNames {
		fmt.Printf("File [%s] to be processed\n", fname)
		lines := readOneFile(fname)
		allContents = append(allContents, lines)
	}

	// use first non empty line as header
	idx := 0
	for len(strings.TrimSpace(allContents[0][idx])) == 0 {
		idx++
	}
	buf.WriteString(allContents[0][idx] + "\n")

	fmt.Println("\nResults after post process:")
	for idx, line := range allContents[0] {
		result := &Result{}
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "CONCURRENCY") || strings.HasPrefix(line, "Best") {
			continue
		}

		for i := 0; i < number; i++ {
			temp := strings.TrimSpace(allContents[i][idx])
			clients, tempReq, tempResp := getReqResp(temp)
			result.clients = clients
			result.requests = append(result.requests, tempReq)
			result.responses = append(result.responses, tempResp)
		}

		// back up requests and responses first
		reqBack := make([]int, len(result.requests))
		copy(reqBack, result.requests)
		respBack := make([]int, len(result.responses))
		copy(respBack, result.responses)

		sort.Ints(result.requests)
		sort.Ints(result.responses)
		result.reqMid = result.requests[(number-1)/2]
		result.respMid = result.responses[(number-1)/2]
		result.factor = float32(result.reqMid) / float32(result.respMid)
		// fmt.Println(result)

		// Generate geomean output line
		var lineTemp, lineTemp1, lineFinal string
		reqIdx := 0
		for i, req := range reqBack {
			if req == result.reqMid {
				lineTemp = allContents[i][idx]
				reqIdx = i
				break
			}
		}

		idx := strings.LastIndex(lineTemp, "(")
		if idx > 0 {
			lineTemp1 = lineTemp[:idx-1]
		} else {
			lineTemp1 = lineTemp
		}
		for _, resp := range respBack {
			if resp == result.respMid {
				// Only replace first occurance from back
				lineFinal = Rev(strings.Replace(Rev(lineTemp1), Rev(strconv.Itoa(respBack[reqIdx])),
					Rev(strconv.Itoa(resp)), 1))
			}
		}
		buf.WriteString(lineFinal + "\n")
	}
	fmt.Println("=====================================================================\n")
	fmt.Println(buf.String())
	fmt.Println("=====================================================================\n")

	outputfile := viper.GetString("ppoutputfile")
	if len(outputfile) > 0 {
		err := ioutil.WriteFile(outputfile, buf.Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("New result file %s after geomean is generated successfully!\n\n", outputfile)
		}
	}
}

func Rev(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
