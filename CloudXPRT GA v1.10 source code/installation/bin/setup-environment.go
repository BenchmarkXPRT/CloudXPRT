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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	httpProxy  = ""
	httpsProxy = ""
	noProxy    = false
	nodeType   = ""
	reboot     = false

	aptConf = `Acquire::http::proxy "%s";
Acquire::https::proxy "%s";`

	etcEnv = `http_proxy="%s"
https_proxy="%s"
no_proxy="localhost,127.0.0.1,%s,%s:6443,10.233.0.0/16"
`

	allYaml = `http_proxy: "%s"
https_proxy: "%s"
no_proxy: "localhost,127.0.0.1,%s,%s:6443,10.233.0.0/16"`

	allHTTPProxy  = `http_proxy: "%s"`
	allHTTPSProxy = `https_proxy: "%s"`
	allNoProxy    = `no_proxy: "localhost,127.0.0.1,%s,%s:6443,10.233.0.0/16"`
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Sets up the node's date and time, proxies, and saves original hostname\n\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  ./setup-environment -nodetype master -http_proxy <proxy> -https_proxy <proxy> -reboot
  ./setup-environment -nodetype master -noproxy
`)
	}

	flag.StringVar(&httpProxy, "http_proxy", "", "HTTP proxy to use")
	flag.StringVar(&httpsProxy, "https_proxy", "", "HTTPS proxy to use")
	flag.BoolVar(&noProxy, "noproxy", false, "If passed, no proxies will be set")
	flag.StringVar(&nodeType, "nodetype", "", "Type of node to configure, options are 'master' or 'worker'")
	flag.BoolVar(&reboot, "reboot", false, "If passed, node will be rebooted")
}

func inputCheck() {
	if noProxy && httpProxy != "" ||
		noProxy && httpsProxy != "" {
		fmt.Println("no_proxy flag was set, but proxies were given")
		flag.Usage()
		os.Exit(1)
	}

	if noProxy == false && httpProxy == "" {
		fmt.Println("http_proxy must be provided")
		flag.Usage()
		os.Exit(1)
	}

	if noProxy == false && httpsProxy == "" {
		fmt.Println("https_proxy must be provided")
		flag.Usage()
		os.Exit(1)
	}

	if nodeType == "" {
		fmt.Println("node_type must be provided")
		flag.Usage()
		os.Exit(1)
	}

	if nodeType != "master" && nodeType != "worker" {
		fmt.Println("node_type is not a valid type")
		flag.Usage()
		os.Exit(1)
	}
}

// Get preferred outbound ip of this machine
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")

	return localAddr[:idx]
}

func createAptConf(aptConfContent string) {
	// Remove the file if it exists already, best efforts
	os.Remove("/etc/apt/apt.conf")

	err := ioutil.WriteFile("/etc/apt/apt.conf", []byte(aptConfContent), 644)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func appendEtcEnv(input string) {
	// check if proxy has been set in this file, if so, do nothing
	content, err := ioutil.ReadFile("/etc/environment")
	if err != nil {
		log.Fatalf("Error reading /etc/environment file, %s", err)
	}

	if strings.Contains(string(content), input) {
		fmt.Println("You already have proxy settings in /etc/environment file.")
		return
	}

	f, err := os.OpenFile("/etc/environment", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()
	if _, err := f.WriteString(input); err != nil {
		log.Fatal(err.Error())
	}
}

func replaceAllYaml(input []string) {
	var filepath = "./kubespray/inventory/cnb-cluster/group_vars/all/all.yml"

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Error reading all.yaml file, %s", err)
	}

	// Only need replace no_proxy line in this file
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "http_proxy") || strings.HasPrefix(line, "# http_proxy") {
			lines[i] = input[0]
		}
		if strings.HasPrefix(line, "https_proxy") || strings.HasPrefix(line, "# https_proxy") {
			lines[i] = input[1]
		}
		if strings.HasPrefix(line, "no_proxy") || strings.HasPrefix(line, "# no_proxy") {
			lines[i] = input[2]
			break
		}
	}

	err = ioutil.WriteFile(filepath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func removeProxyAllYaml() {
	var filepath = "./kubespray/inventory/cnb-cluster/group_vars/all/all.yml"

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Error reading all.yaml file, %s", err)
	}

	// Only need replace no_proxy line in this file
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "http_proxy") {
			lines[i] = `# http_proxy: ""`
		}
		if strings.HasPrefix(line, "https_proxy") {
			lines[i] = `# https_proxy: ""`
		}
		if strings.HasPrefix(line, "no_proxy") {
			lines[i] = `# no_proxy: ""`
			break
		}
	}

	err = ioutil.WriteFile(filepath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func backupHostname() {
	// If the file is already there, do nothing
	if _, err := os.Stat("./hostname.back"); err == nil {
		return
	}
	err := exec.Command("cp", "/etc/hostname", "./hostname.back").Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func setDateTime() {
	_, err := exec.LookPath("date")
	if err != nil {
		log.Fatalf("Date binary not found, cannot set system date: %s\n", err.Error())
	}

	var formatTime string

	if !noProxy {
		formatTime = "date -s \"$(wget -qSO- --max-redirect=0 google.com 2>&1 -e use_proxy=yes -e http_proxy=" + httpProxy + " | grep Date: | cut -d' ' -f5-8)Z\""
	} else {
		formatTime = "date -s \"$(wget -qSO- --max-redirect=0 google.com 2>&1 | grep Date: | cut -d' ' -f5-8)Z\""
	}

	err = exec.Command("bash", "-c", formatTime).Run()

	if err != nil {
		log.Fatal(err.Error())
	}
}

func rebootNode() {
	_, err := exec.LookPath("shutdown")
	if err != nil {
		log.Fatalf("shutdown binary not found, cannot reboot this node: %s\n", err.Error())
	}

	err = exec.Command("shutdown", "-r", "now").Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}

// 0: unknown linux distribution, 1: Ubuntu or Debian, 2: CentOS or RedHat
func detectLinux() int {
	out, err := exec.Command("hostnamectl").Output()
	if err != nil {
		return 0
	}

	outmatch := strings.ToLower(string(out))
	if strings.Contains(outmatch, "centos") || strings.Contains(outmatch, "red hat") || strings.Contains(outmatch, "fedora") {
		return 2
	} else if strings.Contains(outmatch, "ubuntu") || strings.Contains(outmatch, "debian") {
		return 1
	}
	return 0
}

func main() {

	flag.Parse()
	inputCheck()

	osType := detectLinux()
	if osType == 0 {
		log.Fatal("Linux distribution not supported by CNB")
	}

	// backup /etc/hostname file
	fmt.Print("Creating hostname backup file\n\n")
	backupHostname()

	fmt.Print("Setting Date and Time from the Internet\n\n")
	setDateTime()

	if noProxy {
		fmt.Println("No proxy settings configured")
	} else {
		fmt.Println("Configuring proxy settings")

		ip := getOutboundIP()

		// Add proxy for APT, only for Ubuntu
		if osType == 1 {
			createAptConf(fmt.Sprintf(aptConf, httpProxy, httpsProxy))
		}

		// Add proxy for /etc/environment
		appendEtcEnv(fmt.Sprintf(etcEnv, httpProxy, httpsProxy, ip, ip))

		if nodeType == "master" {
			// Add proxy to Kubespray all.yml configuration file
			replaceAllYaml([]string{fmt.Sprintf(allHTTPProxy, httpProxy), fmt.Sprintf(allHTTPSProxy, httpsProxy),
				fmt.Sprintf(allNoProxy, ip, ip)})
		}

		if reboot {
			fmt.Println("Node will be rebooted in 5 seconds")
			time.Sleep(5 * time.Second)
			rebootNode()
		} else {
			fmt.Println("Please reboot your system for proxy changes to take effect!")
		}
	}
}
