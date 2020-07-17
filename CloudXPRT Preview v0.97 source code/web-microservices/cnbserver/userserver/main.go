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
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/intel/goredis"
)

var session *gocql.Session
var client *goredis.MyRedis

type record struct {
	ID       string `json:"id"`       // in the format of user_123
	Password string `json:"password"` // in the format of pass_123
	Email    string `json:"email"`
}

func main() {
	// connect to redis server
	client = goredis.InitializeRedis("redis-user-service.default.svc.cluster.local:6379")

	// connect cassandra
	var err error
	// connect to the cluster
	cluster := gocql.NewCluster("cassandra-0.cassandra", "cassandra-1.cassandra", "cassandra-2.cassandra")
	cluster.Keyspace = "cnb"
	cluster.Consistency = gocql.Any
	session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer session.Close()

	fmt.Println("Connected to cassandra DB")

	r := mux.NewRouter()

	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	http.ListenAndServe(":8079", r)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON in the body and store it to DB
	d := json.NewDecoder(r.Body)
	p := &record{}
	err := d.Decode(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	result := findUserByID(p.ID)
	if len(result) > 0 {
		w.Write([]byte("User ID " + p.ID + " exists in DB already\n"))
		return
	}

	storeToDB(p)
	w.Write([]byte(p.ID + " record is created successfully\n"))
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var results []record

	// Query in cassandra first
	err := getUsersDB(&results)
	if err != nil {
		w.Write([]byte("Failed to find user data in Cassandra DB: " + err.Error() + "\n"))
		return
	}

	if len(results) == 0 {
		w.Write([]byte("No user records found in Cassandra DB\n"))
		return
	}

	keys, err := client.Keys("*")
	if err != nil {
		w.Write([]byte("Failed to find user records in redis: " + err.Error() + "\n"))
	} else if len(keys) <= len(results) {
		updateCache := false
		if len(keys) < len(results) {
			updateCache = true
		}
		for _, rec := range results {
			// Set retrieved values from Cassandra in cache
			if updateCache {
				client.SetValue(rec.ID, rec)
			}
			temp, _ := json.Marshal(rec)
			w.Write(temp)
			w.Write([]byte("\n"))
		}
	} else if len(keys) > len(results) {
		w.Write([]byte("Found mismatched records in redis and Cassandra DB\n"))
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if len(id) == 0 {
		w.Write([]byte("User ID can not be empty\n"))
		return
	}
	recvalue := findUserByID(id)
	w.Write([]byte(recvalue))
	w.Write([]byte("\n"))
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if len(id) == 0 {
		w.Write([]byte("User ID can not be empty\n"))
		return
	}

	err := deleteFromDB(id)
	if err != nil {
		w.Write([]byte("Failed to update user info in DB: " + err.Error()))
		w.Write([]byte("\n"))
		return
	}

	d := json.NewDecoder(r.Body)
	p := &record{}
	err = d.Decode(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	storeToDB(p)
	w.Write([]byte("User record is updated successfully\n"))

}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if len(id) == 0 {
		w.Write([]byte("User ID can not be empty\n"))
		return
	}
	err := deleteFromDB(id)
	if err == nil {
		w.Write([]byte("User " + id + " record is deleted successfully\n"))
	} else {
		w.Write([]byte("Failed to delete user from DB: " + err.Error()))
		w.Write([]byte("\n"))
	}
}

func findUserByID(id string) string {
	result, err := client.GetValue(id)
	// If not find in cache, still could be in Cassandra
	if err != nil {
		errCass, rec := findUserByIDDB(id)
		// also not found in Cassandra
		if errCass != nil {
			return ""
		}

		// Update cache if only found in Cassandra
		serializedValue, _ := json.Marshal(rec)
		client.SetValue(id, rec)
		return string(serializedValue)
	}
	return result
}

// below are all the DB related APIs
func getUsersDB(results *[]record) error {
	m := map[string]interface{}{}
	iter := session.Query(`SELECT * FROM user`).Consistency(gocql.One).Iter()

	for iter.MapScan(m) {
		*results = append(*results, record{
			ID:       m["id"].(string),
			Password: m["password"].(string),
			Email:    m["email"].(string),
		})
		m = map[string]interface{}{}
	}

	if err := iter.Close(); err != nil {
		return err
	}
	return nil
}

// Only one record should be found
func findUserByIDDB(id string) (error, record) {
	var rec = record{}
	err := session.Query(`SELECT id, password, email FROM user WHERE id = ? LIMIT 1`,
		id).Consistency(gocql.One).Scan(&rec.ID, &rec.Password, &rec.Email)
	if err != nil {
		return err, rec
	}
	return nil, rec
}

func deleteFromDB(id string) error {
	err := client.DelKey(id)
	if err != nil {
		return err
	}

	// delete from cassandra
	err = session.Query("DELETE FROM user WHERE id = ?", id).Exec()
	if err != nil {
		return err
	}

	return nil
}

func storeToDB(r *record) {
	// store to redis
	ok, err := client.SetValue(r.ID, r)
	if err != nil || !ok {
		log.Fatal(err)
	}

	// store to cassandra
	err = session.Query("INSERT INTO user(id, password, email) VALUES(?, ?, ?)",
		r.ID, r.Password, r.Email).Exec()
	if err != nil {
		log.Fatal(err)
	}
}
