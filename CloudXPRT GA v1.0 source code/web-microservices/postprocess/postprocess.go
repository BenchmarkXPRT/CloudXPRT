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
	"flag"
	"fmt"
	"html/template"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var (
	directory  string
	configfile string
	number     int
	title      string
	outputfile string
)

type Result struct {
	clients   string
	requests  []int
	responses []int
	reqMid    int
	respMid   int
	factor    float32
}

type ReportItem struct {
	Request   int
	RequestC  bool
	Response  int
	ResponseC bool
}

type Report struct {
	Time    string
	Clients string
	Items   []ReportItem
	Ratio   float32
}

var (
	Reports []Report

	// Predefine some colors for plot to use
	plotColors = []color.RGBA{
		color.RGBA{R: 255, A: 255},
		color.RGBA{G: 255, A: 255},
		color.RGBA{B: 255, A: 255},
		color.RGBA{R: 255, B: 128, A: 255},
		color.RGBA{G: 255, B: 128, A: 255},
		color.RGBA{R: 255, G: 128, A: 255},
		color.RGBA{B: 255, G: 128, A: 255},
		color.RGBA{G: 255, R: 128, A: 255},
		color.RGBA{B: 255, R: 128, A: 255}}

	// support up to 9 runs!
	Runtime = []string{"FIRST", "SECOND", "THIRD", "FOURTH",
		"FIFTH", "SIXTH", "SEVENTH", "EIGHTH", "NINTH"}
)

type ReportData struct {
	ReportTitle string
	Reports     []Report
	RunTimes    []string
}

type Plot struct {
	Legend string
	Data   string
}

type Config struct {
	Title  string
	Output string
	Plots  []Plot
}

const DEBUG = false

func init() {
	flag.StringVar(&directory, "d", "./output", "Directory contains files to be processed")
	flag.StringVar(&configfile, "p", "", "Plot configuration file in json format")
	flag.IntVar(&number, "n", 3, "Number of files to be processed")
	flag.StringVar(&title, "t", "mc", "Files with title to be processed (mc|ocr)")
	flag.StringVar(&outputfile, "o", "", "Post process output file name")
}

func main() {
	flag.Parse()

	if len(configfile) > 0 {
		// Plot assigned log files with titles, ignore other options
		plotLogFiles()
	} else {
		if number <= 1 {
			log.Fatalf("Too small number of files to do post process")
		}
		if even(number) {
			log.Fatalf("Need odd number of input files to do post process")
		}
		if number > 9 {
			log.Fatalf("Too big number of files to do post process (max = 9)")
		}
		processMutliFile()

		// if number == 3, 5, 7, 9, also provide web server for HTML format report
		if number <= 9 {
			tmpl := template.Must(template.ParseFiles("report.html"))

			runtimes := make([]string, number)
			for i := 0; i < number; i++ {
				runtimes[i] = Runtime[i]
			}
			http.Handle("/", http.FileServer(http.Dir("css/")))
			http.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
				data := ReportData{
					ReportTitle: "CNB " + strings.ToUpper(title) + " Test Results",
					Reports:     Reports,
					RunTimes:    runtimes,
				}
				tmpl.Execute(w, data)
			})

			fmt.Println("Please open browser and visit http://IP:8088/report for HTML format report.")
			log.Fatal(http.ListenAndServe(":8088", nil))
		}
	}
}

func even(num int) bool {
	return num%2 == 0
}

func readOneFile(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading input file, %s", err)
	}
	lines := strings.Split(string(content), "\n")
	return lines
}

func getReqResp(line string) (string, string, int, int) {
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
	return tokens[9], tokens[0], requests, response
}

// Plot one to more log files with title
func plotLogFiles() {
	// decode configuration file
	viper.SetConfigFile(configfile)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	conf := &Config{}
	err := viper.Unmarshal(conf)
	if err != nil {
		log.Fatalf("Unable to decode into config struct, %s", err.Error())
	}

	var plotData = make(map[string]plotter.XYs)

	for _, plot := range conf.Plots {
		lines := readOneFile(plot.Data)

		var reqTotal, respTotal []int
		var interval string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 || strings.HasPrefix(line, "CONCURRENCY") {
				continue
			}

			time, _, requests, response := getReqResp(line)
			// use first time as interval
			if len(interval) == 0 {
				interval = time
			}
			reqTotal = append(reqTotal, requests)
			respTotal = append(respTotal, response)
		}
		plotData[plot.Legend] = createPoints(reqTotal, respTotal, interval)
	}

	plotResults(plotData, conf.Title, conf.Output)
}

func createPoints(requests []int, responses []int, interval string) plotter.XYs {
	pts := make(plotter.XYs, len(requests))
	for i := range pts {
		pts[i].X = float64(responses[i])
		pts[i].Y = float64(requests[i])
	}
	return pts
}

func plotResults(plotData map[string]plotter.XYs, title string, outputfile string) {
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}

	p.Title.Text = title
	p.X.Label.Text = "95%ile Latency(ms)"
	p.X.Min = 0
	p.X.Max = 3000
	p.Y.Min = 0
	p.Y.Max = 1500
	p.Y.Label.Text = "Throughput(Successful Requests in 60 Sec)"

	// get all the keys of map and sort them
	keys := make([]string, 0)
	for key, _ := range plotData {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// change legend font and size to our desired values
	p.Legend.Font, _ = vg.MakeFont("Courier", 8)

	i := 0
	for _, key := range keys {
		lpLine, lpPoints, err := plotter.NewLinePoints(plotData[key])
		if err != nil {
			log.Fatal(err)
		}
		lpLine.Color = plotColors[i]
		p.Add(lpLine, lpPoints)
		p.Legend.Add(key, lpLine, lpPoints)
		i++
		if i >= len(plotColors) {
			i = 0
		}
	}
	// Save the plot to output PNG file.
	if err := p.Save(6*vg.Inch, 4*vg.Inch, outputfile); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Please check your plot in %s file\n", outputfile)
}

func findFileNamesByDate() []string {
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

func processMutliFile() {
	var filesizes []int
	var buf bytes.Buffer
	fileNames := findFileNamesByDate()

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

	fmt.Println("\nResults after geomean post process:")
	for idx, line := range allContents[0] {
		result := &Result{}
		report := Report{}
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "CONCURRENCY") {
			continue
		}

		time, clients, _, _ := getReqResp(line)
		result.clients = clients
		report.Clients = clients
		report.Time = time
		report.Items = make([]ReportItem, number)
		for i := 0; i < number; i++ {
			temp := strings.TrimSpace(allContents[i][idx])
			_, _, tempReq, tempResp := getReqResp(temp)
			result.requests = append(result.requests, tempReq)
			result.responses = append(result.responses, tempResp)
			report.Items[i].Request = tempReq
			report.Items[i].Response = tempResp
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
		var reqChosen = false
		var respChosen = false
		for i := 0; i < len(report.Items); i++ {
			item := &report.Items[i]
			if item.Request == result.reqMid && !reqChosen {
				item.RequestC = true
				reqChosen = true
			}
			if item.Response == result.respMid && !respChosen {
				item.ResponseC = true
				respChosen = true
			}
		}

		result.factor = float32(result.reqMid) / float32(result.respMid)
		report.Ratio = result.factor
		if DEBUG {
			fmt.Println(result)
		}
		Reports = append(Reports, report)

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
