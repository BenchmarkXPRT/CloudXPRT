#### Setup
To setup Kafka environment, run the setup script.

```
./setup.sh
```

To test the Kafka environment, you need to know the cluster-ip address of the bootstrap service.
```
kubectl get service -n kafka
```

Open up two terminals. In each terminal run the test script. These scripts will automatically log you into a kafkacat pod.

```
./test.sh
```

In one terminal, setup a subscriber
```
kafkacat -C -b broker:9092 -t test
```

In the other terminal, publish some messages
```
for i in `seq 1 10`; do echo “hello kafka world” | kafkacat -P -b broker:9092 -t test; done
```

#### Cleanup
To cleanup Kafka environment, run the cleanup script.

```
./cleanup.sh
```

