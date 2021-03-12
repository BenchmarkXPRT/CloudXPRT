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
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/Shopify/sarama.v2"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/asticode/go-astikit"
	"github.com/leesper/go_rng"
	"github.com/montanaflynn/stats"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
)

//globals for timing and logging
var (
	numberOfPods     string = ""
	numberOfCpus     string = ""
	numKAFKAmessages string = ""
	dataset          string = ""
	numberOfMessages int    = 3
	minio_service_ip        = ""
	lambda_param            = ""

	testinProgress   bool    = false
	resultJson       string  = ""
	resultBuf                = bytes.Buffer{}
	podsStartTime            = time.Now()
	podsEndTime              = time.Now()
	processStartTime         = time.Now()
	max_timeProcess  float64 = 0.0
	resultsFolder            = "" //this is also the test name. all results are kept in this folder
)

//xgb_ const
const kafkabroker string = "broker:9092"
const transcodeTopic = "ktopic-xgboost"

type record struct {
	Name   string  `json:"name"`
	Addr   string  `json:"addr"`
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}
type transcodeResult struct {
	Key      string
	Duration string
	PodID    string
}

var (
	cnbaddr            = flag.String("addr", ":7079", "TCP address to listen to")
	cnbaddrTLS         = flag.String("addrTLS", ":8443", "TCP address to listen to TLS (aka SSL or HTTPS) requests. Leave empty for disabling TLS")
	byteRange          = flag.Bool("byteRange", false, "Enables byte range requests if set to true")
	certFile           = flag.String("certFile", "./cnbserver.crt", "Path to TLS certificate file")
	compress           = flag.Bool("compress", false, "Enables transparent response compression if set to true")
	dir                = flag.String("dir", "/usr/share/fileserver", "Directory to serve static files from")
	generateIndexPages = flag.Bool("generateIndexPages", true, "Whether to generate directory index pages")
	keyFile            = flag.String("keyFile", "./cnbserver.key", "Path to TLS key file")
	vhost              = flag.Bool("vhost", false, "Enables virtual hosting by prepending the requested path with the requested hostname")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Parse command-line flags.
	flag.Parse()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/stats":
			expvarhandler.ExpvarHandler(ctx)
		case "/xgb":
			xgb_Handler(ctx)
		case "/xgb_StartTest":
			xgb_StartTestHandler(ctx)
		case "/xgb_Status":
			xgb_StatusHandler(ctx)
		case "/xgb_DeleteTopics":
			xgb_DeleteTopicHandler(ctx)
		}
	}

	s := &fasthttp.Server{
		Handler:     requestHandler,
		Concurrency: fasthttp.DefaultConcurrency,
	}

	// Start HTTP server.
	if len(*cnbaddr) > 0 {
		log.Printf("Starting CNB HTTP server on %q", *cnbaddr)
		go func() {
			if err := s.ListenAndServe(*cnbaddr); err != nil {
				log.Fatalf("error in CNB server ListenAndServe: %s", err)
			}
		}()
	}

	// Start HTTPS server.
	if len(*cnbaddrTLS) > 0 {
		log.Printf("Starting CNB HTTPS server on %q", *cnbaddrTLS)
		go func() {
			if err := fasthttp.ListenAndServeTLS(*cnbaddrTLS, *certFile, *keyFile, requestHandler); err != nil {
				log.Fatalf("error in CNB server ListenAndServeTLS: %s", err)
			}
		}()
	}

	// Wait and serve.
	select {}
}

func xgb_Handler(ctx *fasthttp.RequestCtx) {
	host, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
		return
	}
	fmt.Fprintf(ctx, "  Hello from %v - xgb_Handler\n", host)

	resultsFolder = string(ctx.QueryArgs().Peek("resfolder"))
	fmt.Fprintf(ctx, "  Results folder: %v\n\n", resultsFolder)
	var folderName string = "./output/" + resultsFolder + "/"
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		os.Mkdir(folderName, os.FileMode(0577))
	}

	//to clean up
	var kafkaServer, kafkaTopic string
	kafkaServer = "broker:9092"
	kafkaTopic = "ktopic-xgboost"
	fmt.Fprintf(ctx, "  Connecting to kafka broker at %v!\n", kafkaServer+":"+kafkaTopic)
	//usingClientGo()
}

