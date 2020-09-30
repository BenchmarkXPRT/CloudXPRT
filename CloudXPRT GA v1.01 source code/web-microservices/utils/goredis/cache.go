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

package goredis

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

//var redisClient *redis.Client
type MyRedis struct {
	Client *redis.Client
}

func InitializeRedis(addr string) *MyRedis {
	redisClient := redis.NewClient(&redis.Options{
		// Addr: "localhost:6379",
		//Addr:  "redis-service.default.svc.cluster.local:6379",
		Addr:       addr,
		PoolSize:   100,
		MaxRetries: 2,
		Password:   "",
		DB:         0,
	})

	ping, err := redisClient.Ping().Result()
	if err == nil && len(ping) > 0 {
		fmt.Println("Connected to Redis")
	} else {
		log.Fatal("Redis Connection Failed")
	}

	return &MyRedis{Client: redisClient}
}

func (redisClient MyRedis) GetValue(key string) (string, error) {
	//var deserializedValue interface{}
	serializedValue, err := redisClient.Client.Get(key).Result()
	//json.Unmarshal([]byte(serializedValue), &deserializedValue)
	return serializedValue, err
}

func (redisClient MyRedis) SetValue(key string, value interface{}) (bool, error) {
	serializedValue, _ := json.Marshal(value)
	err := redisClient.Client.Set(key, string(serializedValue), 0).Err()
	return true, err
}

func (redisClient MyRedis) SetRawValue(key string, value interface{}) (bool, error) {
	err := redisClient.Client.Set(key, value, 0).Err()
	return true, err
}

func (redisClient MyRedis) GetRawValue(key string) ([]byte, error) {
	rawValue, err := redisClient.Client.Get(key).Bytes()
	return rawValue, err
}

func (redisClient MyRedis) SetValueWithTTL(key string, value interface{}, ttl int) (bool, error) {
	serializedValue, _ := json.Marshal(value)
	err := redisClient.Client.Set(key, string(serializedValue), time.Duration(ttl)*time.Second).Err()
	return true, err
}

func (redisClient *MyRedis) RPush(key string, valueList []string) (bool, error) {
	err := redisClient.Client.RPush(key, valueList).Err()
	return true, err
}

func (redisClient *MyRedis) RpushWithTTL(key string, valueList []string, ttl int) (bool, error) {
	err := redisClient.Client.RPush(key, valueList, ttl).Err()
	return true, err
}

func (redisClient *MyRedis) LRange(key string) ([]string, error) {
	temp := redisClient.Client.LRange(key, 0, -1)
	err := temp.Err()
	return temp.Val(), err
}

func (redisClient *MyRedis) ListLength(key string) int64 {
	return redisClient.Client.LLen(key).Val()
}

func (redisClient *MyRedis) Publish(channel string, message string) {
	redisClient.Client.Publish(channel, message)
}

func (redisClient *MyRedis) GetKeyListByPattern(pattern string) []string {
	return redisClient.Client.Keys(pattern).Val()
}

func (redisClient *MyRedis) IncrementValue(key string) int64 {
	return redisClient.Client.Incr(key).Val()
}

func (redisClient *MyRedis) DelKey(key string) error {
	return redisClient.Client.Del(key).Err()
}

func (redisClient *MyRedis) Keys(pattern string) ([]string, error) {
	temp := redisClient.Client.Keys(pattern)
	err := temp.Err()
	return temp.Val(), err
}
