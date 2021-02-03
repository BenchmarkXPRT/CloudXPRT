# CloudXPRT Data Analytics Workload

- [Introduction](#introduction)
- [Set up the benchmark](#set-up-the-benchmark)
- [Run the benchmark](#run-the-benchmark)
- [Benchmark results](#benchmark-results)
- [Set up and run the benchmark on AWS](#set-up-and-run-the-benchmark-on-aws)
- [Set up and run the benchmark on GCP](#set-up-and-run-the-benchmark-on-gcp)
- [Set up and run the benchmark on Azure](#set-up-and-run-the-benchmark-on-azure)
- [Set up and run the benchmark on Azure multi-node](#set-up-and-run-the-benchmark-on-azure-multi-node)
- [Uninstall the benchmark](#uninstall-the-benchmark)
- [Build the benchmark from source](#build-the-benchmark-from-source)
- [Notes, Known Issues, FAQ](#known-issues)


## Introduction

The CloudXPRT - Data Analytics workload measures scalable data analysis of particle physics experiment data (HIGGS dataset) using XGBoost. XGBoost is a gradient-boosting framework that data scientists often use for ML-based regression and classification problems. The purpose of the workload in the context of CloudXPRT is to evaluate how well an on-prem or cloud hardware infrastructure enables XGBoost to speed and optimize model training. The workload reports latency and throughput rates. As with the web microservices workload, testers can use this workload’s metrics to compare IaaS stack performance and to evaluate whether any given stack is capable of meeting SLA thresholds.


## Set up the benchmark

These scripts will install and create a Kubernetes cluster using Kubespray. They will help you:

- setup your environment to run CloudXPRT,
- get the IP addresses for all machines in your cluster,
- set up passwordless SSH,
- install Ansible/Kubespray requirements,
- create the cluster, and
- remove the cluster once you are done running CloudXPRT.

### Terminology

- Node - A single machine or virtual machine
- Master Node - The node running the installation, this will become the Kubernetes master node.
- Worker Node - Each machine that will join the Kubernetes cluster.

### Supported OS

- Ubuntu 18.04

### Minimum Requirements

We highly recommended running this benchmark on high end servers. While running, the benchmark will scale to utilize all the cores available. However, for functional testing, your physical node or VM must have at least:

- 16 logical or virtual CPUs
- 8 GB Ram
- 20 GB Disk Space

### Installation Steps

#### Setup Environment
In each machine in your cluster:

Set the root password (Note: **must be the same on each machine**)

    ```
    sudo passwd root
    ```
Log in as root

    ```
    su
    ```
Ensure openssh-server is installed

    ```
    apt install -y openssh-server sshpass
    ```
Allow Root login access
Edit /etc/ssh/sshd_config

    ```
    apt install -y openssh-server sshpass
    nano /etc/ssh/sshd_config
        #Uncomment and modify the PermitRootLogin line
        PermitRootLogin yes
    ```
Restart sshd

    ```
    service sshd restart
    ```

##### Master Node
In the master node, run "prepare-cluster.sh" script as root to perform preparation steps.

    ```
    su
    cd installation/
    ./prepare-cluster.sh
    ```

    Follow the prompts to add the IP address for each node in your cluster. The script will ask if your machines are behind a proxy. If so, make sure to add the correct proxy settings for http_proxy and https_proxy when asked. This script will add those proxy settings for all the nodes to ensure that they can communicate through the Kubernetes networking plugin.

    Note: If you add proxy configuration, you must reboot the nodes in order for them to take effect (since /etc/environment is modified). You have the option to allow the script to reboot all the nodes automatically. If you select no, please manually reboot your machines, otherwise the cluster creation in the next step will fail.

In the master node, run the "create-cluster.sh" script as root

   **Warning**: Each node must have a different hostname. By defualt, Kubespray will rename each host as node1, node2, ..., nodeN. This means that the master node's hostname will be changed to 'node1'. If this is not desired, you may edit the 'kubespray/inventory/cnb-cluster/hosts.ini' file with the hostnames you want by replacing all instances of 'node1' with the master's hostname. If you have multiple nodes, edit the other entries as well.

    ```
    su
    ./create-cluster.sh
    ./cnb-analytics_OnPrem-MultiNode_setup.sh (If using a multi-node cluster)
    ./cnb-analytics_setup.sh
    ```
   Note: You may receive some errors from _tiller_ and _minio_ pods/services while resources are being created. This is a normal behavior.

   This process may take anywhere from 5 to 10 minutes.

   **NOTE**: If you get an error with respect to docker-ce repository '**RETRYING: ensure docker-ce repository public key is installed ...**', double check that the proxies are set up correctly! You may repeat the "prepare-cluster.sh" script to set this again, or you may manually edit them in each node of your cluster.

   You should also double check that the date and time are the same on all of the nodes.

   For more reference on Kubespray and possible errors, please check out their GitHub repo: https://github.com/kubernetes-sigs/kubespray


## Run the benchmark

Once you complete successful installation

   ```
   cd ../cnbrun/
   ```

#### Configure parameters for a test run
Open cnb-analytics_config.json file to set the parameters for CNB.

   ```
   $ nano CloudXPRT-analytics_config.json
        cpus_per_pod: Number of vCPUs per Pod. default 12
        numKAFKAmessages: Number of transactions to be delivered and executed. default 1
        loadgen_lambda: Inter-arrival time between transactions following Poisson distribution. default 0.33
   ```

#### Run CNB-analytics
Once parameters are configured, run the cnbrun executable.

   ```
   su
   ./cnbrun
   ./cnb-analytics_parse-all-results.sh
   ```
**NOTE:** use cnb-analytics_clear.sh to reset kubernetes in case you have an invalid run. then re-issue ./cnbrun

#### Deep dive analysis to determine best system configuration
A script is provided to create a swept analysis in order to find the best throughput under a particular SLA.

   ```
   su
   ./cnb-analytics_run-automated.sh
   ```

Make sure you set the desired parameters

   ```
   $ nano cnb-analytics_run-automated.sh
        Lambda: sets the desired Inter-arrival time for the Poisson distribution. default Lambda=(0.33 0.66 0.85 1)
        vCPU_per_POD: sets the desired swept for different number of vCPUs per pod. default vCPU_per_POD=(46 23 15 11)
   ```

In case of errors please clear the temp PODs using:

   ```
   su
   ./cnb-analytics_clear.sh
   ```

## Benchmark Results
A script is provided to create a table from output folders

   ```
   ./cnb-analytics_parse-all-results.sh
   ```
You can easily create a csv file using these command: ./cnb-analytics_parse-all-results.sh | sed -e 's/\s\+/,/g' > results.csv

Some of the metrics listed in the output are listed below:
- TotalTimeWithSetup: Includes creation and setup of all working Pods and execution of ExpectedKAFKAmessages
- TotalDuration: Includes creation and setup of all working Pods and execution of ExpectedKAFKAmessages
- NumberOfPods: number of working Pods
- vCPUsperPod: number of vCPUs used per Pod
- ExpectedKAFKAmessages: Expected number of Kafka messages to be process among Pods. By default every Pods will process a single Kafka message
- DeliveredKAFKAmessages: number of Kafka messages that were processed among Pods.
- Dataset: Dataset used
- Throughput_tnx/min: Throughput in transactions per minute
- 90th_Percentile: Tail latency for the 90th percentile


## Set up and run the benchmark on AWS
Preparing to setup benchmark on AWS.

#### On local Ubuntu linux machine, create a new user then switch to this user

    ```
    sudo adduser awsuser
    sudo adduser awsuser sudo
    ```
If you are using GUI on the local Ubuntu machine, logout and log back in as gcpuser. Otherwise, you can directly change over to the new user using the following command:

    ```
    su - awsuser
    cd ~/
    ```

#### Install KOPS version 1.16.0

    ```
    curl -LO https://github.com/kubernetes/kops/releases/download/v1.16.0/kops-linux-amd64
    chmod +x kops-linux-amd64
    sudo mv kops-linux-amd64 /usr/local/bin/kops
    kops version
    ```

#### Create AWS IAM user and change permissions for this user, refer to AWS official documentations:

    ```
    https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html
    https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_change-permissions.html
    ```
The AWS IAM user needs to be granted with the following permissions:
 - AmazonEC2FullAccess
 - AmazonRoute53FullAccess
 - AmazonS3FullAccess
 - IAMFullAccess
 - AmazonVPCFullAccess

#### Install AWS CLI, refer to the AWS official documentations:

    ```
    https://docs.aws.amazon.com/cli/latest/userguide/install-linux.html
    ```

#### Create access keys for IAM user, refer to the AWS official documentations

    ```
    https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
    ```

#### Configure AWS CLI using access keys created

    ```
    aws configure
        AWS Access Key ID [None]: AKIXXXXXXXXXXXXX
        AWS Secret Access Key [None]: sIrkzNOXxXXXXXXXXXXXxxxXXXX
        Default region name [None]: us-west-2
        Default output format [None]: json
    ```

#### Verify key and secret are stored into "~/.aws/credentials" file

    ```
    nano ~/.aws/credentials
        [default]
        aws_access_key_id = AKIXXXXXXXXXXXXX
        aws_secret_access_key = sIrkzNOXxXXXXXXXXXXXxxxXXXX
    ```

#### Update .bashrc file with the following lines

    ```
    nano ~/.bashrc
        export bucket_name=cnb-analytics-store
        export KOPS_CLUSTER_NAME=cnbanalytics.k8s.local
        export KOPS_STATE_STORE=s3://${bucket_name}
    ```

#### Source .bashrc

    ```
    source .bashrc
    ```

#### Create S3 bucket for KOPS to run

    ```
    aws s3 mb s3://${bucket_name} --region us-west-2
    aws s3api put-bucket-versioning --bucket ${bucket_name} --versioning-configuration Status=Enabled
    ```

#### Make sure you have one spare VPC available for KOPS to run
KOPS will automatically create a VPC for your cluster to run within. If you don't have a spare one available, you will not be able to create your cluster.

**Note:** Or if you prefer creating the cluster inside your existing VPC, add the following flag to "kops create cluster"

    ```
      --vpc ${VPCID:-"vpc-27b0b0a6e4sXXX"}
    ```

#### Create a key pair on AWS console, i.e. ('cnb_aws_key').
From EC2, under "Resources" choose "Key pairs" then click "Create key pair". Input your key pair name and select "pem" as file format. Download and save the cnb_aws_key.pem file generated.
**NOTE:** Keep the file in a safe place. Do not lose this file and do not share it!

#### Create a public key using AWS linux machines’ private key

    ```
    mkdir .ssh
    cp cnb_aws_key.pem ~/.ssh/id_rsa
    cd ~/.ssh
    sudo chown awsuser id_rsa
    chmod 400 id_rsa
    ssh-keygen -y
        #Enter file in which the key is (/home/XX/.ssh/id_rsa): hit return to take the default value
        #put the content generated to file: ~/.ssh/id_rsa.pub
    ```

#### Edit config file under .ssh directory to bypass company proxy issues, an example of config file:

    ```
    nano ~/.ssh/config
         Host *
         ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p
    ```
**NOTE:** Only needed if ssh/scp is blocked by company proxy, get the proxy settings from your companies' IT.


## Create cluster

#### Example cluster configuration
The following command will create a Kubernetes cluster configuration consisting of:
- one master node,
- one worker node of instance type m5.4xlarge,
- in us-west-2a Availability Zone,
- with the given cluster name,
- using your SSH public key for authentication

    ```
    kops create cluster --master-count=1 --node-count=1 --node-size=m5.4xlarge --zones=us-west-2a --name=${KOPS_CLUSTER_NAME}
    ```
**NOTE:** You can edit these options to create the cluster with any configuration you'd like. For CNB to run, node size of at least 16 vCPUs is required. If no key is passed with the --ssh-public-key flage, kops will use the public key file ~/.ssh/id_rsa.pub by default.


#### Verify the default secret is created by KOPS for this cluster

    ```
    kops describe secret
    ```
**NOTE:** When create cluster, kops use the public key file ~/.ssh/id_rsa.pub by default.


#### Metrics server needs to be enabled for workload to run
Current there is a bug: https://github.com/kubernetes/kops/pull/6201. The current work around:

    ```
    kops edit cluster
    kubelet:
        anonymousAuth: false
        authenticationTokenWebhook: true     #<--- Add this line
    ```

#### Deploy cluster

    ```
    kops update cluster --name ${KOPS_CLUSTER_NAME} --yes
    ```

#### Wait for some time (around 5-10 minutes) and validate cluster

    ```
    #kops validate cluster
    kops get cluster --state ${KOPS_STATE_STORE}
    kubectl get nodes
    kubectl cluster-info
    ```

### Run CloudXPRT and Save Results

#### Access cluster
- Go to AWS console - EC2 Dashboard
- Go to EC2 running instances page
- Choose instance name "master-us-west-2a.masters.cnb.analytics.k8s.local"
- Click "Connect"
- Get the connection string under the Public DNS section, it will have the following format:

    **ec2-34-212-31-28.us-west-2.compute.amazonaws.com**

#### Copy CloudXPRT release package and security key (cnb_aws_key.pem) to master node

    ```
    scp -i "~/.ssh/id_rsa" CNBv0.xx-analytics_xx.x.tar.gz admin@ec2-34-212-31-28.us-west-2.compute.amazonaws.com:~/
    scp cnb_aws_key.pem admin@ec2-34-212-31-28.us-west-2.compute.amazonaws.com:~/
    ```
**Note:** Make sure you use your own connection string!

#### SSH into master node and execute the consecutive commands in the master node:

    ```
    ssh -i "~/.ssh/id_rsa" admin@ec2-34-212-31-28.us-west-2.compute.amazonaws.com
    # -- you should receive the prompt from aws: admin@ip-172-20-39-73:~$ --
    echo "alias ll='ls -alF'" >> ~/.bashrc; source ~/.bashrc
    cp cnb_aws_key.pem ~/.ssh/id_rsa
    chmod 400 ~/.ssh/id_rsa
    ```

#### Install CloudXPRT

    ```
    tar -xvf CNBv0.xx-analytics_xx.x.tar.gz
    cd CNBv0.xx-analytics_xx.x/installation/
    sudo ./cnb-analytics_aws_setup.sh
    sudo ./cnb-analytics_setup.sh
    ```

### Run the benchmark:
Modify cnb-analytics_config.json file according to README in cnbrun directory

    ```
    cd cnbrun/
    nano cnb-analytics_config.json  <--- Edit as desired
    sudo ./cnbrun
    ```
**NOTE:** Results will be written in the 'output' directory.
**NOTE:** use cnb-analytics_clear.sh to reset kubernetes in case you have an invalid run. then re-issue ./cnbrun

### Save results locally
Leave the SSH connection from the master node. In your local machine, copy the results

    ```
    scp -r -i "~/.ssh/id_rsa" admin@ec2-34-212-31-28.us-west-2.compute.amazonaws.com:~/CNBv0.xx-analytics_xx/cnbrun/output .
    ```

## Clean up cluster
After you are done running CloudXPRT and have saved the results:

    ```
    kops delete  cluster --name ${KOPS_CLUSTER_NAME} --yes
    ```

### References:
- https://www.bogotobogo.com/DevOps/DevOps-Kubernetes-II-kops-on-AWS.php
- https://medium.com/containermind/how-to-create-a-kubernetes-cluster-on-aws-in-few-minutes-89dda10354f4
- https://github.com/kubernetes/kops/blob/master/docs/aws.md
- https://medium.com/@mcyasar/amazon-aws-kubernetes-kops-installation-7a205fe2d118
- https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
- https://docs.aws.amazon.com/cli/latest/userguide/install-linux.html
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html
- https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html
- https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_change-permissions.html


## Set up and run the benchmark on GCP
Preparing to setup and run benchmark on Google Cloud Platform

#### On local Ubuntu linux machine, create a new user then switch to this user

    ```
    sudo adduser gcpuser
    sudo adduser gcpuser sudo
    ```
If you are using GUI on the local Ubuntu machine, logout and log back in as gcpuser. Otherwise, you can directly change over to the new user using the following command:

    ```
    su - gcpuser
    cd ~/
    ```

#### Install KOPS version 1.16.0

    ```
    curl -LO https://github.com/kubernetes/kops/releases/download/v1.16.0/kops-linux-amd64
    chmod +x kops-linux-amd64
    sudo mv kops-linux-amd64 /usr/local/bin/kops
    kops version
    ```

#### Install Google Cloud SDK and other tools by using the Google SDK installer (recommended)

    ```
    # https://cloud.google.com/sdk/docs/downloads-interactive
    curl https://sdk.cloud.google.com | bash
    exec -l $SHELL

    Note: To Install Google Cloud SDK in alternative ways, refer to Google official documentations:
          https://cloud.google.com/sdk/install
          Install gsutil, only needed if it is not installed in the previous steps, refer to Google official documentations:
          https://cloud.google.com/storage/docs/gsutil_install#linux
    ```

#### Log on your Google cloud account, make sure your account works and configure default credentials. Need create a project to work on or use an existing one.

    ```
    gcloud init
    gcloud compute zones list
    gcloud auth application-default login
    ```

#### Update .bashrc file with the following lines:

    ```
    export bucket_name=your.bucket.name
    export KOPS_CLUSTER_NAME=cnb-a-cluster.k8s.local
    export KOPS_STATE_STORE=gs://cnbtest-clusters/
    export KOPS_FEATURE_FLAGS=AlphaAllowGCE
    PROJECT=`gcloud config get-value project`
    export NODE_SIZE=n1-standard-4
    export NODE_ZONE=us-central1-a
    export IMAGE="ubuntu-os-cloud/ubuntu-1804-bionic-v20190617"
    ```

#### Source .bashrc

    ```
    source .bashrc
    ```

#### Create bucket for KOPS to run using gsutil

    ```
    gsutil mb $KOPS_STATE_STORE
    gsutil ls
    ```

## Create Cluster

#### Example cluster configuration

    ```
    kops create cluster --name=${KOPS_CLUSTER_NAME} \
    --node-count=1 --node-size=${NODE_SIZE} \
    --master-count=1 --zones=${NODE_ZONE} \
    --image=${IMAGE} \
    --state ${KOPS_STATE_STORE} --project=${PROJECT}
    ```
**NOTE:** You can edit these options to create the cluster you'd like. For CloudXPRT to run, node size of 16 vCPUs are required.

#### Edit cluster before deploy it to work around some known issues:
- Metrics server may not work properly: https://github.com/kubernetes/kops/pull/6201

    ```
    kops edit cluster
    kubelet:
        anonymousAuth: false
        authenticationTokenWebhook: true                    <--- for metrics server
    ```

#### Deploy cluster

    ```
    kops update cluster --name ${KOPS_CLUSTER_NAME} --yes
    ```

#### Wait for some time (around 5-10 minutes) and validate cluster

    ```
    #kops validate cluster
    kops get cluster --state ${KOPS_STATE_STORE}
    kops get instancegroup --state ${KOPS_STATE_STORE}/ --name ${KOPS_CLUSTER_NAME}
    ```

## Run CloudXPRT and Save Results

#### Edit config file under .ssh directory to bypass company proxy issues, an example of config file:

    ```
    mkdir ~/.ssh
    nano ~/.ssh/config
        Host *
        ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p
    ```
**NOTE:** Only needed if ssh/scp is blocked by company proxy, get the proxy settings from your companies' IT.

#### Access master node of the kubernetes cluster

- Go to Google cloud console
- Go to the running VM instances page
- Locate the instance name "master-us-central1-XXXXX"
- Click the drop down menu of "SSH" for this instance
- Identify external ip. We will use it later for scp files into the master. **External_IP**
- Get the connection string under "View gcloud command", it will have the following format:

    ```
        gcloud beta compute --project "yourprojname" ssh --zone "us-central1-a" "master-us-central1--XXXXX"
    ```
- Copy and run this command. For the first time run this command, it will prompt you to enter the passphrase and generate the key files under .ssh directory.
- Those keys and the passphrase you choose will be used to access master node. Username is the linux user when you run this command.
- This command automatically log you into the master node.

    ```
    google_compute_engine
    google_compute_engine.pub
    ```

#### Copy CloudXPRT release package to master node, its IP address could be found on VM instances page, under "External IP".

    ```
    scp -i "~/.ssh/google_compute_engine" .ssh/google_compute_* CNB-Release-Package.tar.gz gcpuser@External_IP:~/
    ```
**Note:** Make sure you use your own user name and master node IP address!

#### SSH into master node
Open a new terminal to access the GCP system

    ```
    ssh -i "~/.ssh/google_compute_engine" username@External_IP
    ```

#### Install CloudXPRT
On GCP system:

    ```
    mv google_compute_* .ssh/
    tar xvfz CNB-Release-Package.tar.gz
    cd CNBv0.xx-analytics_xx.x/installation/
    sudo ./cnb-analytics_gcp_setup.sh
    sudo ./cnb-analytics_setup.sh
    ```

### Run the benchmark:
Modify config.json file according to README in cnbrun directory

    ```
    cd CNBv0.xx-analytics_xx.x/cnbrun
    nano cnb-analytics_config.json  <--- Edit as desired
    sudo ./cnbrun
    ```
**NOTE:** Results will be written in the 'output' directory.
**NOTE:** use cnb-analytics_clear.sh to reset kubernetes in case you have an invalid run. then re-issue ./cnbrun

### Save results locally
On your local system

    ```
    scp -i "~/.ssh/google_compute_engine" -r username@External_IP:~/cnb-XXX/cnbrun/output .
    ```

### Clean up Cluster
After you are done running CloudXPRT and have saved the results:

    ```
    kops delete  cluster --name ${KOPS_CLUSTER_NAME} --yes
    ```

### References:
- https://github.com/kubernetes/kops/blob/master/docs/getting_started/gce.md
- https://cloud.google.com/sdk/install
- https://cloud.google.com/storage/docs/gsutil
- https://www.cloudtechnologyexperts.com/kubernetes-google-cloud-kops/


## Set up and run the benchmark on Azure

###Preparing to setup benchmark on Azure:

On local Ubuntu linux machine, create a new user then switch to this user

    ```
    sudo adduser azureuser
    sudo adduser azureuser sudo
    ```
If you are using GUI on the local Ubuntu machine, logout and log back in as gcpuser. Otherwise, you can directly change over to the new user using the following command:

    ```
    su - azureuser
    cd ~/
    ```

#### Install Azure cli

    ```
    curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
    ```
note: if the above command does not work, please refer to Azure's manual install instructions
      https://docs.microsoft.com/en-us/cli/azure/install-azure-cli-apt?view=azure-cli-latest

#### Authenticate with your Azure account, make sure your account works

    ```
    az login
    az account list-locations
    note: To sign in, use a web browser to open the page https://microsoft.com/devicelogin and enter the code GIVEN to authenticate.
    ```

#### Create the resource group cnbrg

    ```
    az group create -l westus2 -n cnbrg-a
    ```

#### Creating VMs to run CloudXPRT
If it is your first time running the 'az vm create' command, use the following parameters to generate a key value pair to connect to your VM's

    ```
    az vm create \
      --resource-group cnbrg-a \
      --name cnb-a-VM \
      --image UbuntuLTS \
      --admin-username azureuser \
      --size Standard_D16S_v3 \
      --ssh-key-value ~/.ssh/id_rsa.pub --generate-ssh-keys
    ```

Once completed, you will get the public IP address in order to connect to the VM.

    ```
    {
      "fqdns": "",
      "id": "/subscriptions/531a053a-XXXX-XXXX-8a6a-aa74aaa15b52/resourceGroups/cnbrg-a/providers/Microsoft.Compute/virtualMachines/cnb-a-VM",
      "location": "westus2",
      "macAddress": "00-0A-3A-AA-00-AA",
      "powerState": "VM running",
      "privateIpAddress": "10.0.0.4",
      "publicIpAddress": "51.143.7.103",         #<-- public IP address to access this VM
      "resourceGroup": "cnbrg-a",
      "zones": ""
    }
    ```

#### SCP cnb package to the VM created and SSH to it
Make sure you use the public IP address provided from the 'az vm create' command

    ```
    scp -o "ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p" cnbrun.tar.gz azureuser@51.143.7.103:~/
    ssh -o "ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p" azureuser@51.143.7.103
    ```

#### Decompress cnbrun.tar.gz file
On Azure VM:

    ```
    tar -xvf CNBv0.xx-analytics_xx.x.tar.gz
    cd CNBv0.xx-analytics_xx.x/installation/
    ip a
    sudo ./prepare-cluster.sh
    sudo ./create-cluster.sh
    sudo ./cnb-analytics_setup.sh
    ```

### Run the benchmark:
Modify config.json file according to README in cnbrun directory

    ```
    cd ../cnbrun/
    nano cnb-analytics_config.json
    sudo ./cnbrun
    ```
**NOTE:** Results will be written in the 'output' directory.
**NOTE:** use cnb-analytics_clear.sh to reset kubernetes in case you have an invalid run. then re-issue ./cnbrun to run again.

### Save results locally
From you local system:

    ```
    scp -o "ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p" azureuser@51.143.7.103:~/cnbrun/output/* .
    ```

### Clean up and delete the resource group cnbrg
After you are done running CNB and have saved the results:

    ```
    az group delete -n cnbrg-a
    ```


## Set up and run the benchmark on Azure multi-node

###Preparing to setup benchmark on Azure:
On local Ubuntu linux machine, create a new user then switch to this user

    ```
    sudo adduser azureuser
    sudo adduser azureuser sudo
    ```
If you are using GUI on the local Ubuntu machine, logout and log back in as gcpuser. Otherwise, you can directly change over to the new user using the following command:

    ```
    su - azureuser
    cd ~/
    ```

#### Install Azure cli

    ```
    curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
    ```
note: if the above command does not work, please refer to Azure's manual install instructions
      https://docs.microsoft.com/en-us/cli/azure/install-azure-cli-apt?view=azure-cli-latest

#### Authenticate with your Azure account, make sure your account works

    ```
    az login
    az account list-locations
    ```
note: To sign in, use a web browser to open the page https://microsoft.com/devicelogin and enter the code GIVEN to authenticate.

#### Create the resource group cnbrg

    ```
    az group create -l westus2 -n cnbrg-a
    ```

#### Create as many VMs as desired on which to run CloudXPRT
For the first time run, replace last line of the command below with "--generate-ssh-keys".
Once completed, you will get the public IP addresses in order to connect to the VMs.

    ```
    az vm create \
      --resource-group cnbrg-a \
      --name cnb-a-VM1 \
      --image UbuntuLTS \
      --admin-username azureuser \
      --size Standard_D16S_v3 \
      --ssh-key-value ~/.ssh/id_rsa.pub
    {
      "fqdns": "",
      "id": "/subscriptions/531a053f-XXXX-XXXX-8b6e-ff74ecc15b52/resourceGroups/cnbrg/providers/Microsoft.Compute/virtualMachines/cnb-a-VM",
      "location": "westus2",
      "macAddress": "00-0D-3A-FD-42-AD",
      "powerState": "VM running",
      "privateIpAddress": "10.0.0.4",
      "publicIpAddress": "51.143.7.103",         #<-- public IP address to access this VM
      "resourceGroup": "cnbrg-a",
      "zones": ""
    }
    ```

Note: Invoque 'az vm create' command as many times as nodes are desired changing the VM --ame parameter. example:

    ```
    az vm create \
      --resource-group cnbrg-a \
      --name cnb-a-VM2 \
      --image UbuntuLTS \
      --admin-username azureuser \
      --size Standard_D16S_v3 \
      --ssh-key-value ~/.ssh/id_rsa.pub
    ```

#### SCP cnb package and id_rsa keys to the first VM created

    ```
    scp -o "ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p" cnbrun.tar.gz ~/.ssh/id_rsa azureuser@51.143.7.103:~/
    ```

#### SSH into every VM created and ensure connectivity
Login into every VM you created by opening a new a terminal and ssh into it, example:

Terminal1 cnb-a-VM1:

    ```
    ssh -o "ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p" azureuser@51.143.7.103
    ```
Terminal2 cnb-a-VM2:

    ```
    ssh -o "ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p" azureuser@51.143.7.104
    ```

Execute following commands on every VM

    ```
    sudo passwd root
    sudo passwd azureuser
    sudo iptables -P FORWARD ACCEPT
    cp -r /home/azureuser/.ssh /root/
    nano /etc/ssh/sshd_config
    #==>> ADD PermitRootLogin yes
    #==>> ADD PasswordAuthentication yes
    sudo service ssh restart
    ```

make sure you can ssh into the other VM

    ```
    ssh 10.0.0.4
    ssh 10.0.0.5
    ```

#### Decompress cnbrun.tar.gz file
on Master VM (cnb-a-VM1):

    ```
    ssh -o "ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p" azureuser@51.143.7.103
    tar -xvf CNBv0.xx-analytics_xx.x.tar.gz
    cd CNBv0.xx-analytics_xx.x/installation/
    ip a
    sudo ./prepare-cluster.sh
    nano kubespray/inventory/cnb-cluster/group_vars/k8s-cluster/k8s-cluster.yml
    #===> EDIT: “kube_network_plugin” CHANGE from “calico” to “weave”
    sudo ./create-cluster.sh
    sudo ./cnb-analytics_OnPrem-MultiNode_setup.sh
    sudo ./cnb-analytics_setup.sh
    ```

#### Run the benchmark:
Modify cnb-analytics_config.json file according to README in cnbrun directory

    ```
    cd ../cnbrun/
    nano cnb-analytics_config.json
    sudo ./cnbrun
    ```
**NOTE:** Results will be written in the 'output' directory.

### Save results locally
From you local system:

    ```
    scp -o "ProxyCommand nc -X 5 -x proxy-us.XXX.com:1080 %h %p" azureuser@51.143.7.103:~/cnbrun/output/* .
    ```

## Clean up and delete the resource group cnbrg
After you are done running CloudXPRT and have saved the results:

    ```
    az group delete -n cnbrg-a
    ```


# Uninstall the benchmark
Reset Docker and Kubernetes cluster

To remove cluster and docker installation on every node, run the "remove-cluster.sh" script in the master node after you finish all the tests on this machine.

    ```
    su
    ./cnb-analytics_cleanup.sh
    ./remove-cluster.sh
    ```
Note: Answer 'y' or 'yes' to the prompt.
Note: This will not remove the proxy settings. If you want to run CloudXPRT again, you can run the "create-cluster.sh" script to re-create the Kubernetes cluster.


## Build the benchmark from source

### Download cnb_ml repository

    ```
    git clone cloudXPRT-analytics.git
    ```

### Download and install GO

    ```
    su
    wget https://dl.google.com/go/go1.13.1.linux-amd64.tar.gz
    sudo tar -C /usr/local/ -xzf go1.13.1.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.profile
    source $HOME/.profile
    ```

### Compile GO binaries

    ```
    su
    cd cnb_ml
    sudo apt update
    sudo apt -y upgrade
    sudo apt install git curl openssh-server pkg-config libssl-dev docker.io -y
    sudo systemctl start docker
    sudo systemctl enable docker
    ./cnb-analytics_build.sh
    ```


## Known Issues

### Notes
- cnbrun, xgboost.sh, cnb-analytics_config.json must all be in the same directory
- xgbooost.sh need to have executable permissions

### FAQ

Q1. The benchmark is looping for long time while setting up any of the "pod/services"
- Please open a new console and use the script to visually verify there are no errors and/or pods in underfined status.

```
su
./cnb-analytics_status.sh
```
