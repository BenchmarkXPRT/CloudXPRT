First mkdir -p /mnt/disks/minio-data
Create pv, create storageclass, create deployment, create service.

Then create miniomc pod, like below:
root@node1:/home/t/minio/stateful# kubectl get pods
NAME                      READY   STATUS    RESTARTS   AGE
minio-0                   1/1     Running   0          21m
miniomc-556697f8b-hjbgf   1/1     Running   0          11m

Check the log of POD minio-0:
# kubectl logs minio-0
Endpoint:  http://10.233.90.180:9000  http://127.0.0.1:9000

Browser Access:
   http://10.233.90.180:9000  http://127.0.0.1:9000

IP posted by minio server:
/ #  mc config host add local http://10.233.90.180:9000 minio minio123
Added `local` successfully.
/ # mc mb local/lalala
Bucket created successfully `local/lalala`.

