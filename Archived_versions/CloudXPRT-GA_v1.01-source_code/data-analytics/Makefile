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

# set go path to CNB directory
GOPATH := $(shell pwd)
export GOPATH

all: cnbrun monitor setup-environment reset-environment

# These Phony target are required since we have identical directory names
.PHONY: cnbrun

# packages needed for building binaries
go-get:
	go get "github.com/spf13/viper"
	go get "github.com/valyala/fasthttp"
	go get "gopkg.in/Shopify/sarama.v2"
	go get "github.com/prometheus/client_golang/api"
	go get "github.com/prometheus/client_golang/api/prometheus/v1"
	go get "github.com/antchfx/xmlquery"
	go get "github.com/levenlabs/golib/timeutil"
	go get "github.com/leesper/go_rng"
	go get "github.com/asticode/go-astikit"
	go get "github.com/montanaflynn/stats"

# binaries
cnbrun:
	cd $(GOPATH)/cnbrun && go build -o cnbrun cnbrun.go
monitor:
	cd $(GOPATH)/monitoring && go build monitoring.go
setup-environment:
	go build -o ../installation/bin/setup-environment ../installation/bin/setup-environment.go
reset-environment:
	go build -o ../installation/bin/reset-environment ../installation/bin/reset-environment.go

# Creating build package for Ubuntu
cnb-release = CloudXPRT_v1.01_data-analytics
build: all
	mkdir -p build
	mkdir -p $(cnb-release)
	cp README.md ../LICENSE ../SystemInfo.csv $(cnb-release)

	# installation
	mkdir -p $(cnb-release)/installation/bin
	cp ../installation/cluster_config.* $(cnb-release)/installation
	cp ../installation/*.sh $(cnb-release)/installation
	cp ../installation/*.tar.gz $(cnb-release)/installation
	cp ../installation/bin/*.sh $(cnb-release)/installation/bin
	cp ../installation/bin/*-environment $(cnb-release)/installation/bin
	cd $(cnb-release)/installation && tar xf kubespray.tar.gz && rm kubespray.tar.gz

	# setup
	cp -ar setup $(cnb-release)/

	# cnbrun
	cp -ar cnbrun $(cnb-release)/
	rm $(cnb-release)/cnbrun/cnbrun.go
	cp monitoring/monitoring $(cnb-release)/cnbrun/

	# cleanup
	tar czvf build/$(cnb-release).tar.gz $(cnb-release)
	rm -rf $(cnb-release)

# Clean up binaries
clean:
	-rm cnbrun/cnbrun
	-rm monitoring/monitoring
	-rm ../installation/bin/setup-environment
	-rm ../installation/bin/reset-environment
	-rm -rf build
