To run kafkacat::
kubectl -n namespace_being_used exec -it kafkacatPODName-- /bin/bash

once you get to kafkat pod:
To create test and write to it:
kafkacat -P -b ClusterIPofKafkaBootstrap:9092 -t test
To read from test:
kafkacat -C -b ClusterIPofKafkaBootstrap:9092 -t test


Example:

kubectl -n kafka exec -it kafkacat-7fccbf8899-jdkt9 -- /bin/bash

kafkacat -P -b 10.233.60.60:9092 -t test

kafkacat -C -b 10.233.60.60:9092 -t test


for i in `seq 1 10`; do echo “hello kafka world” | kafkacat -P -b 10.233.60.60:9092 -t test; done


