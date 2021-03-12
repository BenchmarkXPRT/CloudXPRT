#!/usr/bin/env python3
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

import sys, time
from kubernetes import client, config


def main():
    # it works only if this script is run by K8s as a POD
    try:
        config.load_incluster_config()
    except:
        config.load_kube_config()

    functionToRun = sys.argv[1]
    numberOfPodsTocreate = int(sys.argv[2])
    numberOfCpusPerPod = sys.argv[3]
    minio_service_ip = sys.argv[4]

    namespace="kafka"
    appname= "xgb"
    #sample use of the API
    apiv1 = client.CoreV1Api()
    print("Listing pods with their IPs:")
    ret = apiv1.list_pod_for_all_namespaces(watch=False)
    for i in ret.items:
        print("%s\t%s\t%s" %
               (i.status.pod_ip, i.metadata.namespace, i.metadata.name))
        sys.stdout.flush()

    podname= appname
    if functionToRun == "create":
      for x in range(numberOfPodsTocreate):
        ret = create_pod("cloudxprt/xgboost:v1.00",podname+str(x),namespace,appname,x,numberOfCpusPerPod,minio_service_ip,"")
        if (ret==0):
            print("Error Creating POD")
            podCreationStatus = False
            print("error")
            sys.exit("error")


      while True:
            r = apiv1.list_namespaced_pod(namespace, label_selector="app=xgb")
            numPods = len(r.items)
            print("numPods:" + str(numPods))
            numPodsRunning = 0
            print("numPodsRunning:" + str(numPodsRunning))
            for p in r.items:
               if p.status.phase == "Running":
                 numPodsRunning = numPodsRunning + 1
            if(numPodsRunning == numPods):
                   print("All Running")
                   break
            else:
              print("Waiting to have all pods running...")
              time.sleep(1)
              numPodsRunning = 0
    elif functionToRun == "delete":
      delete_pod(appname,namespace)
    else:
      print("No function/action selected")
      sys.stdout.flush()
      sys.exit(0)

    sys.exit(0)

def create_pod(image, name, namespace, labelstr, pod_id, num_cpus_per_pod, minio_service_ip, service_account):
    print("Creating pod with image {} in {}".format(image, namespace))
    print("POD_ID:"+ str(pod_id))
    try:
        try:
            config.load_incluster_config()
        except:
            config.load_kube_config()
        apiv1 = client.CoreV1Api()
        pod = apiv1.create_namespaced_pod(
            namespace,
            {
                'apiVersion': 'v1',
                'kind': 'Pod',
                'metadata': {
                    'labels': {'app': labelstr},
                    'generateName': name,
                    'namespace': namespace
                },
                'spec': {
                    'containers': [{
                        'image': image,
                        'imagePullPolicy': 'IfNotPresent',
                        'name': 'xgb',
                        'resources': {'requests': {'cpu': num_cpus_per_pod}, 'limits': {'cpu': num_cpus_per_pod}},
                        'env': [{'name': 'POD_ID','value':str(pod_id)},
                                {'name': 'OMP_NUM_THREADS','value':str(num_cpus_per_pod)},
                                {'name': 'MINIO_SERVICE_IP','value':str(minio_service_ip)}],
                     'selector': {'app': labelstr}}],
                }
            }
        )
        print("Created Pod")
        return (pod.metadata.name)
    except Exception as e:
        print("failed - %s" % e)
        return 0

def delete_pod(name, namespace):
    try:
        config.load_incluster_config()
    except:
        config.load_kube_config()

    try:
       print("trying to delete")
       sys.stdout.flush()
       apiv1 = client.CoreV1Api()
       resp = apiv1.delete_namespaced_pod(name, namespace)
       print("trying to delete resp:% " % resp)
    except Exception as e:
       print("Exception when calling CoreV1Api->delete_namespaced_pod_template: %s\n" % e)
       sys.stdout.flush()

if __name__ == '__main__':
    sys.exit(main())
