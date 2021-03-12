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

// Example static file server.
//
// Serves static files from the given directory.
// Exports various stats at /stats .
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
)

type record struct {
	Name   string  `json:"name"`
	Addr   string  `json:"addr"`
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type mcrecord struct {
	Name    string     `json:"name"`
	Date    string     `json:"date"`
	Elapsed float64    `json:"elapsed"`
	Ops     float64    `json:"ops"`
	Results []mcresult `json:"results"`
}

type mcresult struct {
	StockPrice     float32 `json:"stockprice"`
	StrikePrice    float32 `json:"strikeprice"`
	OptionYear     float32 `json:"optionyear"`
	CallResult     float32 `json:"callresult"`
	CallConfidence float32 `json:"callconfidence"`
}

type user struct {
	ID       string `json:"id"`       // in the format of user_123
	Password string `json:"password"` // in the format of pass_123
	Email    string `json:"email"`
}

var (
	cnbaddr            = flag.String("addr", ":8070", "TCP address to listen to")
	cnbaddrTLS         = flag.String("addrTLS", ":8443", "TCP address to listen to TLS (aka SSL or HTTPS) requests. Leave empty for disabling TLS")
	byteRange          = flag.Bool("byteRange", false, "Enables byte range requests if set to true")
	certFile           = flag.String("certFile", "./cnbserver.crt", "Path to TLS certificate file")
	compress           = flag.Bool("compress", false, "Enables transparent response compression if set to true")
	dir                = flag.String("dir", "/usr/share/nginx/html", "Directory to serve static files from")
	generateIndexPages = flag.Bool("generateIndexPages", true, "Whether to generate directory index pages")
	keyFile            = flag.String("keyFile", "./cnbserver.key", "Path to TLS key file")
)

const (
	RANDOMTEXT = `bfeWQobe:9oY+vO.DjRV:V@JBZ)Fvj5UqWg?BrppM:'u0/y[cgi9_<,L4I8mQ>\
13fnp|@zH{'R7_VFuZ4M3Om@hMCh@suVsxA(3msi2oVS{kc)hTyc[#tRBZ"isOH
(d%nWAR48*rqFKEQH&pWIgl*DLMpbXah3C]T4[|Rq\@{6w"b<Q'<i\D,||W"ECx
!sCmnp,qeNK^46L1tE8n12"1Z,<gGh-2!R\*y,|<Mcl-01v$U\w*09qW9o>Bn+q
WWw+yW0jp7SEN8HE2Dr.*yv_yFC'P3B'$f6&)8j%G7lqR[Fn0dKw^I'-'CH\gIB
<8:zK0t<U&bzh2E)2T"|fR*:*qPjYy|!u@nq]Ch;r,=ddh?t@-,f,p&|g,@C'W1
q*G7X!pBci_NEfR)7NMqmkj=Ai8K1zgTH/IA9JgesNNXe81>^{)[Yf)[Y>+5;43
9B\':)5p3^{;6R,-vxo9dkH$I2bNI#62,'+,<OgG*)iSJ3/pv^xF2,l2tz!!cx{
:.V$SD?*ePM';x7).p0,7q>S#07{9J6pn0cP3P_-nv__Jyf;Kzz[GhXI8Ci\2?K
4h4|;>!PD'^H?F^2,QgX&k.6SgCJTM>KMnqS!hC?8s+N"O.EB'*2'$I=O|p.qmg
-.p>8DXn,CFP#w^J4?xrx${1@W)xoww9NMwXSbQ:\:^{"{vtVZ9dcV+X/S&N\61`

	genPageTitle = "General Page Title"

	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Parse command-line flags.
	flag.Parse()

	// Setup FS handler
	fs := &fasthttp.FS{
		Root:               *dir,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: *generateIndexPages,
		Compress:           *compress,
		AcceptByteRange:    *byteRange,
	}
	fsHandler := fs.NewRequestHandler()

	// Create RequestHandler serving server stats on /stats and files
	// on other requested paths.
	// /stats output may be filtered using regexps. For example:
	//
	//   * /stats?r=fs will show only stats (expvars) containing 'fs'
	//     in their names.
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		idx := strings.LastIndex(path, "/")
		// Path like /users/user_1 => /users
		if idx > 0 {
			path = path[0:idx]
		}
		switch path {
		case "/stats":
			expvarhandler.ExpvarHandler(ctx)
		case "/calc":
			calculateHandler(ctx)
		case "/users":
			userHandler(ctx)
		case "/crypt":
			cryptHandler(ctx)
		case "/infer":
			inferenceHandler(ctx)
		case "/ocr":
			ocrHandler(ctx)
		case "/mc":
			mcHandler(ctx)
		case "/general":
			generalHandler(ctx)
		default:
			fsHandler(ctx)
		}
	}

	s := &fasthttp.Server{
		Handler:     requestHandler,
		Concurrency: fasthttp.DefaultConcurrency,
	}
	// s.DisableKeepalive = true

	// Start CNB HTTP server.
	if len(*cnbaddr) > 0 {
		log.Printf("Starting CNB HTTP server on %q", *cnbaddr)
		go func() {
			if err := s.ListenAndServe(*cnbaddr); err != nil {
				log.Fatalf("error in CNB Server  ListenAndServe: %s", err)
			}
		}()
	}

	// Start CNB HTTPS server.
	if len(*cnbaddrTLS) > 0 {
		log.Printf("Starting CNB HTTPS server on %q", *cnbaddrTLS)
		go func() {
			if err := fasthttp.ListenAndServeTLS(*cnbaddrTLS, *certFile, *keyFile, requestHandler); err != nil {
				log.Fatalf("error in CNB Server ListenAndServeTLS: %s", err)
			}
		}()
	}

	// Wait and serve.
	select {}
}

