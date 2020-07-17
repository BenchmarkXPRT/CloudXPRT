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
	"strings"
)

func detectLinux() int {
	out, err := exec.Command("hostnamectl").Output()
	if err != nil {
		return 0
	}

	if strings.Contains(strings.ToLower(string(out)), "centos") {
		return 2
	} else if strings.Contains(strings.ToLower(string(out)), "ubuntu") {
		return 1
	}
	return 0
}

func main() {
	// call kubespray reset
	fmt.Println("Running kubespray reset ......")
	cmd := exec.Command("ansible-playbook", "-i", "./kubespray/inventory/cnb-cluster/inventory.yaml",
		"--become", "--become-user=root", "./kubespray/reset.yml",
		"--extra-vars", "reset_confirmation=yes")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err.Error())
	}

	// 0: unknown linux distribution, 1: Ubuntu, 2: CentOS
	osType := detectLinux()

	// remove docker and images
	fmt.Println("Remove Docker and its images ......")
	if osType == 1 {
		cmd = exec.Command("apt-get", "purge", "-y", "docker-ce", "docker-ce-cli", "--allow-change-held-packages")
	} else if osType == 2 {
		cmd = exec.Command("yum", "remove", "-y", "docker-ce", "docker-ce-cli")
	} else {
		log.Fatal("Linux distribution not support by CNB")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = exec.Command("rm", "-rf", "/var/lib/docker").Run()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = exec.Command("rm", "-rf", "/etc/docker").Run()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = exec.Command("rm", "-rf", "/var/lib/dockershim").Run()
	if err != nil {
		log.Fatal(err.Error())
	}

	if osType == 1 {
		err = exec.Command("apt", "autoremove", "-y").Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// If we backup hostname file before, copy it back
	var hostBack = "./hostname.back"
	if _, err = os.Stat(hostBack); err == nil {
		fmt.Println("Restore system hostname ......")
		err := exec.Command("cp", hostBack, "/etc/hostname").Run()
		if err != nil {
			log.Fatal(err.Error())
		}
		content, err := ioutil.ReadFile(hostBack)
		if err != nil {
			log.Fatalf("Error reading all.yaml file, %s", err)
		}
		err = exec.Command("hostname", strings.TrimSpace(string(content))).Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// If kube-manifests directory are created, remove it
	if _, err = os.Stat("/root/kube-manifests"); err == nil {
		fmt.Println("Remove /root/kube-manifests directory ......")
		err = exec.Command("rm", "-rf", "/root/kube-manifests").Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
