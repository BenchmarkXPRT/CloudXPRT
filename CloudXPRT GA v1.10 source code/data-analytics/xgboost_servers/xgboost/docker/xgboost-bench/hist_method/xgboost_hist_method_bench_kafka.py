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

#*******************************************************************************
# Copyright 2017-2019 by Contributors
# \file xgboost_hist_method_bench.py
# \brief a benchmark for 'hist' tree_method on both CPU/GPU arhitectures
# \author Egor Smirnov
#*******************************************************************************

import argparse
import xgboost as xgb
from bench_utils import *

# kafka, Zokeeper and MinIO modules
import os
import time
import json
from messaging import Consumer
from messaging import Producer
from zkstate import ZKState
# match cdn-webserver - NumPartitions
partitionID = os.getenv('POD_ID') # using the pod id as partition id
KAFKA_TOPIC = "ktopic-xgboost" #testTopicTranscodeFinal1
KAFKA_GROUP = "kgroup-xgboost" #groupfinal
KAFKA_TOPIC_DONE = "ktopic-xgboost-done" #TopicTranscodeDone1
hw = 'cpu'
data_set = 'higgs1m'

N_PERF_RUNS = 1
DTYPE=np.float32

xgb_params = {
    'alpha':                        0.9,
    'max_bin':                      256,
    'scale_pos_weight':             2,
    'learning_rate':                0.1,
    'subsample':                    1,
    'reg_lambda':                   1,
    "min_child_weight":             0,
    'max_depth':                    8,
    'max_leaves':                   2**8,
}

def xbg_fit():
    global model_xgb
    dtrain = xgb.DMatrix(x_train, label=y_train)
    model_xgb = xgb.train(xgb_params, dtrain, xgb_params['n_estimators'])

def xgb_predict_of_train_data():
    global result_predict_xgb_train
    dtest = xgb.DMatrix(x_train)
    result_predict_xgb_train = model_xgb.predict(dtest)

def xgb_predict_of_test_data():
    global result_predict_xgb_test
    dtest = xgb.DMatrix(x_test)
    result_predict_xgb_test = model_xgb.predict(dtest)


def load_dataset(dataset):
    global x_train, y_train, x_test, y_test
    print("Loading... " + dataset + " \n")
    try:
        os.mkdir(DATASET_DIR)
    except:
        pass

    datasets_dict = {
        'higgs1m': load_higgs1m,
        'msrank-10k': load_msrank_10k,
        'airline-ohe':load_airline_one_hot
    }

    x_train, y_train, x_test, y_test, n_classes = datasets_dict[dataset](DTYPE)

    if n_classes == -1:
        xgb_params['objective'] = 'reg:squarederror'
    elif n_classes == 2:
        xgb_params['objective'] = 'binary:logistic'
    else:
        xgb_params['objective'] = 'multi:softprob'
        xgb_params['num_class'] = n_classes

def parse_args():
    global N_PERF_RUNS
    parser = argparse.ArgumentParser()
    parser.add_argument('--n_iter', required=False, type=int, default=1000)
    parser.add_argument('--n_runs', default=N_PERF_RUNS, required=False, type=int)
    parser.add_argument('--hw', choices=['cpu', 'gpu'], metavar='stage', required=False, default='cpu')
    parser.add_argument('--log', metavar='stage', required=False, type=bool, default=False)
    parser.add_argument('--dataset', choices=['higgs1m', "airline-ohe", "msrank-10k"],
            metavar='stage', required=False, default="higgs1m")

    args = parser.parse_args()
    N_PERF_RUNS = args.n_runs

    xgb_params['n_estimators'] = args.n_iter

    if args.log:
        xgb_params['verbosity'] = 3
    else:
         xgb_params['silent'] = 1

    if args.hw == "cpu":
        xgb_params['tree_method'] = 'hist'
        xgb_params['predictor']   = 'cpu_predictor'
    elif args.hw == "gpu":
        xgb_params['tree_method'] = 'gpu_hist'
        xgb_params['predictor']   = 'gpu_predictor'

    load_dataset(args.dataset)

#=============
def parse_args_kafka(data_set):

    n_iter = 1000
    n_runs = 1
    hw = 'cpu'
    log=False
    dataset = data_set

    xgb_params['n_estimators'] = n_iter
    xgb_params['silent'] = 1
    #hw == "cpu":
    xgb_params['tree_method'] = 'hist'
    xgb_params['predictor']   = 'cpu_predictor'

    load_dataset(dataset)

def execute_xgboost(data_set):
    parse_args_kafka(data_set)

    print("Running ...")
    measure(xbg_fit,                   "XGBOOST training            ", N_PERF_RUNS)
    measure(xgb_predict_of_train_data, "XGBOOST predict (train data)", N_PERF_RUNS)
    measure(xgb_predict_of_test_data,  "XGBOOST predict (test data) ", N_PERF_RUNS)

    print("Compute quality metrics...")

    train_loglos = compute_logloss(y_train, result_predict_xgb_train)
    test_loglos = compute_logloss(y_test, result_predict_xgb_test)

    print("LogLoss for train data set = {:.6f}".format(train_loglos))
    print("LogLoss for test  data set = {:.6f}".format(test_loglos))

