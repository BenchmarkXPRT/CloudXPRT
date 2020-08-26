#!/bin/bash
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

# Deploy kafkacat pod
kubectl apply -f kafkacat/kafkacat.yml

# Wait until kafkacat pod is available
while [ ! "$(kubectl get pods -n kafka| grep "kafkacat" | awk '{print $2}')" = "1/1" ]
do
    sleep 2;
done

# Get into kakfacat pod shell
kafkacat_pod=$(kubectl get pod -n kafka | grep "kafkacat" | awk '{print $1}')
kubectl exec -it $kafkacat_pod -n kafka /bin/bash
