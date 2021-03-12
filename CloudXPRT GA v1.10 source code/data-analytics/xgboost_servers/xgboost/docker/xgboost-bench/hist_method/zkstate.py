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

from kazoo.client import KazooClient
from kazoo.exceptions import NoNodeError, NodeExistsError

#ZK_HOSTS = 'localhost:2181'
ZK_HOSTS = 'zookeeper:2181'

class ZKState():
    def __init__(self, path, timeout=30):
        super(ZKState, self).__init__()
        self._zk = KazooClient(hosts=ZK_HOSTS, timeout=timeout)
        self._zk.start(timeout=timeout)
        self._path = path
        self._zk.ensure_path(path)

    def processed(self):
        return self._zk.exists(self._path+"/complete")

    def process_start(self):
        if self.processed():
            return False
        if self._zk.exists(self._path+"/processing"):
            return False
        try:
            self._zk.create(self._path+"/processing", ephemeral=True)
            return True
        except NodeExistsError: # another process wins
            return False

    def process_end(self):
        self._zk.create(self._path+"/complete")
        self._zk.delete(self._path+"/processing")

    def process_abort(self):
        try:
            self._zk.delete(self._path+"/processing")
        except NoNodeError:
            pass

    def close(self):
        self._zk.stop()
        self._zk.close()
