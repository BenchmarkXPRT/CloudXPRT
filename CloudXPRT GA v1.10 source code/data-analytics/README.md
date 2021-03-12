# CloudXPRT Data Analytics Workload

- [Introduction](#introduction)
- [Configure and run on-premises](#configure-and-run-on-premises)
	- [Configure the benchmark environment](#configure-the-benchmark-environment)
	- [Run the benchmark on premises](#run-the-benchmark-on-premises)
- [Benchmark results](#benchmark-results)
- [Configure and run on Amazon Web Services](#configure-and-run-on-amazon-web-services)
	- [Create the virtual machines AWS](#create-the-virtual-machines-aws)
	- [Run the benchmark AWS](#run-the-benchmark-aws)
	- [Save results locally](#save-results-locally)
    - [Clean up the cluster](#clean-up-the-cluster)
- [Configure and run on Google Cloud Platform](#configure-and-run-on-google-cloud-platform)
	- [Create the virtual machines GCP](#create-the-virtual-machines-gcp)
	- [Run the benchmark GCP](#run-the-benchmark-gcp)
	- [Save results locally](#save-results-locally)
    - [Clean up the cluster](#clean-up-the-cluster)
- [Configure and run on Microsoft Azure](#configure-and-run-on-microsoft-azure)
	- [Create the virtual machines Azure](#create-the-virtual-machines-azure)
	- [Run the benchmark Azure](#run-the-benchmark-azure)
    - [Save results locally](#save-results-locally)
    - [Clean up the cluster](#clean-up-the-cluster)
- [Uninstall the benchmark](#uninstall-the-benchmark)
- [Build the benchmark from source](#build-the-benchmark-from-source)
- [Known issues](#known-issues)

## Introduction

The CloudXPRT data analytics workload uses the gradient-boosting technique to classify a moderately large dataset with the XGBoost library. XGBoost is a gradient-boosting framework that data scientists often use for ML-based regression and classification problems. In the context of CloudXPRT, the purpose of the workload is to evaluate how well an IaaS stack enables XGBoost to speed and optimize model training. To do this, the data analytics workload uses Kubernetes, Docker, object storage, message pipeline, and monitorization components to mimic an end-to-end IaaS scenario.

The workload reports latency (response time in seconds in the 95th percentile) and throughput (jobs per minute) rates. Testers can use this workload’s metrics to compare IaaS stack performance and to evaluate whether any given stack is capable of meeting service-level agreement (SLA) thresholds.

### Terminology

- Node: A single machine or virtual machine
- Control-plane node: The node running the installation, which becomes the Kubernetes control-plane node.
- Worker node: Each machine that will join the Kubernetes cluster.

___

## Configure and run on-premises

### Configure the benchmark environment
The instructions below will install and create a Kubernetes cluster using Kubespray. They will help you:

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
- 50 GB of available disk space

### Installation

#### Set up the environment
For each machine in your cluster, perform steps 1 through 5.

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
	cd ~/CloudXPRT_vX.XX_data-analytics/installation
	```

2. Edit `cluster_config.json`.

	For each machine in the cluster, add its IPV4 address (required) and desired hostname (optional). Only one machine per `{..}` section can be in the `nodes` list. The control-plane node must be first.

	**Notes:** Although optional, each hostname must be unique and can only contain **lowercase alphanumeric characters**. If hostnames are not provided, Kubespray will rename each host as `node1`, `node2`, ..., `nodeN`. This means that the control-plane node's hostname will be changed to `node1`.

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

3. On the control-plane node, run the `prepare-cluster.sh` script to perform preparation steps.
	```
	sudo ./prepare-cluster.sh
	```

4. On the control-plane node, run the `create-cluster.sh` script to create the cluster.
	```
	sudo ./create-cluster.sh
	```

	This process can take anywhere from 6 to 20 minutes.

	**Notes**: If you get this error from the docker-ce repository '**RETRYING: ensure docker-ce repository public key is installed ...**'
	- Make sure that the proxies are set up correctly.
	- You may rerun the `prepare-cluster.sh` script, or you may manually set the proxies in each node.
	- You should also double check that the date and time are the same on all of the nodes.

	For more information on Kubespray and errors that occur during cluster creation, see its [documentation](https://github.com/kubernetes-sigs/kubespray).

### Set up the data analytics benchmark
This process may take anywhere from 5 to 10 minutes.
```
cd ~/CloudXPRT_vX.XX_data-analytics/setup

# Execute this command only if using a multinode cluster
./cnb-analytics_OnPrem-MultiNode_setup.sh

# Execute this command for all cluster types
sudo ./cnb-analytics_setup.sh
```

### Run the benchmark on premises

#### Configure parameters for a test run
CloudXPRT provides users with multiple configuration options for the data analytics workload. We recommend that you run the workload once with the default parameters to ensure correct functionallity.

Once you complete the installation, navigate to the run directory.
	```
	cd ~/CloudXPRT_vX.XX_data-analytics/cnbrun/
	```

Edit the `cnb-analytics_config.json` file to set the parameters for CloudXPRT data-analytics.
- `cpus_per_pod`: This setting lets the user designate the number of vCPUs per pod. The default setting is 12.
- `numKAFKAmessages`: This setting lets the user select the number of messages (transactions, jobs) that the workload will deliver and execute during the simulation. Each transaction is a request to the main module to use the gradient-boosting technique to create a classification model and perform two inferences with the resulting data. The default setting for testing is 1. In most cases, users should use at least 100 messages to achieve a statistically sound result.
- `loadgen_lambda`: This setting lets the user configure the interarrival time between transactions following the Poisson distribution. The default setting is 12 seconds.

#### Run CloudXPRT-analytics
Once the parameters are configured, run the cnbrun executable and examine the results.
	```
	sudo ./cnbrun
	sudo ./cnb-analytics_parse-all-results.sh
	```
**Note:** use `cnb-analytics_clear.sh` to reset Kubernetes in case you have an invalid run.

#### Deep dive analysis to determine best cluster configuration for best throughput
CloudXPRT includes a script that performs a swept analysis to find the best throughput under a particular SLA.
	```
	sudo ./cnb-analytics_run-automated.sh
	```

Make sure you set the desired parameters
- `Lambda`: Sets the desired interarrival time for the Poisson distribution. The default Lambda=(0.33 0.66 0.85 1)
- `vCPU_per_POD`: Directs the analysis to account for differing numbers of vCPUs per pod. The default vCPU_per_POD=(46 23 15 11)

In case of errors, please clear the temp PODs using:
	```
	sudo ./cnb-analytics_clear.sh
	```

___

### Benchmark results
Run the following script to create a CSV table of results in the output folders.
	```
	./cnb-analytics_parse-all-results.sh
	```
The main metrics in the results table are --
- `NumberOfPods`: This output reports the number of working pods that executed the XGBoost training task.
- `vCPUsperPod`: This output reports the number of vCPUs used per pod (while executing XGBoost training).
- `DeliveredKAFKAmessages`: This output reports the number of Kafka messages that were processed among the pods.
- `90th_Percentile`: This output reports the 90th percentile value for all recorded transaction times.
- `Throughput_jobs/min`: This output reports recorded throughput as expressed in transactions per minute.

The user has the freedom to define a throughput that complies with 90th_Percentile latency.

The complete format of the results table is as follows --
- `FILE`: The location of output_result.txt file containing the results from a particular simulation.
- `NumberOfPods`: The number of working pods that executed the XGBoost task.
- `vCPUsperPod`: The number of vCPUs used per Pod (while executing XGBoost training).
- `number of Jobs`: The total number of jobs executed during the simulation.
- `Lambda`: The interarrival time in seconds between transactions. The lower the lambda, the more traffic is sent to the CUT.
- `jobs_duration`: This output reports the elapsed time between the creation of the first transaction
and the arrival of the last transaction at the load generator.
- `min_duration`: The minimum recorded transaction time.
- `max_duration`: The maximum recorded transaction time.
- `stdev_duration`: The standard deviation of recorded transaction times.
- `mean_duration`: The mean of recorded transaction times.
- `90th_Percentile`: 90th percentile value for all recorded transaction times.
- `95th_Percentile`: 95th percentile value for all recorded transaction times.
- `Throughput_jobs/min`: Throughput as expressed in transactions per minute.

#### Reset the Docker environment and the Kubernetes cluster

To remove the Kubernetes cluster and Docker installation from every node, run the `remove-cluster.sh` script on the control-plane node after you finish your runs.
	```
	sudo ./remove-cluster.sh
	```
	Answer 'y' or 'yes' to the prompt.

**Note**: This will not remove the proxy settings. If you want to run CloudXPRT again, you can run the `create-cluster.sh` script to re-create the Kubernetes cluster.
___

## Configure and run on Amazon Web Services
### Preparation

1. On a local Ubuntu Linux machine, create a new user, and switch to it.
	```
	sudo adduser awsuser
	sudo adduser awsuser sudo
	```

2. If you are using a GUI on the local Ubuntu machine, log out and log back in as `awsuser`. Otherwise, you can change to the new user by using the following command:
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

	For help, refer to the AWS documentation for [access keys](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html).

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

### Create the virtual machines AWS

1. Extract the CloudXPRT data-analytics coda from the archive.
	```
	cd ~
	tar -xzf ~/CloudXPRT_vX.XX_data-analytics.tar.gz
	cd ~/CloudXPRT_vX.XX_data-analytics/terraform/aws
	```

2. Modify the `variables.tf` file to set the following VM parameters for your needs.
	- aws_region
	- aws_zones
	- instance_ami
	- instance_type
	- vm_name

3. Create the VMs, which takes around 1-2 minutes to complete.
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
	scp ~/CloudXPRT_vX.XX_data-analytics.tar.gz ubuntu@public_IP_of_VM:
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

4. Set up CloudXPRT data-analytics.
	```
	cd ~/CloudXPRT_vX.XX_data-analytics/setup

	# Execute this command if using a multinode cluster
	sudo ./cnb-analytics_OnPrem-MultiNode_setup.sh

	# Execute this command for all cluster types
	sudo ./cnb-analytics_setup.sh
	```

### Run the benchmark AWS
CloudXPRT includes a script that performs a swept analysis to find the best throughput under a particular SLA
	```
	cd ~/CloudXPRT_vX.XX_data-analytics/cnbrun
	sudo ./cnb-analytics_run-automated.sh
	```
**Notes:**
- Results will be written to the `output` directory.
- Use `cnb-analytics_clear.sh` to reset Kubernetes in case you have an invalid run.

### Save results locally AWS
On your local machine, copy the results from the control-plane node to your system.
	```
	scp ubuntu@public_IP_of_VM:CloudXPRT_vX.XX_data-analytics/cnbrun/output_* .
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

## Configure and run on Google Cloud Platform

### Preparing to set up and run the benchmark on GCP

1. On a local Ubuntu Linux machine, create a new user, and switch to it.
	```
	sudo adduser gcpuser
	sudo adduser gcpuser sudo
	```
2. If you are using a GUI on the local Ubuntu machine, log out and log back in as `gcpuser`. Otherwise, you can change to the new user by using the following command:
	```
	su - gcpuser
	cd ~/
	```

3. Download the Terraform binary from the [Terraform site](https://www.terraform.io/downloads.html) and install it in the `/usr/local/bin` directory.

4. Install Google Cloud SDK and other tools by using the Google SDK [installer](https://cloud.google.com/sdk/docs/downloads-interactive).
	```
	curl https://sdk.cloud.google.com | bash
	exec -l $SHELL
	```
	**Note**: To Install Google Cloud SDK in alternative ways, refer to Google's [documentation](https://cloud.google.com/sdk/install).

5. Log in to your Google Cloud account, make sure your account works, and configure default credentials. Create a project to work on or use an existing project.
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

### Create the virtual machines GCP

1. Copy your service-acount private key to the Terraform directory.
	```
	cp your-service-account-key.json CloudXPRT_vX.XX_data-analytics/terraform/gcp
	```

2. Extract the CloudXPRT data-analytics coda from the archive.
	```
	cd ~
	tar -xzf ~/CloudXPRT_vX.XX_data-analytics.tar.gz
	cd ~/CloudXPRT_vX.XX_data-analytics/terraform/gcp
	```

2. Modify the `variables.tf` file to set the following VM parameters for your needs.
	- location
	- region_name
	- project_name
	- cred_file      ---    	 e.g., `your-service-account-key.json`
	- vm_name
	- vm_size

3. Create the VMs, which takes around 1-2 minutes to complete.
	```
	terraform init
	terraform apply
	```

4. If SSH connections to external sites are blocked by your organization, get the proxy settings from your organization's IT department (proxy host and port), and configure SSH to use the proxy by editing the config file `~/.ssh/config` and adding the proxy information, as in this example --
	```
	# example host: proxy.XXX.com
	# example port: 1080
	Host *
	User gcpuser
	ProxyCommand nc -X 5 -x proxy.XXX.com:1080 %h %p
	```
5. Get the public IP addresses of the new VMs.
	```
	terraform show
	```

6. Choose one of the VMs to be the control-plane node and copy your SSH private key file to it. For example,
	```
	scp .ssh/id_rsa gcpuser@public_IP_of_VM:.ssh/
	```

### Create the CloudXPRT Kubernetes environment

1. Copy the CloudXPRT release package to the control-plane VM.
	```
	scp ~/CloudXPRT_vX.XX_data-analytics.tar.gz gcpuser@public_IP_of_VM:
	```

2. SSH into the control-plane node.
	```
	ssh gcpuser@public_IP_of_VM
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

4. Set up CloudXPRT data-analytics.
	```
	cd ~/CloudXPRT_vX.XX_data-analytics/setup

	# Execute this command if using a multinode cluster
	sudo ./cnb-analytics_OnPrem-MultiNode_setup.sh

	# Execute this command for all cluster types
	sudo ./cnb-analytics_setup.sh
	```

### Run the benchmark GCP
CloudXPRT includes a script that performs a swept analysis to find the best throughput under a particular SLA.
	```
	cd ~/CloudXPRT_vX.XX_data-analytics/cnbrun
	sudo ./cnb-analytics_run-automated.sh
	```
**Notes:**
- Results will be written in the `output` directory.
- Use `cnb-analytics_clear.sh` to reset Kubernetes in case you have an invalid run.

### Save results locally
Copy the results from the control-plane node to your local machine.
	```
	scp gcpuser@public_IP_of_VM:CloudXPRT_vX.XX_data-analytics/cnbrun/output_* .
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

## Configure and run on Microsoft Azure

1. On a local Ubuntu Linux machine, create a new user and switch to it.
	```
	sudo adduser azureuser
	sudo adduser azureuser sudo
	```

2. If you are using a GUI on the local Ubuntu machine, log out and log back in as `azureuser`. Otherwise, you can change to the new user using the following command:
	```
	su - azureuser
	```

3. Download the Terraform binary from the [Terraform site](https://www.terraform.io/downloads.html) and install it in the `/usr/local/bin` directory.

4. Install the Azure CLI.
	```
	curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
	```

	**Note:** If this step fails, refer to Microsoft Azure documentation on how to install Azure CLI in an [alternative way](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest).

5. Authenticate with your Azure account.
	```
	az login
	az account list-locations
	```
6. Create SSH keys.
	```
	ssh-keygen
	```

### Create the virtual machines Azure

1. Extract the CloudXPRT data-analytics coda from the archive.
	```
	cd ~
	tar -xzf ~/CloudXPRT_vX.XX_data-analytics.tar.gz
	cd ~/CloudXPRT_vX.XX_data-analytics/terraform/azure
	```

2. Modify the `variables.tf` file to set the following VM parameters for your needs.
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

6. Choose one of the VMs to be the control-plane node and copy your SSH private key file to it. For example,
	```
	scp ~/.ssh/id_rsa azureuser@public_IP_of_VM:.ssh/
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

### Configure the data-analytics benchmark

This process can take anywhere from 5 to 10 minutes.
```
cd ~/CloudXPRT_vX.XX_data-analytics/setup/

# Execute this command if using a multinode cluster
sudo ./cnb-analytics_OnPrem-MultiNode_setup.sh

# Execute this command for all cluster types
sudo ./cnb-analytics_setup.sh
```

### Run the benchmark Azure
CloudXPRT includes a script that performs a swept analysis to find the best throughput under a particular SLA.
```
cd ~/CloudXPRT_vX.XX_data-analytics/cnbrun
sudo ./cnb-analytics_run-automated.sh
```
**Notes:**
- Results will be written in the `output` directory.
- Use `cnb-analytics_clear.sh` to reset Kubernetes in case you have an invalid run.

### Save results locally
Copy the results from the control-plane node to your local machine.
```
scp azureuser@public_IP_of_VM:CloudXPRT_vX.XX_data-analytics/cnbrun/output_* .
```

### Clean up the cluster
Run the following to clean up the cluster after you have finished running CloudXPRT and have saved the results:
```
terraform destroy
```
___

## Uninstall the benchmark
After you finish your runs, the following steps will uninstall Kubernetes and Docker. If you want to run CloudXPRT again, you can run the `create-cluster.sh` script to re-create the Kubernetes cluster.

**Note**: These steps will not remove the proxy settings.

1. Uninstall the components used in Cloud Analytics.
	```
	cd ~/CloudXPRT_vX.XX_data-analytics/setup
	sudo ./cnb-analytics_cleanup.sh
	```

2. Run the `remove-cluster.sh` script on the control-plane node to remove the cluster and Docker installations on **every** node. Answer `y` or `yes` at the prompt.
	```
	cd ~/CloudXPRT_vX.XX_data-analytics/setup
	sudo ./remove-cluster.sh
	```

___

## Build the benchmark from source
Instructions for building the benchmark from source on Ubuntu 18.04

1. Download and install the Go compiler.
	```
	wget https://dl.google.com/go/go<version>.linux-amd64.tar.gz
	sudo tar -C /usr/local/ -xzf go<version>.linux-amd64.tar.gz
	echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.profile
	source $HOME/.profile
	```

2. Compile the Go code and create the release package in the `build` directory as file `CloudXPRT_vX.XX_data-analytics.tar.gz`.
	```
	cd ~/CloudXPRT-src/data-analytics
	make build
	ls build/
	```

___

## Known issues

### FAQ

Q1. The benchmark is looping for long time while setting up any of the "pod/services".

Please open a new console and use the script to verify that there are no errors, and no pods are in an underfined status.
```
sudo ./cnb-analytics_status.sh
```

Q2. How do I clean the current execution and start again?

1. Identify any running scripts with the following command, and then kill them.
	```
	ps -aux | grep cnbrun | grep automated
	```
2. Remove all pods associated with the workload.
	```
	sudo ./cnb-analytics_clear.sh
	```
3. Reset the `kafka` subsystem.
	```
	cd ../setup/setup_kafka
	sudo ./cleanup.sh
	sudo ./setup.sh
	```
4. Restart the run.
	```
	cd ../../cnbrun
	sudo ./cnbrun
	```