func calculateHandler(ctx *fasthttp.RequestCtx) {
	host, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
		return
	}
	fmt.Fprintf(ctx, "Hello from %v!\n\n", host)

	// Could be changed to cal-service.default.svc.cluster.local:8072 for kubernetes
	// Use localhost:8072 to run under docker --network="host"
	// resp, err := http.Get("http://localhost:8072")
	resp, err := http.Get("http://cal-service.default.svc.cluster.local:8072")
	if err != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response: %d\n", resp.StatusCode)
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(ctx, "Error response: %s!\n\n", err.Error())
		}
		fmt.Fprintf(ctx, "\n %s!\n\n", string(body))
	}
}

func cryptHandler(ctx *fasthttp.RequestCtx) {
	host, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
		return
	}
	fmt.Fprintf(ctx, "Hello from %v!\n\n", host)

	// Could be changed to crypt-service.default.svc.cluster.local:8076 for kubernetes
	// Use localhost:8076 to run under docker --network="host"
	// resp, err := http.Get("http://localhost:8076")
	resp, err := http.Get("http://crypt-service.default.svc.cluster.local:8076")
	if err != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response: %d\n", resp.StatusCode)
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(ctx, "Error response: %s!\n\n", err.Error())
		}
		fmt.Fprintf(ctx, "OK! Run result: \n %s\n\n", string(body))
	}
}

func userHandler(ctx *fasthttp.RequestCtx) {
	// Could be changed to http://user-service.default.svc.cluster.local:8079 for kubernetes
	// Use localhost:8079 to run under docker --network="host"
	// resp, err := http.Get("http://localhost:8079")
	url := "http://user-service.default.svc.cluster.local:8079" + string(ctx.Path())

	switch string(ctx.Method()) {
	case "POST":
		fallthrough
	case "PUT":
		req, err := http.NewRequest(string(ctx.Method()), url, bytes.NewBuffer(ctx.PostBody()))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(ctx, "Error response: %d\n", resp.StatusCode)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(ctx, "Error response: %s!\n\n", err.Error())
			}
			fmt.Fprintf(ctx, "%s", string(body))
		}
	case "GET":
		fallthrough
	case "DELETE":
		req, err := http.NewRequest(string(ctx.Method()), url, nil)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(ctx, "Error response: %d\n", resp.StatusCode)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(ctx, "Error response: %s!\n\n", err.Error())
			}
			fmt.Fprintf(ctx, "%s", string(body))
		}

	default:
		fmt.Fprintf(ctx, "Method %s is not supported\n", string(ctx.Method()))
	}
}