func xgb_StatusHandler(ctx *fasthttp.RequestCtx) {
	if testinProgress == true {
		if len(resultJson) > 0 {
			fmt.Fprintf(ctx, "Test in Progress...\n")
			fmt.Fprintf(ctx, "Returned: %s\n", resultJson)
		} else {
			fmt.Fprintf(ctx, "Test in Progress...\n")
		}
	} else {
		fmt.Fprintf(ctx, "Test is Done:\n")
		fmt.Fprintf(ctx, "%s\n", resultBuf.String())
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}
func xgb_CreateTopic() bool {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	brokers := []string{"broker:9092"}
	admin, err := sarama.NewClusterAdmin(brokers, config) //c.Kafka.Brokers is of type: []string
	if err != nil {
		fmt.Println("error is", err)
		return false
	}
	detail := sarama.TopicDetail{NumPartitions: 64, ReplicationFactor: 1}
	err = admin.CreateTopic("ktopic-xgboost", &detail, true)
	if err != nil {
		fmt.Println("error is", err)
		return false
	} else {
		return true
	}

}
func deleteTopic() bool {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0

	brokers := []string{"broker:9092"}
	admin, err := sarama.NewClusterAdmin(brokers, config) //c.Kafka.Brokers is of type: []string
	if err != nil {
		fmt.Println("error is", err)
		return false
	}

	err = admin.DeleteTopic("ktopic-xgboost")
	if err != nil {
		fmt.Println("error is", err)
		return false
	}

	err = admin.Close()
	if err != nil {
		fmt.Println("error is", err)
		return false
	}

	return true
}
func xgb_DeleteTopicHandler(ctx *fasthttp.RequestCtx) {
	ret := deleteTopic()
	if ret {
		log.Printf("Deleted topic")
	} else {
		log.Printf("Failed to delete topic")
	}
}
func xgb_StartTestHandler(ctx *fasthttp.RequestCtx) {

	host, err0 := os.Hostname()
	if err0 != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err0.Error())
		return
	}
	fmt.Fprintf(ctx, "Hello from %v - xgb_StartTestHandler\n\n", host)

	testinProgress = true
	log.Printf("  In xgb_StartTestHandler")
	ret := xgb_CreateTopic() //to do handle errors
	if ret {
		log.Printf(" Created topic")
	} else {
		log.Printf(" Failed to create topic")
	}

	fmt.Fprintf(ctx, "Hello xgb_StartTestHandler1 %d!\n", numberOfMessages)
	numberOfPods = string(ctx.QueryArgs().Peek("numberOfPods"))
	numberOfCpus = string(ctx.QueryArgs().Peek("cpusPerPod"))
	numKAFKAmessages = string(ctx.QueryArgs().Peek("numKAFKAmessages"))
	dataset = string(ctx.QueryArgs().Peek("dataset"))
	minio_service_ip = string(ctx.QueryArgs().Peek("minioSERVICEip"))
	lambda_param = string(ctx.QueryArgs().Peek("loadgen_lambda"))

	log.Printf("numberOfPods: %s\n", numberOfPods)
	log.Printf("numberOfCpus: %s\n", numberOfCpus)
	log.Printf("numKAFKAmessages: %s\n", numKAFKAmessages)
	log.Printf("dataset: %s\n", dataset)
	log.Printf("loadgen_lambda: %s\n", lambda_param)

	numberOfMessages, err1 := strconv.Atoi(numKAFKAmessages)
	lambda, err1 := strconv.ParseFloat(lambda_param, 64)
	if err1 != nil {
		fmt.Println(err1)
		os.Exit(2)
	}

	log.Printf("numberOfMessages: %d\n", numberOfMessages)
	fmt.Fprintf(ctx, "Hello numberOfMessages %d!\n", numberOfMessages)

	time.Sleep(30 * time.Second)
	podsStartTime = time.Now()
	cmd := exec.Command("python3", "/root/app/changecluster.py", "create", numberOfPods, numberOfCpus, minio_service_ip) //Run python script with function to run, in case of create number of pods and number of cpus per pod
	log.Printf("%s\n", "Running")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Wait for the Python program to exit.
	err2 := cmd.Run()
	log.Printf("Done with changecluster.py...")
	if err2 != nil {
		log.Printf("Finished when attempting changecluster.py: %s", err2)
		log.Fatalf("cmd.Run() failed with %s\n", err2)
	}

	go startKafkaProducer_PoissonGenerator(lambda, numberOfMessages)
	go startKafkaConsumer(numberOfMessages)

	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if string(stdout.Bytes()) == "error" {
		log.Printf("outStr:%s", string(stdout.Bytes()))
		log.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
	} else {
		ctx.SetStatusCode(fasthttp.StatusAccepted)
	}
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

