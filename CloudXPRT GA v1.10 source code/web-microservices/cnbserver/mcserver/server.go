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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cloudxprt/cnbserver/goredis"
	"github.com/valyala/fasthttp"
)

var client *goredis.MyRedis
var cpuArch = 0
var numThreads = os.Getenv("OMP_NUM_THREADS")
var divisor = os.Getenv("MC_WL_DIVISOR")

func main() {
	testcpu, err := exec.Command("lscpu").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(string(testcpu), "avx512") {
		cpuArch = 2
	} else if strings.Contains(string(testcpu), "avx2") {
		cpuArch = 1
	} else if strings.Contains(string(testcpu), "sse4") {
		// Add support for atom CPU
		cpuArch = 3
	}

	// connect to redis server
	client = goredis.InitializeRedis("redis-crypt-service.default.svc.cluster.local:6379")

	s := &fasthttp.Server{
		Handler:     handler,
		Concurrency: fasthttp.DefaultConcurrency,
	}

	if err := s.ListenAndServe(":8074"); err != nil {
		log.Fatalf("Error in ListenAndServe monte carlo server: %s", err)
	}
}

/*
func getLoad() string {
	if len(divisor) == 0 || divisor == "1" {
		return "4096"
	}

	switch divisor {
	case "2":
		return "2048"
	case "4":
		return "1024"
	case "8":
		return "512"
	case "16":
		return "256"
	case "32":
		return "128"
	default:
		log.Println("Not a valid divisor value")
	}
	//default value is 4096
	return "4096"
}
*/

func getLoad() string {
	// Only support when numThreads is 1, 2 or 4
	switch numThreads {
	case "4":
		return "4096"
	case "2":
		return "2048"
	case "1":
		return "1024"
	default:
		log.Println("Not a valid num of threads value")
	}
	//default value is 4096
	return "4096"
}

func handler(ctx *fasthttp.RequestCtx) {
	var out []byte
	var err error

	name := ctx.QueryArgs().Peek("name")
	if len(name) == 0 {
		log.Fatal("User name is missing")
	}

	contents, err := client.GetRawValue(string(name))
	if err != nil || len(contents) == 0 {
		log.Fatal("Could not find user profile to process")
	}

	tmpfile, err := ioutil.TempFile(".", "mcinput")
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(tmpfile.Name(), contents, 644)
	// clean up
	defer os.Remove(tmpfile.Name())

	tmp, err := strconv.Atoi(numThreads)
	if len(numThreads) == 0 || err != nil || tmp < 1 {
		log.Println(numThreads, " is not a valid value, use 4 instead")
		numThreads = "4"
	}

	workload := getLoad()

	// if support avx512
	if cpuArch == 2 {
		out, err = exec.Command("./MonteCarloInsideBlockingDP.avx512", numThreads, workload, "262144", "4k", "0",
			tmpfile.Name()).CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
	} else if cpuArch == 1 {
		out, err = exec.Command("./MonteCarloInsideBlockingDP.arch_avx2", numThreads, workload, "262144", "4k", "0",
			tmpfile.Name()).CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
	} else if cpuArch == 3 {
		out, err = exec.Command("./MonteCarloInsideBlockingDP.atom", numThreads, workload, "262144", "4k", "0",
			tmpfile.Name()).CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		out = []byte("CNB can only be run on CPUs with SSE4, AVX2, or AVX512 enabled\n")
	}
	fmt.Fprint(ctx, string(out))

	// return POD info
	podName, _ := exec.Command("printenv", "MY_POD_NAME").CombinedOutput()
	fmt.Fprint(ctx, "\nFrom POD: "+string(podName))
}