def MinIORetrieveFiles():
    from minio import Minio
    from minio.error import ResponseError
    ## use minio service ip address

    minio_ip=os.environ['MINIO_SERVICE_IP']
    print("Testing minio access using IP: " + minio_ip + " \n")
    minioClient = Minio(minio_ip, access_key='minio', secret_key='minio123', secure=False)

    # Get a full object and prints the original object stat information.
    # try:
    #     #print(minioClient.fget_object('cnb-ml-bucket', 'msrank.tar.gz', '/var/www/archive/msrank.tar.gz'))
    #     print(minioClient.fget_object('cnb-ml-bucket', 'MSRank-test', '/var/www/archive/test.txt'))
    #     print(minioClient.fget_object('cnb-ml-bucket', 'MSRank-train', '/var/www/archive/train.txt'))
    #     print(minioClient.fget_object('cnb-ml-bucket', 'MSRank-vali', '/var/www/archive/vali.txt'))
    # except ResponseError as err:
    #     print("MINIO ERROR msrank.tar.gz:" + err)

    try:
        print(minioClient.fget_object('cnb-ml-bucket', 'HIGGS.csv.gz', '/var/www/archive/HIGGS.csv.gz'))
    except ResponseError as err:
        print("MINIO ERROR HIGGS.csv.gz:" + err)

    # try:
    #     print(minioClient.fget_object('cnb-ml-bucket', 'train-10m.csv', '/var/www/archive/train-10m.csv'))
    # except ResponseError as err:
    #     print("MINIO ERROR airline-ohe train-10m:" + err)
    #
    # try:
    #     print(minioClient.fget_object('cnb-ml-bucket', 'test.csv', '/var/www/archive/test.csv'))
    # except ResponseError as err:
    #     print("MINIO ERROR airline-ohe test:" + err)

    #try:
    #    df = pd.read_csv('/var/www/archive/abalone.data')
    #    print(df.head(5))
    #except ResponseError as err:
    #    print("MINIO ERROR:" + err)

def sendDoneMessage(json_data):
    producer = Producer()
    producer.send(KAFKA_TOPIC_DONE,json_data)

def main(streamparam):

    print(streamparam) #a dict is returned
    print(streamparam["dataset"])

    startTime = 0.0
    endTime = 0.0
    totalDuration = 0.0

    # part of kafka definition
    data_set = streamparam["data"]
    ompCPUs = streamparam["ompcpus"]
    hw = streamparam["hw"]

    #print(os.environ['OMP_NUM_THREADS'])
    #os.environ["OMP_NUM_THREADS"] = ompCPUs # export OMP_NUM_THREADS=4
    print('Parameters from kafka message:')
    print('data_set = ',data_set)
    print('ompCPUs = ',ompCPUs)
    print('hw = ',hw)
    print(os.environ['OMP_NUM_THREADS'])

    stream = streamparam["dataset"]
    key = stream.replace("/", "_")
    zk = ZKState(key)
    #print(key)
    if zk.processed():
        print("  already processed by zk")
        zk.close()
        return

    if zk.process_start():
       try:
           print("  staring execution of xgboost...")
           startTime = time.time()
           execute_xgboost(data_set)
           endTime = time.time()
           totalDuration = endTime - startTime
           print("  KAFKA_TOPIC_DONE: " + KAFKA_TOPIC_DONE)
           print("  Duration:" + str(round(totalDuration,2)))
           data = {}
           data['key'] = key
           data['duration'] = str(round(totalDuration,2))
           data['PodID'] = str(partitionID)
           json_data = json.dumps(data)
           zk.process_end()
           sendDoneMessage(json_data)
       except Exception as e:
           print(str(e))
           zk.process_abort()

#def main():
#    parse_args()

#=============

if __name__ == '__main__':
    #c = Consumer(KAFKA_GROUP +""+ str(partitionID ))
    c = Consumer(KAFKA_GROUP)
    print("  xgboost_bench_kafka - KAFKA_TOPIC: " + KAFKA_TOPIC + " partitionID: "+ str(partitionID))
    print(os.environ['OMP_NUM_THREADS'])
    print(os.environ['MINIO_SERVICE_IP'])

    #minio
    MinIORetrieveFiles()

    while True:
        try:
            print("  messages...")
            #for message in c.messages(KAFKA_TOPIC,partitionID):
            for message in c.messages(KAFKA_TOPIC):
                print("  MESSAGE:")
                print(message)
                print("  key decoded")
                print(message.key.decode('ASCII'))
                # use to control which POD executes what message
                # if(message.key.decode('ASCII') == str(partitionID)):
                print("  POD "+ partitionID + " processing... ")
                main(json.loads(message.value.decode('utf-8')))
        except Exception as e:
            print(str(e))
        time.sleep(2)

    #main()
