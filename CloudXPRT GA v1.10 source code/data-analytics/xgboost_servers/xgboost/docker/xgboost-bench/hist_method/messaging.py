#!/usr/bin/python3
#===============================================================================
# Copyright 2020 BenchmarkXPRT Development Community
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#===============================================================================

import socket
import json
from kafka import KafkaProducer, KafkaConsumer, TopicPartition

#KAFKA_HOSTS = ["localhost:9092"]
KAFKA_HOSTS = ["broker:9092"]

class Producer():
    def __init__(self):
        super(Producer, self).__init__()
        self._client_id = socket.gethostname()
        self._producer = None

    def send(self, topic, message):
        if not self._producer:
            try:
                self._producer = KafkaProducer(bootstrap_servers=KAFKA_HOSTS,
                                               client_id=self._client_id,
                                               api_version=(0, 10), retries=1)
            except Exception as e:
                print(str(e))
                self._producer = None

        if self._producer:
            try:
                self._producer.send(topic, value=message.encode('utf-8'),partition=0)
                print("sending "+topic+": ")
                print(message)
                self.flush()
            except Exception as e:
                print(str(e))
        else:
            print("producer not available")

    def flush(self):
        if self._producer:
            self._producer.flush()

    def close(self):
        if self._producer:
            self.flush()
            self._producer.close()

class Consumer():
    def __init__(self, group=None):
        super(Consumer, self).__init__()
        self._client_id = socket.gethostname()
        self._group = group
        #self._client_id = "test_client_id"
        print(self._group)
        print(self._client_id)
 

    def messages(self, topic, timeout=None):
        c = KafkaConsumer(topic, bootstrap_servers=KAFKA_HOSTS,auto_offset_reset='earliest', enable_auto_commit=True, client_id=self._client_id,
                          group_id=self._group, api_version=(0, 10))
        topicsset = c.topics()
        print("from messaging.py: topics")
        print(topicsset)
        partitions = c.partitions_for_topic(topic)
        if not partitions:
            raise Exception("Topic "+topic+" does not exist")

        timeout1 = 100 if timeout is None else timeout
        while True:
            partitions = c.poll(timeout1)
            #print("partitions after poll:")
            #print(partitions)
            if partitions:
                print("yes partitions:")
                for p in partitions:
                    print(p.partition)
                    #print("partID:")
                    #print(partID)
                    for msg in partitions[p]:
                        print("msg.key:" + msg.key.decode('utf-8'))
                        #print(json.loads(msg.decode('utf-8')))
                        #yield json.loads(msg.value.decode('utf-8'))
                        yield msg
            if timeout is not None:
                yield ""

        c.close()

    def debug(self, topic, partID):
        c = KafkaConsumer(bootstrap_servers=KAFKA_HOSTS, client_id=self._client_id,
                          group_id=None, api_version=(0, 10))

        # assign/subscribe topic
        partitions = c.partitions_for_topic(topic)
        if not partitions:
            raise Exception("Topic "+topic+" not exist")
        c.assign([TopicPartition(topic, p) for p in partitions])

        # seek to beginning if needed
        #c.seek_to_beginning()
        print("seek to end")
        c.seek_to_end()
        # fetch messages
        while True:
            partitions = c.poll(100)
            if partitions:
                for p in partitions:
                    for msg in partitions[p]:
                        #yield msg.value.decode('utf-8')
                        print("msg.key:" + str(msg.key, 'utf-8'))
                        #print(json.loads(msg.decode('utf-8')))
                        #yield json.loads(msg.value.decode('utf-8'))
                        yield msg
            yield ""

        c.close()
