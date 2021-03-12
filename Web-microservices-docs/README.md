# CloudXPRT Web Microservices Workload

- [Introduction](#introduction)
- [Configure and run on-premises](#configure-and-run-on-premises)
	- [Installation](#installation-on-onp)
	- [Run the benchmark](#run-onp)
- [Benchmark results](#benchmark-results)
- [Configure and run on Amazon Web Services](#configure-and-run-on-amazon-web-services)
	- [Create the virtual machines](#create-vms-aws)
	- [Run the benchmark](#run-aws)
	- [Save results](#save-results-aws)
- [Configure and run on Microsoft Azure](#configure-and-run-on-microsoft-azure)
	- [Create the virtual machines](#create-vms-azu)
	- [Run the benchmark](#run-azu)
	- [Save results](#save-results-azu)
- [Configure and run on Google Cloud Platform](#configure-and-run-on-google-cloud-platform)
	- [Create the virtual machines](#create-vms-gcp)
	- [Run the benchmark](#run-gcp)
	- [Save results](#save-results-gcp)
- [Running in demo mode with a UI](#demo-with-a-ui)
- [Build the benchmark from source](#build-the-benchmark-from-source)

## Introduction

In the web microservices workload, a simulated user logs in to a web application that does three things: provides a selection of stock options, performs Monte-Carlo simulations with those stocks, and presents the user with options that may be of interest. This scenario enables the workload to model a traditional three-tier web application with services in the web, application, and data layers. The workload uses Kubernetes, Docker, NGNIX, REDIS, Cassandra, and monitoring modules to mimic an end-to-end IaaS scenario.

The workload reports performance in transactions per second, which reflects the number of successful requests per second the stack achieves for each level of concurrency. Testers can use this workload’s metrics to compare IaaS stack performance and to evaluate whether any given stack is capable of meeting SLA thresholds.

### Terminology

- Node: A single machine or virtual machine
- Control-plane node: The node running the installation, which becomes the Kubernetes control-plane node.
- Worker node: Each machine that will join the Kubernetes cluster.

___

## Configure and run on-premises

The installation scripts in the installation directory will install and create a Kubernetes cluster using Kubespray. They will help you:

- configure your environment to run CloudXPRT,
- get the IP addresses for all machines in your cluster,
- set up passwordless SSH,
- install the required software for Ansible and Kubespray,
- create the cluster, and
- remove the cluster once you are done running CloudXPRT.

### Supported operating systems

- On-premises: Ubuntu 20.04.2 or later
- Cloud:       Ubuntu 18.04 or later

### Minimum requirements

We recommend running this benchmark on high-end servers because the benchmark scales in steps until it uses all available cores. For functional testing, you may use hosts or VMs with fewer resources as long as they have at least:

- 16 logical or virtual CPUs
- 8 GB RAM
- 10 GB of available disk space

### Installation

#### Set up the environment
In each machine in your cluster:

1. Set the sudo-user password if necessary. (**Note**: the passwords must be the same on each machine.)
	```
	sudo passwd sudouser
	```

2. Add the following line at the end of `/etc/sudoers` file.
	```
	sudouser ALL=(ALL) NOPASSWD: ALL
	```

3. Ensure that the openssh-server package is installed.
	```
	sudo apt-get install openssh-server -y
	```

4. Allow password authentication, if disabled.

	Edit `/etc/ssh/sshd_config` and uncomment and modify the `PasswordAuthentication` line so that it becomes -
	```
	PasswordAuthentication yes
	```
5. Restart `sshd`.
	```
	sudo service sshd restart
	```

#### Control-plane node

1. Go to the installation directory.
	```
	cd ~/CloudXPRT_vX.XX_web-microservices/installation
	```

2. Edit `cluster_config.json`.

	For each machine in the cluster, add its IPV4 address (required) and desired hostname (optional). Only one machine per `{..}` section should be in the `nodes` list. The control-lane node must be first.

	**Note**: Although optional, each hostname must be unique and can only contain **lowercase alphanumeric characters**. If hostnames are not provided, Kubespray will rename each host as `node1`, `node2`, ..., `nodeN`. This means that the control-plane node's hostname will be changed to `node1`.

	If your machines are behind a proxy, make sure the `set_proxy` option is set to `yes` and modify the `http_proxy` and `https_proxy` parameters to be the values used by your organization. These proxy settings will be applied on all the nodes so that they can communicate through the Kubernetes networking plugin. Furthermore, the servers must be restarted for these changes to take affect since `/etc/environment` is modified. The `prepare-cluster.sh` script can reboot all the nodes automatically, if the `reboot` option in `cluster_config.json` is set to `yes` (the default). If you set the reboot option to `no` nad you use a proxy, please manually restart your machines after running `prepare-cluster.sh` otherwise the cluster creation in the Step 3 will fail.

	Example configuration for a three node cluster:
	```
	"nodes": [
	   {
		"ip_address": "192.168.0.11",
		"hostname": "controlplane"
	   },
	   {
		"ip_address": "192.168.0.12",
		"hostname": "worker1"
	   },
	   {
		"ip_address": "192.168.0.13",
		"hostname": "worker2"
	   }
	],
	```

3. On the control-plane node, run the `prepare-cluster.sh` script to perform the necessary preparation steps.
	```
	sudo ./prepare-cluster.sh
	```

4. On the control-plane node, run the `create-cluster.sh` script to create the cluster.
	```
	sudo ./create-cluster.sh
	```

	This process may take anywhere from 6 to 20 minutes.

	**Notes**: If you get this error from the docker-ce repository '**RETRYING: ensure docker-ce repository public key is installed ...**'
	- Make sure that the proxies are set up correctly.
	- You may repeat the `prepare-cluster.sh` script to set this up again, or you may manually edit them in each node of your cluster.
	- You should also double check that the date and time are the same on all of the nodes.

	For more information on Kubespray and errors that occur during cluster creation, see its [documentation](https://github.com/kubernetes-sigs/kubespray).

#### Reset the Docker environment and the Kubernetes cluster

To remove the cluster and Docker installations on every node, run the `remove-cluster.sh` script on the control-plane node after you finish all the tests on the machine.
```
sudo ./remove-cluster.sh
```

Answer 'y' or 'yes' to the prompt.

Note: This will not remove the proxy settings. If you want to run CloudXPRT again, you can run the `create-cluster.sh` script to re-create the Kubernetes cluster.

### Run the benchmark

#### Configure benchmark parameters
```
cd ~/CloudXPRT_vX.XX_web-microservices/cnbrun
```

Edit the `config.json` file to set the parameters for the CloudXPRT web microservices workload.

- `iterations`: This setting lets the user select the number of iterations to run for each microservice. The default setting is 1. If a user designates 3, 5, 7, or 9 iterations, the postprocess binary (if the user has enabled it) produces a set of median values using the results from each run.
- `hpamode`: With the default setting of false, the workload creates the maximum number of pods from the beginning of the run. If the user sets this option to true, the workload uses Kubernetes Horizontal Pod Autoscaler (HPA) to scale pods as the load increases.
- `postprocess`: With the default setting of false, the workload does not run the postprocess binary at the end of a test run. If the user sets this option to true, the workload runs the postprocess binary at the end of a test run.
- `ppoutputfile`: If the user sets the postprocess binary to run at the end of a test, this setting directs the benchmark to save the postprocess results to a file. The default setting is "".
- `autoloader.initialclients`: This setting lets the user select the initial number of clients the load generator will create. The default is 1.
- `autoloader.clientstep`: This setting lets the user select the number of clients to increase for each load generator iteration. The default is 1.
- `autoloader.lastclient`: This setting lets the user designate the number of clients after which the load generator stops. At the default setting of -1, the load generator continues to run until the cluster is saturated (i.e., CPU ~100%).
- `autoloader.SLA`: This setting lets the user designate the SLA for a 95th percentile latency constraint.  The default setting is 3,000 milliseconds. If set to -1, there is no latency constraint.
- `autoloader.timeinterval`: This setting lets the user designate how long the load generator spends for each iteration with the specified number of clients. The default setting is 60 seconds.
- `workload.version`: This setting lets the user designate which Docker image version the test uses. The current default setting is v1.0.
- `workload.cpurequests`: This setting lets the user designate the number of CPU cores (integers only) the workload assigns to each pod. Currently, the benchmark supports values of 1, 2, and 4. The default setting is 4. Values of 1 and 2 are more appropriate for relatively low-end systems or configurations with few vCPUs.

**Note**: `cnbrun`, `gobench`, `autoloader`, and all shell scripts must have executable permissions.

#### Start the benchmark run

##### Running the load generator on the System Under Test (SUT)

Once the parameters in `config.json` are configured, run the `cnbrun` executable.
```
./cnbrun
```

To clean up benchmark-generated resources if the test run is interrupted, run the following script.
```
./cleanups.sh
```

To collect system information for the cluster, run the following script on the control-plane node only.
```
./system_info.sh
```

### Running the load generator outside of the SUT

**Requirement:** The machine running the load generator must be on the same network as the Kubernetes cluster.

Copy the following directories from the control-plane node of the cluster to the machine on which you want to run the load generator:
- `cnbrun` directory
- `$HOME/.kube` directory

On the load generator machine:

1. Move the `.kube` directory to the user's home directory.

2. Install `kubectl`.
	```
	sudo apt-get install kubectl
	```

3. Rename `mc.remote.sh` to `mc.sh`.
	```
	mv mc.remote.sh mc.sh
	```

4. Ensure that the `autoloader`, `cnbrun`, `cnbweb`, `gobench`, and `mc.sh` in the `cnbrun` directory have executable permissions.
	```
	chmod +x autoloader cnbrun cnbweb gobench mc.sh
	```

5. Once the parameters in `config.json` are configured, run the `cnbrun` executable.
	```
	./cnbrun
	```

___

## Benchmark results

After a run, test results will be located on the same node that ran the load generator. If you are testing a multi-node cluster, CloudXPRT will not save results files to the other nodes in the cluster. Results files will
accumulate in the `cnbrun/output` directory as you conduct runs, and CloudXPRT will not delete or overwrite them.

For each run, CloudXPRT automatically generates four files:

1. A log file with the results in a formatted table
2. A csv file with the results
3. A log file with all the stdout output during the run
4. A copy of the config file used for that run

#### Metrics

The results can be summarized using the following metrics:

1. Max successful requests per minute under a specified SLA
2. Total Max successful requests per minute (regardless of SLA)

Currently, testers must use the log file containing the formatted output table to manually record their chosen metrics.

By default, the max Service Level Agreement (SLA) for 95 percentile latency is 3 seconds (specified in `config.json`'s `autoloader.SLA` parameter). The load generator will stop when the system under test can no longer meet the specified SLA or when the SUT cannot beat the current maximum number of successful requests within 5 retries.

Below are the condensed results from a sample run. You can choose different SLAs of interest from the log file. For example, to determine throughput within SLAs of 1,000, 2,000, and 3,000 milliseconds, a tester could compare the number of successful requests that the SUT was able to execute within those times.

From the results, you can see that the system was able to handle:
- 604 requests per minute under 1000 ms,
- 889 requests per minute under 2000 ms, and
- 900 requests per minute under 3000 ms

The max throughput for this run is 900 successful requests per minute.

| CONCURRENCY | SUCC_REQS | ... | SUCC_REQS_RATE(REQ/S) | ... | MC_RESP_TIME(95%ile)(MS) |
| ----------- | --------  | --- | --------------------- | --- | ------------------------ |
|           1 |       50  | ... |                     1 | ... |                      607 |
|           2 |       98  | ... |                     3 | ... |                      667 |
|     ...     |   ...     | ... |         ...           | ... |           ...            |
|          17 |      587  | ... |                    19 | ... |                      942 |
|          18 |      604  | ... |                    20 | ... |                      987 |
|          19 |      624  | ... |                    20 | ... |                     1007 |
|          20 |      639  | ... |                    21 | ... |                     1093 |
|     ...     |   ...     | ... |         ...           | ... |           ...            |
|          49 |      884  | ... |                    29 | ... |                     1965 |
|          50 |      889  | ... |                    29 | ... |                     1994 |
|          51 |      889  | ... |                    29 | ... |                     2041 |
|          52 |      890  | ... |                    29 | ... |                     2189 |
|     ...     |   ...     | ... |         ...           | ... |           ...            |
|          59 |      897  | ... |                    29 | ... |                     2819 |
|          60 |      907  | ... |                    30 | ... |                     2706 |
|          61 |      900  | ... |                    30 | ... |                     2800 |
|          62 |      888  | ... |                    29 | ... |                     3032 |

___

## Configure and run on AWS
### Install a Kubernetes cluster with KOPS on AWS and run CloudXPRT

### Preparation

1. On a local Ubuntu Linux machine, create a new user, and switch to it.
	```
	sudo adduser awsuser
	sudo adduser awsuser sudo
	```

2. If you are using a GUI on the local Ubuntu machine, log out and log back in as `awsuser`. Otherwise, you can change to the new user directly by using the following command:
	```
	su - awsuser
	```

3. Download the Terraform binary from the [Terraform site](https://www.terraform.io/downloads.html) and install it in the `/usr/local/bin` directory.

4. Create an AWS IAM user and change permissions for this user. For help, refer to AWS documentation:
	- [Create AWS users](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html)
	- [Change permissions](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_change-permissions.html)

5. Assign the following permissions to the AWS IAM user:
	- AmazonEC2FullAccess
	- AmazonRoute53FullAccess
	- AmazonS3FullAccess
	- IAMFullAccess
	- AmazonVPCFullAccess

6. Install the AWS CLI.

	Refer to the AWS documentation for its [CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-linux.html).

7. Create access keys for the IAM user.

	Refer to the AWS documentation for help with [access keys](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html).

8. Configure your AWS CLI environment using the new access keys.
	```
	aws configure
	AWS Access Key ID [None]: AKIXXXXXXXXXXXXX
	AWS Secret Access Key [None]: sIrkzNOXxXXXXXXXXXXXxxxXXXX
	Default region name [None]: us-west-2
	Default output format [None]: json
	```

9. Verify that the AWS key and secret are stored in the `~/.aws/credentials` file. For example,
	```
	[default]
	aws_access_key_id = AKIXXXXXXXXXXXXX
	aws_secret_access_key = sIrkzNOXxXXXXXXXXXXXxxxXXXX
	```

10. Create SSH key pairs.
	```
	ssh-keygen
	```

### Create the virtual machines
1. Untar the CloudXPRT data-analytics code.
	```
	cd ~
	tar -xzf ~/CloudXPRT_vXXXX_web-microservices.tar.gz
	cd ~/CloudXPRT_vXXXX_web-microservices/terraform/aws
	```
2. Modify the `variables.tf` file to set the following VM parameters for your needs.
	- aws_region
	- aws_zones
	- instance_ami
	- instance_type
	- vm_name

3. Create the VMs, which takes around 1 to 2 minutes to complete.
	```
	terraform init
	terraform apply
	```

4. If SSH connections to external sites are blocked by your organization, get the proxy settings from your organization's IT department (proxy host and port), and configure SSH to use the proxy by editing the config file `~/.ssh/config` and adding the proxy information, as in this example --
	```
	# example host: proxy.XXX.com
	# example port: 1080
	Host *
	User ubuntu
	ProxyCommand nc -X 5 -x proxy.XXX.com:1080 %h %p
	```

5. Get the public IP addresses of the new VMs.
	```
	terraform show
	```

6. Choose one of the VMs to be the control-plane node and copy your SSH private key file to it. For example,
	```
	scp ~/.ssh/id_rsa ubuntu@public_IP_of_VM:.ssh/
	```

### Create the CloudXPRT Kubernetes environment

1. Copy the CloudXPRT release package to the control-plane VM.
	```
	scp ~/CloudXPRT_vXXXX_web-microservices.tar.gz ubuntu@public_IP_of_VM:
	```

2. SSH into the control-plane node.
	```
	ssh ubuntu@public_IP_of_VM
	```

3. Create the Kubernetes cluster.
	```
	tar -xzf ~/CloudXPRT_vX.XX_data-analytics.tar.gz
	cd ~/CloudXPRT_vX.XX_data-analytics/installation

	# modify the cluster_config.json file with the IP address shown with ifconfig
	ifconfig

	sudo ./prepare-cluster-CSP.sh
	sudo ./create-cluster.sh
	```

### Run the benchmark
```
cd ~
tar xzf ~/CloudXPRT_vXXXX_web-microservices.tar.gz
cd ~/CloudXPRT_vXXXX_web-microservices/cnbrun
```

Modify the `config.json` file according to `README.md` in the `cnbrun` directory.

Run the benchmark:
```
./cnbrun
```

**Note:** Results will be written to the `output` directory.

### Save results locally
Copy the results to your local machine.
```
scp ubuntu@public_IP_of_VM:CloudXPRT_vXXXX_web-microservices/cnbrun/output/* .
```

### Clean up the cluster

Run the following to clean up the cluster after you have finished running CloudXPRT and have saved the results:
```
terraform destroy
```

### References:
- [KOPS on AWS](https://www.bogotobogo.com/DevOps/DevOps-Kubernetes-II-kops-on-AWS.php)
- [Getting Started with KOPS on AWS](https://github.com/kubernetes/kops/blob/master/docs/getting_started/aws.md)
- [Managing access keys for IAM users](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html)
- [Installing AWS CLI on LInux](https://docs.aws.amazon.com/cli/latest/userguide/install-linux.html)
- [Amazon EC2 key pairs and Linux instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)
- [Creating IAM Users](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html)
- [Changing IAM permissions](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_change-permissions.html)

___

## Configure and run on Microsoft Azure

### Preparation

1. On a local Ubuntu Linux machine, create a new user and switch to it.
	```
	sudo adduser azureuser
	sudo adduser azureuser sudo
	```

2. If you are using a GUI on the local Ubuntu machine, log out and log back in as `azureuser`. Otherwise, you can change to the new user directly by using the following command:
	```
	su - azureuser
	```

3. Download the Terraform binary from the [Terraform site](https://www.terraform.io/downloads.html) and install it in the `/usr/local/bin` directory.

4. Install the Azure CLI.
	```
	curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
	```
	**Note:** If this step fails, refer to Microsoft Azure official documentation to install Azure CLI in an alternative way: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest

5. Authenticate with your Azure account.
	```
	az login
	az account list-locations
	```

6. Create SSH key pairs.
	```
	ssh-keygen
	```

### Create the virtual machines
1. Untar the CloudXPRT data-analytics code.
	```
	cd ~
	tar -xzf ~/CloudXPRT_vXXXX_web-microservices.tar.gz
	cd ~/CloudXPRT_vXXXX_web-microservices/terraform/azure
	```
2.  Modify the `variables.tf` file to set the following VM parameters for your needs.
	- location
	- vm_name
	- vm_size
	- storage_type


3. Create the VMs, which takes around 1 to 2 minutes to complete.
	```
	terraform init
	terraform apply
	```

4. If SSH connections to external sites are blocked by your organization, get the proxy settings from your organization's IT department (proxy host and port), and configure SSH to use the proxy by editing the config file `~/.ssh/config` and adding the proxy information, as in this example --
	```
	# example host: proxy.XXX.com
	# example port: 1080
	Host *
	User azureuser
	ProxyCommand nc -X 5 -x proxy.XXX.com:1080 %h %p
	```

5. Get the public IP addresses of the new VMs.
	```
	terraform show
	```

6. Choose one of the VMs to be the control-plane node and copy your SSH private key file to the node. For example,
	```
	scp ~/.ssh/id_rsa azureuser@public_IP_of_VM:.ssh
	```

### Create the CloudXPRT Kubernetes environment

1. Copy the CloudXPRT release package to the control-plane VM.
	```
	scp ~/CloudXPRT_vX.XX_data-analytics.tar.gz azureuser@public_IP_of_VM:
	```

2. SSH in to the control-plane node.
	```
	ssh azureuser@public_IP_of_VM
	```

3. Create the Kubernetes cluster.
	```
	tar -xzf ~/CloudXPRT_vX.XX_data-analytics.tar.gz
	cd ~/CloudXPRT_vX.XX_data-analytics/installation

	# modify the cluster_config.json file with the IP address shown with ifconfig
	ifconfig

	sudo ./prepare-cluster-CSP.sh
	sudo ./create-cluster.sh
	```

### Run the benchmark

1. Untar the CloudXPRT files.
	```
	cd ~
	tar xf ~/CloudXPRT_vXXXX_web-microservices.tar.gz
	cd ~/CloudXPRT_vXXXX_web-microservices/cnbrun
	```

2. Modify the `config.json` file according to the README in the `cnbrun` directory.

3. Run the benchmark.
	```
	./cnbrun
	```
**Note:** Results will be written to the `output` directory.

### Save results locally
Copy the results to your local machine.
```
scp azureuser@public_IP_of_VM:CloudXPRT_vXXXX_web-microservices/cnbrun/output/* .
```

### Clean up the cluster
Run the following to clean up the cluster after you have finished running CloudXPRT and have saved the results:
```
terraform destroy
```

___

## Configure and run on Google Cloud Platform (GCP)

### Install a Kubernetes cluster with KOPS on Google Cloud and run CloudXPRT

### Preparation

1. On a local Ubuntu Linux machine, create a new user, and switch to it.
	```
	sudo adduser gcpuser
	sudo adduser gcpuser sudo
	```

2. If you are using a GUI on the local Ubuntu machine, log out and log back in as `gcpuser`. Otherwise, you can change to the new user directly by using the following command:
	```
	su - gcpuser
	```

3. Download the Terraform binary from the [Terraform site](https://www.terraform.io/downloads.html) and install it in the `/usr/local/bin` directory.

4. Install Google Cloud SDK and other tools by using the Gooogle SDK [installer](https://cloud.google.com/sdk/docs/downloads-interactive).
	```
	curl https://sdk.cloud.google.com | bash
	exec -l $SHELL
	```

	**Note**: To install Google Cloud SDK in alternative ways, refer to Google's [documentation](https://cloud.google.com/sdk/install).

5. Log in to your Google Cloud account to make sure your account works, and configure the default credentials. Create a project to work with or use an existing project.
	```
	gcloud init
	gcloud compute zones list
	gcloud auth application-default login
	```

6. Create a service account and download the service account's private key to the location on your machine where KOPS will run the file `your-service-account-key.json`. For help, refer to Google documentation on [service accounts](https://cloud.google.com/docs/authentication/production).


7. Create SSH keys.
	```
	ssh-keygen
	```

### Create the virtual machines

1. Copy your service-acount private key to the Terraform directory.
	```
	cp your-service-account-key.json CloudXPRT_vX.XX_data-analytics/terraform/gcp
	```

2. Untar the CloudXPRT data-analytics code.
	```
	cd ~
	tar -xzf ~/CloudXPRT_vXXXX_web-microservices.tar.gz
	cp your-service-account-key.json CloudXPRT_vXXXX_web-microservices/terraform/gcp
	cd ~/CloudXPRT_vXXXX_web-microservices/terraform/gcp
	```

3. Modify the `variables.tf` file to set the following VM parameters for your needs.
	- location
	- region_name
	- project_name
	- cred_file      ---          e.g., `your-service-account-key.json`
	- vm_name
	- vm_size

4. Create the VMs, which takes around 1 to 2 minutes to complete.
	```
	terraform init
	terraform apply
	```

5. If SSH connections to external sites are blocked by your organization, get the proxy settings from your organization's IT (proxy host and port), and configure SSH to use the proxy by editing the config file `~/.ssh/config` and adding the proxy information, as in this example --
	```
	# example host: proxy.XXX.com
	# example port: 1080
	Host *
	User gcpuser
	ProxyCommand nc -X 5 -x proxy.XXX.com:1080 %h %p
	```

6. Get the public IP addresses of the new VMs.
	```
	terraform show
	```

7. Choose one of the VMs to be the control-plane node and copy your SSH private key file to the node. For example,
	```
	scp ~/.ssh/id_rsa gcpuser@public_IP_of_VM:.ssh/
	```

### Create the CloudXPRT Kubernetes environment

1. Copy the CloudXPRT release package to the control-plane VM.
	```
	scp ~/CloudXPRT_vX.XX__web-microservices.tar.gz gcpuser@public_IP_of_VM:
	```

2. SSH into the control-plane node.
	```
	ssh gcpuser@public_IP_of_VM
	```

3. Create the Kubernetes cluster.
	```
	tar -xzf ~/CloudXPRT_vX.XX__web-microservices.tar.gz
	cd ~/CloudXPRT_vX.XX__web-microservices/installation

	# modify the cluster_config.json file with the IP address shown with ifconfig
	ifconfig

	sudo ./prepare-cluster-CSP.sh
	sudo ./create-cluster.sh
	```

### Run the benchmark
```
cd ~/CloudXPRT_vXXXX_web-microservices/cnbrun
```

Modify the `config.json` file according to the README in the cnbrun directory.

Run the benchmark:
```
./cnbrun
```

**Note:** Results will be written to the `output` directory.

### Save results locally
Copy the results to your local machine.
```
scp gcpuser@External_IP:CloudXPRT_vXXXX_web-microservices/cnbrun/output/* .
```

### Clean up the cluster
Run the following to clean up the cluster after you have finished running CloudXPRT and have saved the results:
```
terraform destroy
```
### References for Google Cloud Platform:
- [Getting started with KOPS on GCP](https://github.com/kubernetes/kops/blob/master/docs/getting_started/gce.md)
- [Installing Google Cloud SDK](https://cloud.google.com/sdk/install)
- [Kubernetes on Google Cloud with KOPS](https://www.cloudtechnologyexperts.com/kubernetes-google-cloud-kops)

___

## Running in demo mode with a UI
These scripts make it easy to bring up all of the services that are used during a normal CloudXPRT run. The main difference is that only one replica of each service is deployed, and the services remain deployed until you want to remove them. This feature gives users time to interact with the web pages that the web server is serving.

### Deploy all services
```
./services.sh up
```

### Viewing web server pages

Once all of the services are up, the script will print out an IP address you can use to interact with the front end. Here is an example output from the script:
```
You may access the web server UI by visiting one of the following addresses in your web browser:
  http://10.233.47.78:8070 on any machine in the cluster, or
  http://192.168.0.11:30800 externally on any machine within the same network
```

The second address printed is the IP address of the control-plane node. If you have a mult-inode cluster, you can access the web service by visiting the IP address of either node, along with the same port number listed from the script.

To expose the web-service ClusterIP address and ports:
```
kubectl get service web-service
NAME          TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
web-service   NodePort   10.233.47.78   <none>        8070:30800/TCP   14m
```

Port 8070 on the ClusterIP address is only accessible from any nodes within Kubernetes cluster.

Externally, use the actual node's IP address along with the port listed within the 30000-32767 range.

### Remove all services
```
./services.sh down
```

___

## Build the benchmark from source
Instructions for building the benchmark from source on Ubuntu 18.04.

1. Download and install the Go compiler.
	```
	wget https://dl.google.com/go/go<version>.linux-amd64.tar.gz
	sudo tar -C /usr/local/ -xzf go<version>.linux-amd64.tar.gz
	echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.profile
	source $HOME/.profile
	```

2. Compile the Go code and create the release package.
	```
	cd ~/CloudXPRT-src/web-microservices
	sudo apt install pkg-config libssl-dev -y
	```

3. Create the release archive in directory `build` as file `CloudXPRT_vX.XX_web-microservices.tar.gz`.
	```
	make build
	```
