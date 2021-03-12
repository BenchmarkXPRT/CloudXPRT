/*******************************************************************************
* The code in file 'gobench.go' is based on gobench.go, found at
* https://github.com/cmpxchg16/gobench, and licensed under New BSD License
* by Uri Shamay (shamayuri@gmail.com).
*
* Additional code and modification to gobench.go are Copyright 2020 BenchmarkXPRT
* Development Community, and licensed under Apache License, Version 2.0.
*
*
* Copyright 2020 Copyright 2020 BenchmarkXPRT Development Community
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
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	requests         int64
	period           int64
	clients          int
	url              string
	urlsFilePath     string
	keepAlive        bool
	postDataFilePath string
	writeTimeout     int
	readTimeout      int
	authHeader       string
	cookieHeader     string
	expResult        string
)

type Configuration struct {
	urls         []string
	method       string
	postData     []byte
	requests     int64
	period       int64
	keepAlive    bool
	authHeader   string
	cookieHeader string

	myClient fasthttp.Client
}

type Result struct {
	requests      int64
	success       int64
	networkFailed int64
	badFailed     int64
	mismatched    int64
}

var readThroughput int64
var writeThroughput int64

var DEBUG = false
var timeThreshold = "-1"

var respTimeMap = make(map[string][]int)
var timeThresholdMap = make(map[string]string)

var m sync.RWMutex

type MyConn struct {
	net.Conn
}

func (mc *MyConn) Read(b []byte) (n int, err error) {
	len, err := mc.Conn.Read(b)

	if err == nil {
		atomic.AddInt64(&readThroughput, int64(len))
	}

	return len, err
}

func (mc *MyConn) Write(b []byte) (n int, err error) {
	len, err := mc.Conn.Write(b)

	if err == nil {
		atomic.AddInt64(&writeThroughput, int64(len))
	}

	return len, err
}

func init() {
	flag.Int64Var(&requests, "r", -1, "Number of requests per client")
	flag.IntVar(&clients, "c", 100, "Number of concurrent clients")
	flag.StringVar(&url, "u", "", "URL")
	flag.StringVar(&urlsFilePath, "f", "", "URL's file path (line seperated)")
	flag.BoolVar(&keepAlive, "k", true, "Do HTTP keep-alive")
	flag.StringVar(&postDataFilePath, "d", "", "HTTP POST data file path")
	flag.Int64Var(&period, "t", -1, "Period of time (in seconds)")
	flag.IntVar(&writeTimeout, "tw", 5000, "Write timeout (in milliseconds)")
	flag.IntVar(&readTimeout, "tr", 5000, "Read timeout (in milliseconds)")
	flag.StringVar(&authHeader, "auth", "", "Authorization header")
	flag.StringVar(&cookieHeader, "cookie", "", "Cookie header")
	flag.StringVar(&expResult, "e", "", "Expected string pattern from response")
	// flag.StringVar(&timeThreshold, "tt", "-1", "Time threshold for Apdex score (in milliseconds)")
}

func printResults(results map[int]*Result, startTime time.Time) {
	var requests int64
	var success int64
	var networkFailed int64
	var badFailed int64
	var mismatched int64

	for _, result := range results {
		requests += result.requests
		success += result.success
		networkFailed += result.networkFailed
		badFailed += result.badFailed
		mismatched += result.mismatched
	}

	elapsed := int64(time.Since(startTime).Seconds())

	if elapsed == 0 {
		elapsed = 1
	}

	fmt.Println()
	fmt.Printf("Requests:                       %10d hits\n", requests)
	fmt.Printf("Successful requests:            %10d hits\n", success)
	fmt.Printf("Network failed:                 %10d hits\n", networkFailed)
	fmt.Printf("Bad requests failed (!2xx):     %10d hits\n", badFailed)
	fmt.Printf("Pattern mismatch:               %10d hits\n", mismatched)
	fmt.Printf("Successful requests rate:       %10.2f hits/sec\n", (float32(success))/(float32(elapsed)))
	fmt.Printf("Read throughput:                %10d bytes/sec\n", readThroughput/elapsed)
	fmt.Printf("Write throughput:               %10d bytes/sec\n", writeThroughput/elapsed)
	fmt.Printf("Test time:                      %10d sec\n", elapsed)

	// Time elapsed results
	fmt.Println("\nPercentage of the requests served within a certain time (ms)")
	// Need sort respTimeMap by its keys
	keys := make([]string, len(respTimeMap))
	var i = 0
	for key := range respTimeMap {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := respTimeMap[key]
		size := len(value)
		sort.Ints(value)
		fmt.Println("For URL:", key, "\nTotal", size, "responses are received")
		fmt.Printf(" 50%% %10d\n", value[(int)(size/2)])
		fmt.Printf(" 60%% %10d\n", value[(int)(size*6/10)])
		fmt.Printf(" 70%% %10d\n", value[(int)(size*7/10)])
		fmt.Printf(" 80%% %10d\n", value[(int)(size*8/10)])
		fmt.Printf(" 90%% %10d\n", value[(int)(size*9/10)])
		fmt.Printf(" 95%% %10d\n", value[(int)(size*95/100)])
		fmt.Printf("100%% %10d\n", value[(int)(size-1)])

		// Mainly for file contains multi URLs case
		if val, ok := timeThresholdMap[key]; ok {
			timeThreshold = val
		}

		// If time threshold is set as lower:upper format, calculate the Apdex score here
		if strings.Contains(timeThreshold, ":") {

			tlower, tupper := handleTimeThreshold(timeThreshold)
			var satisfied, tolerated, frustrated = 0, 0, 0
			for _, v := range value {
				if v <= tlower {
					satisfied++
				} else if v > tlower && v <= tupper {
					tolerated++
				} else {
					frustrated++
				}
			}
			fmt.Printf("\nFor time threshold values [%d:%dms]\n", tlower, tupper)
			fmt.Printf("Satisfied requests count:  %10d\n", satisfied)
			fmt.Printf("Tolerated requests count:  %10d\n", tolerated)
			fmt.Printf("Frustrated requests count: %10d\n", frustrated)
			fmt.Printf("Apdex score is: %.5f\n\n",
				(float32(satisfied)+float32(tolerated)/2.0)/(float32(size)))
		}
	}
}

// In the format of lower:upper
func handleTimeThreshold(input string) (int, int) {
	tokens := strings.Split(input, ":")
	tlower, err := strconv.Atoi(tokens[0])
	if err != nil {
		log.Fatalf("Invalid satisfied time threshold %s", tokens[0])
	}
	tupper, err := strconv.Atoi(tokens[1])
	if err != nil {
		log.Fatalf("Invalid tolerated time threshold %s", tokens[1])
	}
	if tlower >= tupper {
		log.Fatalf("Satisfied time threshold should be smaller than tolerated [%d:%d]",
			tlower, tupper)
	}
	return tlower, tupper
}

func shuffle(urls []string, size int) {
	for i := 0; i < size; i++ {
		temp := rand.Intn(size)
		urls[i], urls[temp] = urls[temp], urls[i]
	}
}

func readLines(path string) (lines []string, err error) {

	var file *os.File
	var part []byte
	var prefix bool
	var currWeight = 0

	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			temp := strings.TrimSpace(buffer.String())
			if strings.HasPrefix(temp, "WEIGHT:") {
				// Format of this line is: WEIGHT:3
				tempWeight := strings.TrimPrefix(temp, "WEIGHT:")
				currWeight, err = strconv.Atoi(tempWeight)
				if err != nil {
					return
				}
				buffer.Reset()
			} else {
				if currWeight > 0 {
					// Get time threshold value and trim this part
					if strings.Contains(temp, "[THOLD]") {
						results := strings.Split(temp, "[THOLD]")
						idx := strings.Index(temp, "[")
						timeThresholdMap[temp[0:idx]] = results[1]
						temp = results[0]
					}
					if DEBUG {
						fmt.Println(temp, " is attached ", currWeight)
					}
					for i := 0; i < currWeight; i++ {
						lines = append(lines, temp)
					}
				}
				buffer.Reset()
			}
		}
	}
	if err == io.EOF {
		err = nil
	}

	shuffle(lines, len(lines))
	return
}

func NewConfiguration() *Configuration {

	if urlsFilePath == "" && url == "" {
		flag.Usage()
		os.Exit(1)
	}

	if requests == -1 && period == -1 {
		fmt.Println("Requests or period must be provided")
		flag.Usage()
		os.Exit(1)
	}

	if requests != -1 && period != -1 {
		fmt.Println("Only one should be provided: [requests|period]")
		flag.Usage()
		os.Exit(1)
	}

	configuration := &Configuration{
		urls:         make([]string, 0),
		method:       "GET",
		postData:     nil,
		keepAlive:    keepAlive,
		requests:     int64((1 << 63) - 1),
		authHeader:   authHeader,
		cookieHeader: cookieHeader,
		myClient:     fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}}}

	if period != -1 {
		configuration.period = period

		timeout := make(chan bool, 1)
		go func() {
			<-time.After(time.Duration(period) * time.Second)
			timeout <- true
		}()

		go func() {
			<-timeout
			pid := os.Getpid()
			proc, _ := os.FindProcess(pid)
			err := proc.Signal(os.Interrupt)
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}

	if requests != -1 {
		configuration.requests = requests
	}

	if urlsFilePath != "" {
		fileLines, err := readLines(urlsFilePath)

		if err != nil {
			log.Fatalf("Error in ioutil.ReadFile for file: %s Error: %s", urlsFilePath, err.Error())
		}

		configuration.urls = fileLines
		if DEBUG {
			fmt.Println(configuration.urls)
		}
	}

	if len(url) > 10 {
		configuration.urls = append(configuration.urls, url)
	}

	if postDataFilePath != "" {
		configuration.method = "POST"

		data, err := ioutil.ReadFile(postDataFilePath)

		if err != nil {
			log.Fatalf("Error in ioutil.ReadFile for file path: %s Error: %s", postDataFilePath, err.Error())
		}

		configuration.postData = data
	}

	configuration.myClient.ReadTimeout = time.Duration(readTimeout) * time.Millisecond
	configuration.myClient.WriteTimeout = time.Duration(writeTimeout) * time.Millisecond
	configuration.myClient.MaxConnsPerHost = clients

	configuration.myClient.Dial = MyDialer()

	return configuration
}

func MyDialer() func(address string) (conn net.Conn, err error) {
	return func(address string) (net.Conn, error) {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return nil, err
		}

		myConn := &MyConn{Conn: conn}

		return myConn, nil
	}
}

func updateElapsed(url string, elapsed int) {
	m.Lock()
	respTimeMap[url] = append(respTimeMap[url], elapsed)
	m.Unlock()
}

func client(configuration *Configuration, result *Result, done *sync.WaitGroup) {
	var errSet = make(map[string]struct{})
	for result.requests < configuration.requests {
		for _, tmpURL := range configuration.urls {
			// not a valid URL, ignore it
			if len(tmpURL) < 10 {
				continue
			}
			// expected contents from response
			pattern := make([]byte, 0, 256)

			req := fasthttp.AcquireRequest()

			if strings.Contains(tmpURL, "[EXPECT]") {
				result := strings.Split(tmpURL, "[EXPECT]")
				pattern = []byte(result[1])
				tmpURL = result[0]
			} else if len(expResult) > 0 {
				pattern = []byte(expResult)
			}

			if strings.Contains(tmpURL, "[POST]") {
				result := strings.Split(tmpURL, "[POST]")
				req.SetRequestURI(result[0])
				req.Header.SetMethodBytes([]byte("POST"))
				req.SetBody([]byte(result[1]))
			} else {
				req.SetRequestURI(tmpURL)
				req.Header.SetMethodBytes([]byte("GET"))
				req.SetBody(configuration.postData)
			}

			if configuration.keepAlive == true {
				req.Header.Set("Connection", "keep-alive")
			} else {
				req.Header.Set("Connection", "close")
			}

			// Add set cookie, for example usrId=6
			if len(configuration.cookieHeader) > 0 {
				temp := strings.Split(configuration.cookieHeader, "=")
				req.Header.SetCookie(temp[0], temp[1])
			}

			if len(configuration.authHeader) > 0 {
				req.Header.Set("Authorization", configuration.authHeader)
			}

			start := time.Now()
			resp := fasthttp.AcquireResponse()
			err := configuration.myClient.Do(req, resp)

			// Time to get response in ms
			elapsed := (int)(time.Since(start) / 1000000)
			updateElapsed(req.URI().String(), elapsed)

			// total request number always increase by one here
			result.requests++

			if err != nil {
				if _, ok := errSet[err.Error()]; !ok {
					errSet[err.Error()] = struct{}{}
					// debug only
					fmt.Println(err.Error())
				}
				result.networkFailed++
				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)
				continue
			}

			statusCode := resp.StatusCode()

			if statusCode == fasthttp.StatusOK {
				if len(pattern) > 0 {
					if bytes.Contains(resp.Body(), pattern) {
						result.success++
					} else {
						if DEBUG {
							fmt.Println("not match: ", pattern)
						}
						result.mismatched++
					}
				} else {
					result.success++
				}
			} else {
				result.badFailed++
				if DEBUG {
					fmt.Println(statusCode)
				}
			}

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
		}
	}

	done.Done()
}

func main() {

	startTime := time.Now()
	var done sync.WaitGroup
	results := make(map[int]*Result)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		_ = <-signalChannel
		printResults(results, startTime)
		os.Exit(0)
	}()

	flag.Parse()

	configuration := NewConfiguration()

	goMaxProcs := os.Getenv("GOMAXPROCS")

	if goMaxProcs == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	fmt.Printf("Dispatching %d clients\n", clients)

	done.Add(clients)
	for i := 0; i < clients; i++ {
		result := &Result{}
		results[i] = result
		go client(configuration, result, &done)

	}
	fmt.Println("Waiting for results...")
	done.Wait()
	printResults(results, startTime)
}
