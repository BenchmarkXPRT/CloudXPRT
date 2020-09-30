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

from minio import Minio
from minio.error import ResponseError
import os
import sys
from os import listdir
##
# Follow bucket naming rules
#'10.233.54.122:9000'
#"cnb-ml-bucket"
if __name__ == "__main__":
    minioserviceIP = sys.argv[1]
    bucketname = sys.argv[2]
    print("Deleting minio bucket \n")

    minioClient = Minio(minioserviceIP,
                  access_key='minio',
                  secret_key='minio123',
                  secure=False)

    try:
       if minioClient.bucket_exists(bucketname):
           objects = minioClient.list_objects(bucketname,recursive=True)
           for obj in objects:
              print(obj.bucket_name, obj.object_name.encode('utf-8'), obj.last_modified,
                obj.etag, obj.size, obj.content_type)
              minioClient.remove_object(bucketname, obj.object_name)

           minioClient.remove_bucket(bucketname)
    except ResponseError as err:
       print(err)
