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

FROM ubuntu:18.04
LABEL Description="This is the base image for XGBoost under Ubuntu 18.04"

# Update Ubuntu Software repository
RUN apt update
RUN apt install -y git build-essential cmake g++

RUN apt install -y python3
RUN apt install -y python3-pip
RUN pip3 install requests scikit-learn pandas
RUN /bin/bash -c "python3 --version"

# kafka, Zookeeper and MinIO
RUN pip3 install kafka-python kazoo minio
ENV PYTHONUNBUFFERED=0

#RUN pip3 install xgboost v1.0
WORKDIR /root/
RUN mkdir xgboost \
    && cd xgboost \
    && git clone --recursive https://github.com/dmlc/xgboost . \
    && git checkout 27a8e36fc36fb150a4af80f08c5501df94d82fba \
    && git submodule init \
    && git submodule update \
    && make -j8 \
    && cd python-package \
    && python3 setup.py install \
    && cd ../

COPY xgboost-bench.tar.gz .
RUN tar -xvf xgboost-bench.tar.gz
RUN cd /root/xgboost-bench/hist_method/
RUN chmod 755 *

WORKDIR /root/xgboost-bench/hist_method/

RUN apt-get clean
EXPOSE 8078
CMD    ["python3","xgboost_hist_method_bench_kafka.py"]
