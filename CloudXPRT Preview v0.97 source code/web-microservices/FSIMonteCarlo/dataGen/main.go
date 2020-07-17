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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Record struct {
	Stock  float64 `json:"stock"`
	Strike float64 `json:"strike"`
	Year   float64 `json:"year"`
}

const LEN = 4096

func main() {
	var mc [LEN]Record

	for j := 0; j < 10; j++ {
		rand.Seed(time.Now().UTC().UnixNano())
		for i := 0; i < LEN; i++ {
			v1 := rand.Float64()*45 + 5
			v2 := rand.Float64()*15 + 10
			v3 := rand.Float64()*4 + 1
			r := Record{Stock: v1, Strike: v2, Year: v3}
			mc[i] = r
		}

		output, _ := json.MarshalIndent(mc, "", "    ")
		err := ioutil.WriteFile(fmt.Sprintf("output%d.json", j), output, 0644)
		check(err)
	}
}