func inferenceHandler(ctx *fasthttp.RequestCtx) {
	host, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
		return
	}
	fmt.Fprintf(ctx, "Hello from %v!\n\n", host)

	// Could be changed to inf-service.default.svc.cluster.local:8071 for kubernetes
	// Use localhost:8071 to run under docker --network="host"
	// resp, err := http.Get("http://localhost:8071")
	resp, err := http.Get("http://inf-service.default.svc.cluster.local:8071")
	if err != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response: %d\n", resp.StatusCode)
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(ctx, "Error response: %s!\n\n", err.Error())
		}
		fmt.Fprintf(ctx, "OK! Run result: \n %s\n\n", string(body))
	}
}

func printCtx(ctx *fasthttp.RequestCtx, input string, demo bool) {
	if demo == false {
		fmt.Fprintln(ctx, input)
	}
}

func createOneUser(name string) []byte {
	var record = user{ID: name, Password: userPass(name),
		Email: name + "@" + randString() + ".com"}
	result, _ := json.Marshal(record)
	return result
}

func mcHandler(ctx *fasthttp.RequestCtx) {
	var demoMode bool
	name := string(ctx.QueryArgs().Peek("name"))
	if len(name) > 0 {
		// request from UI
		demoMode = true
	} else {
		// request from load generator
		demoMode = false
		name = randUser()
	}

	// Check if user exist in user db, demo mode should exist, benchmark mode may need create user
	startTime := time.Now()
	respU, err := http.Get("http://user-service.default.svc.cluster.local:8079/users/" + name)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from user service: %s\n", err.Error())
		return
	}
	defer respU.Body.Close()

	if respU.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response code from user service: %d\n", respU.StatusCode)
		return
	}

	bodyTemp, err := ioutil.ReadAll(respU.Body)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from user service: %s!\n\n", err.Error())
		return
	}
	bodyU := strings.TrimSpace(string(bodyTemp))
	// No user found, we need create a new user here
	if len(bodyU) == 0 {
		url := "http://user-service.default.svc.cluster.local:8079/users"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(createOneUser(name)))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(ctx, "Error response when create user: %s!\n\n", err.Error())
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(ctx, "Error response code when create user: %d\n", resp.StatusCode)
			return
		}
		userBody, _ := ioutil.ReadAll(resp.Body)
		printCtx(ctx, string(userBody), demoMode)
	}
	printCtx(ctx, fmt.Sprintf("User check/creation finished successfully, time used %s\n",
		time.Since(startTime)), demoMode)
	// End of user management

	startTime = time.Now()
	// Next step encryption/decryption user profile data
	resp, err := http.Get("http://crypt-service.default.svc.cluster.local:8076?name=" + name)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from decryption service: %s\n", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response code from decryption service: %d\n", resp.StatusCode)
		return
	}

	bodyE, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from decryption service: %s!\n\n", err.Error())
		return
	}
	if !strings.Contains(string(bodyE), "successfully") {
		fmt.Fprintf(ctx, "Errors happen during decryption: %s!\n\n", string(bodyE))
		return
	}
	printCtx(ctx, fmt.Sprintf("Decryption finished successfully, time used %s\n",
		time.Since(startTime)), demoMode)

	// Could be changed to mc-service.default.svc.cluster.local:8074 for kubernetes
	// Use localhost:8074 to run under docker --network="host"
	// resp, err := http.Get("http://localhost:8074")
	startTime = time.Now()
	respMC, err := http.Get("http://mc-service.default.svc.cluster.local:8074?name=" + name)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from Monte Carlo service: %s\n", err.Error())
		return
	}

	defer respMC.Body.Close()
	if respMC.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response code from Monte Carlo service: %d\n", respMC.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(respMC.Body)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from Monte Carlo service: %s!\n\n", err.Error())
		return
	}
	printCtx(ctx, fmt.Sprintf("Monte Carlo finished successfully, time used %s\n",
		time.Since(startTime)), demoMode)
	printCtx(ctx, fmt.Sprintf("Run result: \n %s\n", string(body)), demoMode)

	// Extract mcrecord from MC calculation results
	var re = &mcrecord{
		Name: name,
		Date: timeNow(),
	}
	err = parseMCResults(re, string(body))
	if err != nil {
		fmt.Fprintf(ctx, "Monte Carlo results format error %s won't be stored in DB", err.Error())
		return
	}

	url := "http://db-service.default.svc.cluster.local:8078/mc"
	client := &http.Client{}

	startTime = time.Now()
	jsonStr, _ := json.Marshal(re)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	respDB, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from DB service: %s\n", err.Error())
		return
	}

	defer respDB.Body.Close()
	if respDB.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response code from DB service: %d\n", respDB.StatusCode)
		return
	}

	bodyDB, err := ioutil.ReadAll(respDB.Body)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from DB service: %s!\n\n", err.Error())
		return
	}
	if !strings.Contains(string(bodyDB), "successfully") {
		fmt.Fprintf(ctx, "Errors happen during store to db: %s!\n\n", string(bodyDB))
		return
	}
	printCtx(ctx, fmt.Sprintf("Store to DB finished successfully, time used %s\n\nMC complete!\n",
		time.Since(startTime)), demoMode)
	// Or just return json format MC calculation results here if in demo mode
	results, _ := json.Marshal((*re).Results)
	if demoMode {
		fmt.Fprintln(ctx, string(results))
	}
}

func generalHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "%s\n", genPageTitle)

	round := rand.Intn(5) + 5
	for i := 0; i < round; i++ {
		fmt.Fprint(ctx, RANDOMTEXT)
	}
	fmt.Fprint(ctx, "\n")
}

func ocrHandler(ctx *fasthttp.RequestCtx) {
	host, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(ctx, "Error response: %s\n", err.Error())
		return
	}
	fmt.Fprintf(ctx, "Hello from %v!\n\n", host)

	startTime := time.Now()
	// First step encryption/decryption
	resp, err := http.Get("http://crypt-service.default.svc.cluster.local:8076")
	if err != nil {
		fmt.Fprintf(ctx, "Error response from decryption service: %s\n", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response code from decryption service: %d\n", resp.StatusCode)
		return
	}

	bodyE, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from decryption service: %s!\n\n", err.Error())
		return
	}
	if !strings.Contains(string(bodyE), "successfully") {
		fmt.Fprintf(ctx, "Errors happen during decryption: %s!\n\n", string(bodyE))
		return
	}
	fmt.Fprintf(ctx, "Decryption finished successfully, time used %s\n\n",
		time.Since(startTime))

	// Could be changed to ocr-service.default.svc.cluster.local:8073 for kubernetes
	// Use localhost:8073 to run under docker --network="host"
	// resp, err := http.Get("http://localhost:8073")
	startTime = time.Now()
	respOCR, err := http.Get("http://ocr-service.default.svc.cluster.local:8073")
	if err != nil {
		fmt.Fprintf(ctx, "Error response from OCR service: %s\n", err.Error())
		return
	}

	defer respOCR.Body.Close()
	if respOCR.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response code from OCR service: %d\n", respOCR.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(respOCR.Body)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from OCR service: %s!\n\n", err.Error())
		return
	}
	fmt.Fprintf(ctx, "OCR finished successfully, time used %s\n\n",
		time.Since(startTime))
	// TODO add check to OCR results and make use of contents of body!
	fmt.Fprintf(ctx, "OK! Run result: \n %s\n\n", string(body))

	// Extract address and total amount from OCR results
	addr, amount, err := parseOCRResults(string(body))
	if err != nil {
		fmt.Fprintf(ctx, "Receipt format error %s won't be stored in DB", err.Error())
		return
	}
	var re = &record{
		Name:   randUser(),
		Addr:   addr,
		Date:   ranDate(),
		Amount: amount,
	}
	url := "http://db-service.default.svc.cluster.local:8078/trans"
	client := &http.Client{}

	startTime = time.Now()
	jsonStr, _ := json.Marshal(re)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	respDB, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from DB service: %s\n", err.Error())
		return
	}

	defer respDB.Body.Close()
	if respDB.StatusCode != http.StatusOK {
		fmt.Fprintf(ctx, "Error response code from DB service: %d\n", respDB.StatusCode)
		return
	}

	bodyDB, err := ioutil.ReadAll(respDB.Body)
	if err != nil {
		fmt.Fprintf(ctx, "Error response from DB service: %s!\n\n", err.Error())
		return
	}
	if !strings.Contains(string(bodyDB), "successfully") {
		fmt.Fprintf(ctx, "Errors happen during store to db: %s!\n\n", string(bodyDB))
		return
	}
	fmt.Fprintf(ctx, "Store to DB finished successfully, time used %s\n\nOCR complete!\n\n",
		time.Since(startTime))
}