func startKafkaConsumer(numMessages int) {
	log.Printf("In startKafkaConsumer")

	var errors int
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Specify brokers address. This is default one
	brokers := []string{"broker:9092"}

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	topic := "ktopic-xgboost-done"
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	quit := make(chan bool)

	// Count how many message processed
	msgCount := 0
	errors = 0

	// Get signal for finish
	doneCh := make(chan struct{})
	go func() {
		tnx_duration := make([]float64, numMessages)

		log.Printf("In Go func")
		for {
			select {
			case err := <-consumer.Errors():
				errors++
				fmt.Println(err)
			case msg := <-consumer.Messages():
				msgCount++
				log.Printf("msgCount:%d", msgCount)
				fmt.Println("Received messages", string(msg.Key), string(msg.Value))

				var tr transcodeResult
				resultJson = string(msg.Value)
				json.Unmarshal([]byte(resultJson), &tr)

				processTime := time.Now()
				timeProcess := processTime.Sub(processStartTime).Seconds()
				processDuration := strconv.FormatFloat(timeProcess, 'f', 2, 64)
				log.Println(tr.Key + ":" + processDuration + ", ")

				message_startqueue := tr.Key[:strings.IndexByte(tr.Key, '_')]
				message_startqueue_int, err := strconv.Atoi(message_startqueue)
				if err != nil {
					panic(err)
				}
				transaction_duration := int(time.Now().Unix()) - message_startqueue_int
				tnx_duration[msgCount-1] = float64(transaction_duration)
				log.Println("transaction: " + tr.Key + " " + strconv.Itoa(int(time.Now().Unix())) + " " + message_startqueue + " " + strconv.Itoa(transaction_duration))

				resultBuf.WriteString(tr.Key + ":" + processDuration + " " + tr.Duration + " " + tr.PodID + " " + strconv.Itoa(int(time.Now().Unix())) + " " + strconv.Itoa(transaction_duration) + ", ")

				if timeProcess > max_timeProcess {
					max_timeProcess = timeProcess
				}

				if msgCount == numMessages {
					testinProgress = false
					podsEndTime = time.Now()
					timeDiffsetup := podsEndTime.Sub(podsStartTime).Seconds()
					timeDiffprocess := podsEndTime.Sub(processStartTime).Seconds()
					Throughput_mins := float64((float64(msgCount) / max_timeProcess) * 60)

					podsTotalDuration := strconv.FormatFloat(timeDiffsetup, 'f', 2, 64)
					processTotalDuration := strconv.FormatFloat(timeDiffprocess, 'f', 2, 64)
					max_processDuration := strconv.FormatFloat(max_timeProcess, 'f', 2, 64)

					fmt.Println("processTotalDuration: " + processTotalDuration)
					resultBuf.WriteString("TotalTimeWithSetup:" + podsTotalDuration + ", ")
					resultBuf.WriteString("TotalDuration:" + max_processDuration + ", ")
					resultBuf.WriteString("NumberOfPods:" + numberOfPods + ", ")
					resultBuf.WriteString("vCPUsperPod:" + numberOfCpus + ", ")
					resultBuf.WriteString("ExpectedKAFKAmessages:" + numKAFKAmessages + ", ")
					resultBuf.WriteString("DeliveredKAFKAmessages:" + strconv.Itoa(msgCount) + ", ")
					resultBuf.WriteString("Dataset:" + dataset + ", ")

					resultBuf.WriteString("Throughput_tnx/min:" + strconv.FormatFloat(Throughput_mins, 'f', 2, 64) + ", ")
					Percentile90, _ := stats.Percentile(tnx_duration, 90)
					resultBuf.WriteString("90th_Percentile:" + strconv.FormatFloat(Percentile90, 'f', 2, 64) + ", ")
					Percentile95, _ := stats.Percentile(tnx_duration, 95)
					resultBuf.WriteString("95th_Percentile:" + strconv.FormatFloat(Percentile95, 'f', 2, 64) + ", ")
					mean_duration, _ := stats.Mean(tnx_duration)
					resultBuf.WriteString("mean_duration:" + strconv.FormatFloat(mean_duration, 'f', 2, 64) + ", ")
					min_duration, _ := stats.Min(tnx_duration)
					resultBuf.WriteString("min_duration:" + strconv.FormatFloat(min_duration, 'f', 2, 64) + ", ")
					max_duration, _ := stats.Max(tnx_duration)
					resultBuf.WriteString("max_duration:" + strconv.FormatFloat(max_duration, 'f', 2, 64) + ", ")
					stdev_duration, _ := stats.StandardDeviation(tnx_duration)
					resultBuf.WriteString("stdev_duration:" + strconv.FormatFloat(stdev_duration, 'f', 2, 64) + ", ")
					variance_duration, _ := stats.SampleVariance(tnx_duration)
					resultBuf.WriteString("variance_duration:" + strconv.FormatFloat(variance_duration, 'f', 2, 64))

					quit <- true
					break
				}

			case <-quit:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	log.Printf("Processed: %d; errors: %d\n", msgCount, errors)
}

func GaussianGenerator() {
	fmt.Println("=====Testing for GaussianGenerator begin=====")
	grng := rng.NewGaussianGenerator(time.Now().UnixNano())
	fmt.Println("Gaussian(5.0, 2.0): ")
	hist := map[int64]int{}
	for i := 0; i < 10000; i++ {
		hist[int64(grng.Gaussian(5.0, 2.0))]++
	}

	keys := []int64{}
	for k := range hist {
		keys = append(keys, k)
	}
	keys = astikit.SortInt64Slice(keys)

	for _, key := range keys {
		fmt.Printf("%d:\t%s\n", key, strings.Repeat("*", hist[key]/200))
	}

	fmt.Println("=====Testing for GaussianGenerator end=====")
	fmt.Println()
}

func startKafkaProducer_PoissonGenerator(lambda float64, numMessages int) {

	// Setup configuration
	log.Printf("In startKafkaProducer")
	config := sarama.NewConfig()
	// Return specifies what channels will be populated.
	// If they are set to true, you must read from
	// config.Producer.Return.Successes = true
	// The total number of times to retry sending a message (default 3).
	config.Producer.Retry.Max = 3
	// First, we tell the producer that we are going to partition ourselves.
	//config.Producer.Partitioner = sarama.NewManualPartitioner
	//config.Producer.Return.Successes = true
	// The level of acknowledgement reliability needed from the broker.
	config.Producer.RequiredAcks = sarama.WaitForAll

	brokers := []string{"broker:9092"}
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Minute)

	defer func() {
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	log.Printf("%s\n", "Before channel")
	var msg_id_enqueued, errors int
	msg_id_enqueued = 0
	errors = 0
	quit := make(chan bool)

	// Get signal for finish
	doneCh := make(chan struct{})
	processStartTime = time.Now()

	go func() {
		log.Printf("%s\n", "Running...")
		log.Printf("Go funcPoisson Produced message:%d", msg_id_enqueued)
		log.Printf("Go funcPoisson numMessages:%d", numMessages)
		log.Printf("Go funcPoisson dataset:%s", dataset)
		fmt.Println("===== PoissonGenerator begin =====")
		seed := int64(1582221789833178776)
		fmt.Println("Poisson (lambda): ", lambda)
		fmt.Println("number_of_messages: ", numMessages)
		fmt.Println("Poisson (seed): ", seed)

		prng := rng.NewPoissonGenerator(seed)
		hist := map[int64]int{}
		nextTime := int64(0)
		for i := 0; i < numMessages+2; i++ {
			nextTime = int64(prng.Poisson(lambda))
			hist[nextTime]++
			time.Sleep(time.Duration(nextTime) * time.Second)

			//generate the message
			log.Printf(" msg_id_enqueued=%d nextTime=%d \n", msg_id_enqueued, nextTime)

			strTime := strconv.Itoa(int(time.Now().Unix()))
			value := fmt.Sprintf(`{"dataset": "%s/%s", "hw": "cpu", "data": "%s", "ompcpus": "%s"}`, strTime+"_"+strconv.Itoa(msg_id_enqueued), dataset, dataset, numberOfCpus)
			log.Printf("Message queued to be sent: %s\n", value)

			msg := &sarama.ProducerMessage{
				Topic: "ktopic-xgboost",
				Key:   sarama.ByteEncoder(strconv.Itoa(msg_id_enqueued)),
				Value: sarama.ByteEncoder(value),
			}

			select {
			case producer.Input() <- msg:
				msg_id_enqueued++
				log.Printf("case Input: Produced message:%d out of numMessages:%d", msg_id_enqueued, numMessages)
				if msg_id_enqueued == numMessages {
					fmt.Println("===== PoissonGeneration Summary =====")
					keys := []int64{}
					for k := range hist {
						keys = append(keys, k)
					}
					//sort the Slice
					int64AsIntValues := make([]int, len(keys))
					for i, val := range keys {
						int64AsIntValues[i] = int(val)
					}
					sort.Ints(int64AsIntValues)
					for i, val := range int64AsIntValues {
						keys[i] = int64(val)
					}
					for _, key := range keys {
						fmt.Printf("%d|\t%d|\t%s\n", key, hist[key], strings.Repeat("*", hist[key]))
					}

					quit <- true
					break
				}

			case err := <-producer.Errors():
				errors++
				fmt.Println("Failed to produce message:", err)
				fmt.Println("===== Kafka producer is not stable - please repeat the test - =====")
				panic(err)
				break

			case <-quit:
				log.Printf("In Quit")
				doneCh <- struct{}{}
			} // select case
		}

	}()

	<-doneCh
	log.Printf("msg_id_enqueued: %d; errors: %d\n", msg_id_enqueued, errors)

}

//==============

func startKafkaProducer(numMessages int) {

	// Setup configuration
	log.Printf("In startKafkaProducer")
	config := sarama.NewConfig()
	// Return specifies what channels will be populated.
	// If they are set to true, you must read from
	// config.Producer.Return.Successes = true
	// The total number of times to retry sending a message (default 3).
	config.Producer.Retry.Max = 3
	// First, we tell the producer that we are going to partition ourselves.
	//config.Producer.Partitioner = sarama.NewManualPartitioner
	//config.Producer.Return.Successes = true
	// The level of acknowledgement reliability needed from the broker.
	config.Producer.RequiredAcks = sarama.WaitForAll
	brokers := []string{"broker:9092"}
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		// Should not reach here
		panic(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			// Should not reach here
			panic(err)
		}
	}()

	log.Printf("%s\n", "Before channel")
	var msg_id_enqueued, errors int
	msg_id_enqueued = 0
	errors = 0
	quit := make(chan bool)

	// Count how many message processed
	// Get signal for finish
	doneCh := make(chan struct{})

	go func() {
		log.Printf("%s\n", "Running")
		log.Printf("Go func Produced message:%d", msg_id_enqueued)
		log.Printf("Go func numMessages:%d", numMessages)
		log.Printf("Go func dataset:%s", dataset)
		for {
			log.Printf(" msg_id_enqueued=%d\n", msg_id_enqueued)
			time.Sleep(500 * time.Millisecond)
			strTime := strconv.Itoa(int(time.Now().Unix()))

			value := fmt.Sprintf(`{"dataset": "%s/%s", "hw": "cpu", "data": "%s", "ompcpus": "%s"}`, strTime+"_"+strconv.Itoa(msg_id_enqueued), dataset, dataset, numberOfCpus)
			log.Printf("Message queued to be sent: %s\n", value)

			msg := &sarama.ProducerMessage{
				Topic: "ktopic-xgboost",
				Key:   sarama.ByteEncoder(strconv.Itoa(msg_id_enqueued)),
				Value: sarama.ByteEncoder(value),
			}
			select {
			case producer.Input() <- msg:
				msg_id_enqueued++
				log.Printf("case Input: Produced message:%d", msg_id_enqueued)
				log.Printf("case Input: numMessages:%d", numMessages)
				if msg_id_enqueued == numMessages {
					quit <- true
					break
				}
			case err := <-producer.Errors():
				errors++
				fmt.Println("Failed to produce message:", err)
			case <-quit:
				log.Printf("In Quit")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	log.Printf("msg_id_enqueued: %d; errors: %d\n", msg_id_enqueued, errors)
}


