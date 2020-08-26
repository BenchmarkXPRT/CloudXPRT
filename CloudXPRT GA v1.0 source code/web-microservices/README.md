# CloudXPRT Web Microservices Workload
- [Introduction](#introduction)
- [Setup and Run On-prem](#setup-and-run-onprem)
- [Run benchmark on Prem](#run-benchmark)
- [Benchmark results](#benchmark-results)
- [Setup and Run on AWS](#setup-and-run-on-aws)
- [Setup and Run on Azure](#setup-and-run-on-azure)
- [Setup and Run on GCP](#setup-and-run-on-gcp)
- [Running in Demo mode with UI](#demo-with-ui)
- [Build benchmark from source](#build-benchmark)

## Introduction
In the web-tier microservices workload, a simulated user logs in to a web application that does three things: provides a selection of stock options, performs Monte-Carlo simulations with those stocks, and presents the user with options that may be of interest. The workload reports performance in transactions per second, which testers can use to directly compare IaaS stacks and to evaluate whether any given stack is capable of meeting service-level agreement (SLA) thresholds.

## Setup and Run Onprem

The installation scripts within the installation directory will install and create a Kubernetes cluster using Kubespray. They will help you:

- setup your environment to run CloudXPRT,
- get the IP addresses for all machines in your cluster,
- set up passwordless SSH,
- install Ansible/Kubespray requirements,
- create the cluster, and
- remove the cluster once you are done running CloudXPRT.

#### Terminology

- Node - A single machine or virtual machine
- Master Node - The node running the installation, this will become the Kubernertes master node.
- Worker Node - Each machine that will join the Kubernetes cluster.

#### Supported OS

- Ubuntu 18.04

#### Minimum Requirements

It is highly recommended to run this benchmark on high end servers. While running, the benchmark will scale to utilize all the cores available. However, for functional testing, your physical node or VM must have at least:

- 16 logical or virtual CPUs
- 8GB Ram
- 10GB Disk Space

#### Installation Steps

##### Setup Environment
1. In each machine in your cluster:

    - Set the root password (Note: **must be the same on each machine**)

      ```
      sudo passwd root
      ```

    - Log in as root

      ```
      su
      ```

    - Ensure openssh-server is installed

      ```
      apt-get install openssh-server -y
      ```

    - Allow Root login access

      - Edit /etc/ssh/sshd_config
      - Uncomment and modify the PermitRootLogin line to allow SSH login as root

          ```
          PermitRootLogin yes
          ```

    - Restart sshd 

      ```
      service sshd restart
      ```

##### Master Node
1. Edit cluster_config.json

  For each machine in the cluster, add its IPV4 address and optionally desired hostname. One machine per {..} section within the "nodes" list, starting with the master node.

  **Notes:** Although optional, each hostname must be unique and can only contain **lowercase alphanumeric characters**. If hostnames are not provided, Kubespray will rename each host as node1, node2, ..., nodeN. This means that the master node's hostname will be changed to 'node1'.

  If your machines are behind a proxy, make sure to set "set_proxy" to "yes" add the correct proxy settings for "http_proxy" and "https_proxy". Those proxy settings will be applied on all the nodes to ensure that they can communicate through the Kubernetes networking plugin. Furthermore, you must reboot the nodes in order for them to take effect since /etc/environment is modified. You have the option "reboot" to allow the prepare-cluster.sh script to reboot all the nodes automatically. By default, the reboot option in cluster_config.json is set to 'yes'. If you set it to 'no', please manually reboot your machines after running prepare-cluster.sh, otherwise the cluster creation in the step 3 will fail.

  Example configuration for a three node cluster:

  ```
  "nodes": [
        {
            "ip_address": "192.168.0.11",
            "hostname": "master"
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

2. In the master node, run "prepare-cluster.sh" script as root to perform preparation steps.
  ```
  sudo ./prepare-cluster.sh
  ```

3. In the master node, run the "create-cluster.sh" script as root
  ```
  sudo ./create-cluster.sh
  ```

  This process may take anywhere from 6 up to 20 minutes.

  **NOTE**: If you get an error with respect to docker-ce repository '**RETRYING: ensure docker-ce repository public key is installed ...**', double check that the proxies are set up correctly! You may repeat the "prepare-cluster.sh" script to set this again, or you may manually edit them in each node of your cluster.

  You should also double check that the date and time are the same on all of the nodes.

  For more reference on Kubespray and possible errors, please check out their GitHub repo: https://github.com/kubernetes-sigs/kubespray

#### Reset Docker and Kubernetes cluster

To remove cluster and docker installation on every node, run the "remove-cluster.sh" script in the master node after you finish all the tests on this machine.

```
sudo ./remove-cluster.sh
```

Answer 'y' or 'yes' to the prompt.

Note: This will not remove the proxy settings. If you want to run CloudXPRT again, you can run the "create-cluster.sh" script to re-create the Kubernetes cluster.

## Run Benchmark

#### Running CloudXPRT Web Microservices

#### Configure benchmark parameters

Open config.json file to set the parameters for CloudXPRT.

- **runoption**: Run only a specific microservice or all of them individually (not in multitenancy)
- **iterations**: Number of iterations to run for each microservice
- **hpamode**: If true, use Kubernetes HPA to scale pods as load increases. Otherwise, create max pods from the beginning
- **postprocess**: If true, run the postprocess binary at the end of the run
- **ppoutputfile**: Sef if want post process results to be saved in a file
- **autoloader.initialclients**: The initial number of clients the load generator will create
- **autoloader.clientstep**: The number of clients to increase for each load generator iteration
- **autoloader.lastclient**: Stops the load generator  when reaching this number of clients. If set to -1, it will continue to  run until cluster is saturated (i.e. CPU ~100%)
- **autoloader.SLA**: Service Level Agreement for 95 percentile latency constraint. If set to -1, there is no latency constraint
- **autoloader.timeinterval**: The amount of time to spend within each load generator iteration with the specified amount of clients
- **workload.version**: Docker image version to use
- **workload.cpurequests**: Amount of CPU cores requested to assign to each pod (integer only)

**Note**: cnbrun, gobench, autoloader, and all shell scripts need to have executable permissions

#### Start the Benchmark Run

##### Running the load generator within the System Under Test (SUT)

Once parameters in config.json are configured, run the cnbrun executable.

```
./cnbrun
```

If benchmark run is interrupted in the middle, to clean up the resources generated

```
./cleanups.sh
```

To collect system information for the cluster, run the following script, on the master node only.

```
sudo ./system_info.sh
```

##### Running the load generator outside of the SUT

**Requirement:** The machine running the load generator must be on the same network as the Kubernetes cluster.

Copy the following directories from the master node of the cluster to the machine you want to run the load generator on:

1. cnbrun directory
2. $HOME/.kube directory

On the load generator machine:

1. Move the .kube directory to the user's home directory $HOME/
2. Install kubectl
	```
	sudo apt-get install kubectl
	```
3. Rename mc.remote.sh to mc.sh
	```
	mv mc.remote.sh mc.sh
	```
4. Ensure that the autoloader, cnbrun, cnbweb, gobench, and mc.sh within the cnbrun directory have executable permissions.
	```
	chmod +x autoloader cnbrun cnbweb gobench mc.sh
	```
5. Once parameters in config.json are configured, run the cnbrun executable.
	```
	./cnbrun
	```

## Benchmark Results

After a run, results will be located on the same node that ran the load generator. If you setup a multiple node cluster, the other nodes within the cluster will not have any result files. The results will be written to the 'cnbrun/output' directory. These files do not get deleted or overwritten. They will accumulate in the output directory after each run.

For each run, you will have 4 files:

1. A log file with the results in a formatted table
2. A csv file with the results
3. A log file with all the stdout output during the run
4. A copy of the config file used for that run

#### Metrics

The results can be summarized using the following metrics:

1. Max successful requests per minute under a specified SLA
2. Total Max successful requests per minute (regardless of SLA)

Currently deriving the metric from the results is a manual process from the log file with the formatted table.

By default, the max Service Level Agreement (SLA) for 95 percentile latency is 3 seconds (specified in config.json's 'autoloader.SLA' parameter). The load generator will stop when either the system under test can no longer meet the specified SLA or when the SUT cannot beat the current maximum number of successful request within 5 retries.

Below are the condensed results from a run. You can choose different SLA's of interest from the logfile. For example, choosing SLA's as 1000ms, 2000ms, and 3000ms, we can compare the request rate that the SUT was able to consistently respond to within that time.

From the results, you can see that the system was able to handle:
- 604 requests per minute under 1000ms,
- 889 requests per minute under 2000ms, and
- 900 requests per minute under 3000ms

The max throughput within this run is 900 successful requests per minute.


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


## Setup and Run on AWS
### Install Kubernetes cluster with KOPS on AWS and run CloudXPRT


### References:
- https://medium.com/containermind/how-to-create-a-kubernetes-cluster-on-aws-in-few-minutes-89dda10354f4
- https://github.com/kubernetes/kops/blob/master/docs/getting_started/aws.md
- https://medium.com/@mcyasar/amazon-aws-kubernetes-kops-installation-7a205fe2d118
- https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
- https://docs.aws.amazon.com/cli/latest/userguide/install-linux.html
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html
- https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html
- https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_change-permissions.html

## Preparation:

#### On local Ubuntu linux machine, create a new user then switch to this user
```
sudo adduser awsuser
sudo adduser awsuser sudo
su - awsuser
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

#### The AWS IAM user needs to be granted with the following permissions:
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
[default]
aws_access_key_id = AKIXXXXXXXXXXXXX
aws_secret_access_key = sIrkzNOXxXXXXXXXXXXXxxxXXXX
```

#### Update .bashrc file with the following lines. If bucket already exists, change it to an unique name.
```
export bucket_name=your.bucket.name
export KOPS_CLUSTER_NAME=yourclusername.k8s.local
export KOPS_STATE_STORE=s3://${bucket_name}
```
**NOTE:** It is recommended to create a gossip-based cluster. The only requirement to trigger this is to have the cluster name ends with ".k8s.local".

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
KOPS will automatically create a VPC for your cluster to run within. 

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
Enter file in which the key is (/home/XX/.ssh/id_rsa): hit return to take the default value
put the content generated to file: ~/.ssh/id_rsa.pub
```

### Create Cluster

#### Example cluster configuration
The following command will create a Kubernetes cluster configuration consisting of:
- one master node,
- one worker node of instance type m5.4xlarge,
- in us-west-2a Availability Zone,
- with the given cluster name,
- using your SSH public key for authentication
```

kops create cluster --master-count=1 --node-count=1 --node-size=m5.4xlarge --zones=us-west-2a --name="${KOPS_CLUSTER_NAME}"
```
**NOTE:** You can edit these options to create the cluster with any configuration you'd like. For CloudXPRT to run, node size of at least 16 vCPUs is required. If no key is passed with the --ssh-public-key flage, kops will use the public key file ~/.ssh/id_rsa.pub by default.

#### Verify the default secret is created by KOPS for this cluster
```
kops describe secret
```

#### Edit cluster configuration
Workaround a known KOPS issue where the metrics server may not work properly: https://github.com/kubernetes/kops/pull/6201
```
kops edit cluster
kubelet:
    anonymousAuth: false
    authenticationTokenWebhook: true     <--- Add this line
```

#### Deploy the cluster
```
kops update cluster --name "${KOPS_CLUSTER_NAME}" --yes
```

#### Wait for some time (around 5-10 minutes) and validate cluster
```
kops validate cluster
kops get cluster --state "${KOPS_STATE_STORE}"
```

### Run CloudXPRT and Save Results

#### Access cluster

- Go to AWS console
- Go to EC2 running instances page
- Choose instance name "master-us-west-2a.masters.yourclusername.k8s.local"
- Click "Connect"
- Get the connection string under the Public DNS section, it will have the following format:

    **ec2-54-189-181-18.us-west-2.compute.amazonaws.com**

#### Edit config file under .ssh directory to bypass company proxy issues, an example of config file:
```
Host *
ProxyCommand nc -X 5 -x proxy.XXX.com:1080 %h %p
```
+**NOTE:** Only needed if ssh/scp is blocked by company proxy, get the proxy settings from your companies' IT.

#### Copy CloudXPRT release package to master node
```
scp CloudXPRT_vXXXX_web-microservices-AWS.tar.gz admin@ec2-54-189-181-18.us-west-2.compute.amazonaws.com:~/
```
**Note:** Make sure you use your own connection string!

#### SSH into master node
```
ssh admin@ec2-54-189-181-18.us-west-2.compute.amazonaws.com
```

#### Run CloudXPRT
```
tar xzf CloudXPRT_vXXXX_web-microservices-AWS.tar.gz
cd CloudXPRT_vXXXX_web-microservices/cnbrun
```

Modify config.json file according to README in cnbrun directory

Run the benchmark:
```
./cnbrun
```

**NOTE:** Results will be written in the 'output' directory.

#### Save results locally
Leave the SSH connection from the master node
```
exit
```
In your local machine, copy the results
```
scp admin@ec2-54-189-181-18.us-west-2.compute.amazonaws.com:~/CloudXPRT_vXXXX_web-microservices/cnbrun/output/* .
```

#### Clean up Cluster
After you are done running CloudXPRT and have saved the results:
```
kops delete cluster --name "${KOPS_CLUSTER_NAME}" --yes
```

## Setup and Run on Azure

### Preparation:

#### On local Ubuntu linux machine, create a new user then switch to this user
```
sudo adduser azureuser
sudo adduser azureuser sudo
```

If you are using GUI on the local Ubuntu machine, logout and log back in as azureuser. Otherwise, you can directly change over to the new user using the following command.
```
su - azureuser
```

#### Install Azure CLI
```
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
```
**Note:** If this step fails, refer to Microsoft Azure official documentation to install Azure CLI in an alternative way:
```
https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest
```

#### Authenticate with your Azure account
```
az login
az account list-locations
```

#### Create the resource group cnbrg
```
az group create -l yourlocation -n cnbrg
```

### Create a VM and AKS Cluster to Run CloudXPRT

####  Create a virtual machine (VM)
If it is your first time running the 'az vm create' command, use the following parameters to generate a key value pair to connect to your VM's
```
az vm create \
  --resource-group cnbrg \
  --name myVM \
  --image UbuntuLTS \
  --admin-username azureuser \
  --generate-ssh-keys
```
Otherwise, you can provide the path to the previously generated public SSH key
```
az vm create \
  --resource-group cnbrg \
  --name myVM \
  --image UbuntuLTS \
  --admin-username azureuser \
  --ssh-key-value ~/.ssh/id_rsa.pub
```
Once completed, you will get the public IP address in order to connect to the VM.
```
{
  "fqdns": "",
  "id": "/subscriptions/531a053f-XXXX-XXXX-8b6e-ff74ecc15b52/resourceGroups/cnbrg/providers/Microsoft.Compute/virtualMachines/myVM",
  "location": "westus2",
  "macAddress": "00-0D-3A-FD-42-AD",
  "powerState": "VM running",
  "privateIpAddress": "10.0.0.4",
  "publicIpAddress": "51.143.7.103",         <-- public IP address to access this VM
  "resourceGroup": "cnbrg",
  "zones": ""
}
```

#### Create a service principal
Make a note of your own appId and password. These values are used when you create the AKS cluster in the next step.
```
az ad sp create-for-rbac --skip-assignment --name myAKSClusterServicePrincipal
{
  "appId": "559513bd-0c19-4c1a-87cd-851a26afd5fc",        <----------
  "displayName": "myAKSClusterServicePrincipal",
  "name": "http://myAKSClusterServicePrincipal",
  "password": "e763725a-5eee-40e8-a466-dc88d980f415",     <----------
  "tenant": "72f988bf-86f1-41af-91ab-2d7cd011db48"

```

#### Create an AKS cluster
```
az aks create \
    --resource-group cnbrg \
    --name cnbCluster \
    --node-count 1 \
    --enable-addons monitoring \
    --node-vm-size Standard_XXXX_XX \
    --ssh-key-value ~/.ssh/id_rsa.pub \
    --service-principal 559513bd-0c19-4c1a-87cd-851a26afd5fc \
    --client-secret e763725a-5eee-40e8-a466-dc88d980f415
```
You will have to wait a while for cluster to be created.

**NOTE:** You can edit these options to create the cluster you'd like. Ensure you use the correct values for --resource-group, --service-principal, and --client-secret from the previous command

#### (Optional) SSH Proxy configuration

If you are in a proxy environment, create the config file in the ~/.ssh directory to bypass proxy issues. Below is an example of config file:
```
Host *
ProxyCommand nc -X 5 -x proxy.XXX.com:1080 %h %p
```
**NOTE:** This is only needed if SSH/SCP is blocked by a proxy server, get the proxy settings from your IT department.

#### Copy CloudXPRT release package to the VM created and SSH to it
Make sure you use the public IP address provided from the 'az vm create' command.
```
scp CloudXPRT_vXXXX_web-microservices-Azure.tar.gz azureuser@51.143.7.103:~
ssh azureuser@51.143.7.103
```

#### Install Azure CLI and kubectl on the VM. Authenticate with your Azure account.
```
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.16.3/bin/linux/amd64/kubectl
chmod +x kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
az login
```

### Run CloudXPRT and Save Results

#### On VM, get the credentials of AKS cluster. Verify that kubectl can access the AKS cluster from the VM.
```
az aks get-credentials --name cnbCluster -g cnbrg
kubectl get nodes
```

#### Untar the CloudXPRT files
```
tar xf CloudXPRT_vXXXX_web-microservices-Azure.tar.gz
cd CloudXPRT_vXXXX_web-microservices/cnbrun
```

#### Modify config.json file according to README in cnbrun directory

Run the benchmark:
```
./cnbrun
```
**NOTE:** Results will be written in the 'output' directory.

#### Save results locally
Leave the SSH connection from the VM machine
```
exit
```
In your local machine, copy the results
```
scp azureuser@51.143.7.103:~/CloudXPRT_vXXXX_web-microservices/cnbrun/output/* .
```

#### Clean up and delete the resource group cnbrg

```
az group delete -n cnbrg
```
## Setup and Run on GCP
### Install Kubernetes cluster with KOPS on Google Cloud and run CloudXPRT


### References:
- https://github.com/kubernetes/kops/blob/master/docs/getting_started/gce.md
- https://cloud.google.com/sdk/install
- https://cloud.google.com/storage/docs/gsutil
- https://www.cloudtechnologyexperts.com/kubernetes-google-cloud-kops/


### Preparation:

#### On local Ubuntu linux machine, create a new user then switch to this user
```
sudo adduser gcpuser
sudo adduser gcpuser sudo
```

If you are using GUI on the local Ubuntu machine, logout and log back in as gcpuser. Otherwise, you can directly change over to the new user using the following command.
```
su - gcpuser
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
https://cloud.google.com/sdk/docs/downloads-interactive
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
```

#### Install Google Cloud SDK in alternative ways, refer to Google official documentations:
```
https://cloud.google.com/sdk/install

```

#### Install gsutil, only needed if it is not installed in the previous steps, refer to Google official documentations:
```
https://cloud.google.com/storage/docs/gsutil_install#linux
```

#### Log on your Google cloud account, make sure your account works and configure default credentials. Need create a project to work on or use an existing one.
```
gcloud init
gcloud compute zones list
gcloud auth application-default login
```

#### Update .bashrc file with the following lines. If bucket already exists, change it to an unique name.
```
export bucket_name=your.bucket.name
export KOPS_CLUSTER_NAME=yourclustername.k8s.local
export KOPS_STATE_STORE=gs://${bucket_name}/
export KOPS_FEATURE_FLAGS=AlphaAllowGCE
export PROJECT="$(gcloud config get-value project)"
```
**NOTE:** It is recommended to create a gossip-based cluster. The only requirement to trigger this is to have the cluster name ends with ".k8s.local".

#### Source .bashrc
```
source .bashrc
```

#### Create bucket for KOPS to run using gsutil
```
gsutil mb "$KOPS_STATE_STORE"
gsutil ls
```

### Create Cluster

#### Example cluster configuration
```
kops create cluster --name="${KOPS_CLUSTER_NAME}" --node-count=1 --node-size=n1-standard-16 --master-count=1 --zones us-central1-a --image "ubuntu-os-cloud/ubuntu-1804-bionic-v20190617" --state "${KOPS_STATE_STORE}" --project="${PROJECT}"
```
**NOTE:** You can edit these options to create the cluster with any configuration you'd like. For CloudXPRT to run, node size of at least 16 vCPUs is required.

#### Edit cluster configuration
Workaround a known KOPS issue where the metrics server may not work properly: https://github.com/kubernetes/kops/pull/6201
```
kops edit cluster
kubelet:
    anonymousAuth: false
    authenticationTokenWebhook: true                    <--- for metrics server
```

#### Deploy the cluster
```
kops update cluster --name "${KOPS_CLUSTER_NAME}" --yes
```

#### Wait for some time (around 5-10 minutes) and validate cluster
```
kops validate cluster
kops get cluster --state "${KOPS_STATE_STORE}"
kops get instancegroup --state "${KOPS_STATE_STORE}/" --name "${KOPS_CLUSTER_NAME}"
```

### Run CloudXPRT and Save Results

#### Edit config file under .ssh directory to bypass company proxy issues, an example of config file:
```
Host *
ProxyCommand nc -X 5 -x proxy.XXX.com:1080 %h %p
```
**NOTE:** Only needed if ssh/scp is blocked by company proxy, get the proxy settings from your companies' IT.

#### Access master node of the k8s cluster created

- Go to Google cloud console
- Go to the running VM instances page
- Locate the instance name "master-us-central1-XXXXX"
- Click the drop down menu of "SSH" for this instance
- Get the connection string under "View gcloud command", it will have the following format:
```
gcloud beta compute --project "yourprojname" ssh --zone "us-central1-a" "master-us-central1--XXXXX"
```
- Copy and run this command. For the first time run this command, it will prompt you to enter the passphrase and generate the key files under .ssh directory. Those keys and the passphrase you choose will be used to access master node. Use name is the linux user when you run this command.
```
google_compute_engine
google_compute_engine.pub
```
Once successful, you will be SSH'ed into the master node. Close the connection.
```
exit
```

#### Copy CloudXPRT release package to master node, its IP address could be found on VM instances page, under "External IP".
```
scp -i "~/.ssh/google_compute_engine" CloudXPRT_vXXXX_web-microservices-GCP.tar.gz gcpuser@External_IP:~/
```
**Note:** Make sure you use your own master node IP address!

#### SSH into master node
```
ssh -i "~/.ssh/google_compute_engine" gcpuser@External_IP
```

### Run CloudXPRT
```
tar xzf CloudXPRT_vXXXX_web-microservices-GCP.tar.gz
cd CloudXPRT_vXXXX_web-microservices/cnbrun
```

Modify config.json file according to README in cnbrun directory

Run the benchmark:
```
./cnbrun
```

**NOTE:** Results will be written in the 'output' directory.

#### Save results locally
Leave the SSH connection from the master node
```
exit
```
In your local machine, copy the results
```
scp -i "~/.ssh/google_compute_engine" gcpuser@External_IP:~/CloudXPRT_vXXXX_web-microservices/cnbrun/output/* .
```

## Clean up Cluster
After you are done running CloudXPRT and have saved the results:
```
kops delete cluster --name "${KOPS_CLUSTER_NAME}" --yes
```

## Demo with UI
Instructions for running the benchmark in demo mode with UI

These scripts make it easy to bring up all of the services that are used during a normal CloudXPRT run. The main difference is that only one replica of each service is deployed and the services remain deployed until you want to remove them. It gives users time to interact with the web pages that the web server is serving.

### Deploy all services
```
sudo ./services.sh up
```

### Viewing Web Server Pages

Once all of the services are up, the script will print out possible address you can visit to interact with the front end. Example output from the script:

```
You may access the web server UI by visiting one of the following addresses in your web browser:
	http://10.233.47.78:8070 on any machine within the cluster, or
	http://192.168.0.11:30800 externally on any machine within the same network
```

The second address printed is the ip address of the master node. If you have a multi node cluster, you can access the web service by visiting the ip address of either node along with the same port number listed from the script.

To get the web-service ClusterIP address and ports exposed:

```
kubectl get service web-service
NAME          TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
web-service   NodePort   10.233.47.78   <none>        8070:30800/TCP   14m
```

The ClusterIP address and port 8070 is only accessible from any nodes within Kubernetes cluster.

Externally, use the actual node's ip address along with the port listed within the 30000-32767 range.

### Remove all services
```
sudo ./services.sh down
```

## Build Benchmark
Instructions for building the benchmark from source on Ubuntu 18.04

#### Download and install GO

```
wget https://dl.google.com/go/go<version>.linux-amd64.tar.gz
sudo tar -C /usr/local/ -xzf go<version>.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.profile
source $HOME/.profile
```

#### Compile GO binaries and create release packages

```
cd CloudXPRT-src/web-microservices
sudo apt install pkg-config libssl-dev -y
```

- Create the on-prem release archive in directory "build" as file CloudXPRT_vX.XX_web-microservices-Onprem.tar.gz

```
make build
```

- Or create the AWS release archive in directory "build" as file CloudXPRT_vX.XX_web-microservices-AWS.tar.gz

```
make -f Makefile.cloud buildaws
```

- Or create the Azure release archive in directory "build" as file CloudXPRT_vX.XX_web-microservices-Azure.tar.gz

```
make -f Makefile.cloud buildazure
```

- Or create the GCP release archive in directory "build" as file CloudXPRT_vX.XX_web-microservices-GCP.tar.gz

```
make -f Makefile.cloud buildgcp
```

