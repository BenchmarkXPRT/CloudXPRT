## Cloud Native Benchmark (CNB)- Minio Server installation
This document describes how to setup and use MINIO object stoarge

#### Notes
1. Supports one node cluster only
2. Creates a storage class called "minio-data-storage" - kubernetes.io/no-provisioner
3. Create a persistent volume of size 10Gi on the host system at /mnt/minio. To change the storage capacity edit the file single/pv.yaml
4. Creates a persistent volume claim - "minio-pv-claim" for 10Gi.(Corresponds to the PV of 10Gi that was created earlier)
5. Creates a MINIO deployment using a Minio image from Docker Hub -  minio/minio:RELEASE.2019-12-19T22-52-26Z. MINIO_ACCESS_KEY is "minio" and MINIO_SECRET_KEY is "minio123"


#### Supported OS

- Ubuntu 18.04.4

#### Installation Steps
1. Run installation/Minio/setup.sh. Setup uses insecure access ( that is not https)
2. The following IP adresses are displayed at the end of install:
   For example:
   minio_service_ip:10.233.54.122
   minioEndpoint:http://10.233.102.190:9000


#### Testing/checking install
1. Navigate to the service ip or endpoint displayed after install finishes - for example: 10.233.54.122:9000 or http://10.233.102.190:9000. The Minio Browser is displayed. Login using the access key and secret key. The bucket with the contnet should be available.

#### Example for using Minio content from a python script running in a pod in a cluster
    from minio import Minio
    from minio.error import ResponseError
    ## use minio service ip address

    print("Testing minio access\n")
    ## use minio service ip address
    minioClient = Minio('10.233.31.132:9000',
                access_key='minio',
                secret_key='minio123',
                 secure=False)

    try:
       print(minioClient.fget_object('cnb-ml-bucket', 'crowd_run_1080p50', '/var/www/archive/crowd_run_1080p50_minio.mp4'))
    except ResponseError as err:
       print("MINIO ERROR:" + err)

#### UnInstall Steps
1. Run installation/Minio/cleanup.sh