func parseOCRResults(body string) (string, float64, error) {
	// If not a valid CNB receipt format, do not continue
	if !strings.Contains(body, "CNB Company") {
		return "", -1, errors.New("Not a valid CNB receipt format")
	}

	var addr, total string
	index := 0
	tokens := strings.Split(body, "\n")
	for idx, token := range tokens {
		if strings.Contains(token, "Some Company") {
			index = idx
		}
		if strings.Contains(token, "Total") && !strings.Contains(token, "Total Discount") {
			total = token[6:]
		}
	}
	addr = tokens[index+1] + ", " + tokens[index+2]
	amount, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return "", -1, err
	}
	return addr, amount, nil
}

func parseMCResults(re *mcrecord, body string) error {
	// If not a valid CNB receipt format, do not continue
	if !strings.Contains(body, "Monte Carlo") {
		return errors.New("Not a valid Monte Carlo result format")
	}

	var e, o string
	var results []mcresult
	tokens := strings.Split(body, "\n")
	for _, token := range tokens {
		if strings.Contains(token, "Time Elapsed") {
			e = token[15:]
		}
		if strings.Contains(token, "Opt/sec") {
			o = token[15:]
		}
		// Calculation result format is like:
		// StockPrice = 49.760925    OptionStrikePrice = 10.332694    OptionYears = 4.955018    CallResult = 42.054034    CallConfidence = 0.042905
		if strings.Contains(token, "StockPrice") {
			s := strings.Fields(token)
			sp, _ := strconv.ParseFloat(s[2], 32)
			st, _ := strconv.ParseFloat(s[5], 32)
			ye, _ := strconv.ParseFloat(s[8], 32)
			cr, _ := strconv.ParseFloat(s[11], 32)
			cc, _ := strconv.ParseFloat(s[14], 32)
			results = append(results, mcresult{float32(sp), float32(st),
				float32(ye), float32(cr), float32(cc)})
		}
	}

	elapsed, err := strconv.ParseFloat(e, 64)
	if err != nil {
		return err
	}
	ops, _ := strconv.ParseFloat(o, 64)
	if err != nil {
		return err
	}
	re.Elapsed = elapsed
	re.Ops = ops
	re.Results = results
	return nil
}

func randString() string {
	n := rand.Intn(6) + 5
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func ranDate() string {
	min := time.Date(2000, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2019, 7, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0).String()
}

func timeNow() string {
	return time.Now().Format(time.RFC850)
}

func randFloat(min, max float64) float64 {
	res := min + rand.Float64()*(max-min)
	res = math.Floor(res*100) / 100
	return res
}

func randUser() string {
	prefix := "user_"
	id := rand.Intn(1000)

	return fmt.Sprintf("%s%d", prefix, id)
}

func userPass(name string) string {
	uPrefix := "user"
	pPrefix := "pass"
	return strings.Replace(name, uPrefix, pPrefix, 1)
}
