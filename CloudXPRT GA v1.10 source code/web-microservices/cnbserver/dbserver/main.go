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
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gocql/gocql"
	"github.com/cloudxprt/cnbserver/goredis"
)

var session *gocql.Session
var client *goredis.MyRedis

const enableCassandra = true

// The `json:"whatever"` bit is a way to tell the JSON
// encoder and decoder to use those names instead of the
// capitalised names
type mcrecord struct {
	Name    string     `json:"name"`
	Date    string     `json:"date"`
	Elapsed float64    `json:"elapsed"`
	Ops     float64    `json:"ops"`
	Results []mcresult `json:"results"`
}

type mcresult struct {
	StockPrice     float64 `json:"stockprice"`
	StrikePrice    float64 `json:"strikeprice"`
	OptionYear     float64 `json:"optionyear"`
	CallResult     float64 `json:"callresult"`
	CallConfidence float64 `json:"callconfidence"`
}

func getMCRecords(key string) ([]mcrecord, error) {
	var results []mcrecord
	var result string
	iter := session.Query(`SELECT JSON * FROM montecarlo WHERE name = ?`, key).Consistency(gocql.One).Iter()

	for iter.Scan(&result) {
		var rec mcrecord
		err := json.Unmarshal([]byte(result), &rec)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}

	err := iter.Close()
	if err != nil {
		return nil, err
	}
	return results, nil
}

func mcHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// GET method is for debug only!
	case "GET":
		// Just send out the JSON version of results
		var key string
		keys, ok := r.URL.Query()["name"]

		if !ok || len(keys[0]) < 1 {
			key = generateKey()
		} else {
			key = keys[0]
		}

		results, err := client.LRange(key)
		countRedis := len(results)
		// If we only have memory cache enabled, no DB storage
		if !enableCassandra {
			if err == nil && countRedis > 0 {
				w.Write([]byte(strings.Join(results, "\n")))
				w.Write([]byte("\n"))
			} else {
				w.Write([]byte("no results found for " + key + "\n"))
			}
		}

		if enableCassandra {
			var count = 0
			if err := session.Query(`SELECT COUNT(*) FROM montecarlo WHERE name = ?`, key).Consistency(gocql.One).Scan(&count); err != nil {
				w.Write([]byte("Failed to find data count in Cassandra DB: " + err.Error() + "\n"))
				return
			}

			if count == 0 {
				w.Write([]byte("\nNo results found for " + key + "\n"))
				return
			}

			if count < countRedis {
				// Usually this case should not happen since DB has more robust data storage than memory cache
				w.Write([]byte(fmt.Sprintf("\nWarning! Cassandra found %d records while redis found %d records \n",
					count, countRedis)))
			} else if count == countRedis {
				// Redis has same records as DB, use redis results
				w.Write([]byte(strings.Join(results, "\n")))
				w.Write([]byte("\n"))
			} else {
				// Redis has less records than DB, should update redis value according to DB and write response here
				// Read from DB now
				dbResults, err := getMCRecords(key)
				if err != nil {
					w.Write([]byte("Failed to find data in Cassandra DB: " + err.Error() + "\n"))
					return
				}
				// Best efforts, try to delete key from redis first
				client.DelKey(key)

				for _, r := range dbResults {
					svalue, err := json.Marshal(r)
					if err != nil {
						log.Fatal(err)
					}
					_, err = client.RPush(r.Name, []string{string(svalue)})
					if err != nil {
						log.Fatal(err)
					}
					w.Write(svalue)
					w.Write([]byte("\n"))
				}
			}
		}

	case "POST":
		// Decode the JSON in the body and store it to DB
		d := json.NewDecoder(r.Body)
		p := &mcrecord{}
		err := d.Decode(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		storeMCToDB(p)
		w.Write([]byte("Monte Carlo results are stored successfully\n"))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not supported.")
	}
}

// store monte carlo results to redis and cassandra
func storeMCToDB(r *mcrecord) {
	svalue, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}

	// store to redis
	_, err = client.RPush(r.Name, []string{string(svalue)})
	if err != nil {
		log.Fatal(err)
	}

	if enableCassandra {
		// store to cassandra
		err = session.Query("INSERT INTO montecarlo JSON ?", svalue).Exec()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	// connect to redis server
	client = goredis.InitializeRedis("redis-service.default.svc.cluster.local:6379")

	if enableCassandra {
		// connect to cassandra
		var err error
		// kubernetes statefulset cassandra, 3 replica sets
		cluster := gocql.NewCluster("cassandra-0.cassandra",
			"cassandra-1.cassandra",
			"cassandra-2.cassandra")
		cluster.Keyspace = "cnb"
		cluster.Consistency = gocql.Any
		session, err = cluster.CreateSession()
		if err != nil {
			panic(err)
		}
		defer session.Close()

		fmt.Println("Connected to cassandra DB")
	}

	http.HandleFunc("/mc", mcHandler)

	http.ListenAndServe(":8078", nil)
}

func generateKey() string {
	prefix := "user_"
	id := rand.Intn(1000)

	return fmt.Sprintf("%s%d", prefix, id)
}
