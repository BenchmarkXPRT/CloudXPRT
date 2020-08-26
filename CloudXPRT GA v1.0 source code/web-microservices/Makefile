# Set go path to CNB directory
GOPATH := $(shell pwd)
export GOPATH

all: autoloader cnbrun gobench postprocess setup-environment reset-environment

# These Phony targets are required since we have identical directory names
.PHONY: cnbrun gobench postprocess

# Packages needed for building binaries
go-get:
	go get "github.com/spf13/viper"
	go get "github.com/c9s/goprocinfo/linux"
	go get "github.com/olekukonko/tablewriter"
	go get "github.com/valyala/fasthttp"
	go get "gonum.org/v1/plot"
	go get "gonum.org/v1/plot/plotter"
	go get "gonum.org/v1/plot/vg"

# Binaries
autoloader:
	cd $(GOPATH)/gobench/autoloader && go build -o ../../cnbrun/autoloader autoloader.go
cnbrun:
	cd $(GOPATH)/cnbrun && go build -o cnbrun cnbrun.go
gobench:
	cd $(GOPATH)/gobench && go build -o ../cnbrun/gobench gobench.go
postprocess:
	cd $(GOPATH)/postprocess && go build -o postprocess postprocess.go
setup-environment:
	go build -o ../installation/bin/setup-environment ../installation/bin/setup-environment.go
reset-environment:
	go build -o ../installation/bin/reset-environment ../installation/bin/reset-environment.go

# Creating build package for Ubuntu
cnb-release = CloudXPRT_v1.00_web-microservices
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

	# cnbrun
	mkdir -p $(cnb-release)/cnbrun/cassandra/onprem
	cd cnbrun && cp -rf autoloader cnbrun config.json gobench mc.sh mc.remote.sh cleanups.sh system_info.sh services ../$(cnb-release)/cnbrun
	cp -r cnbrun/metrics-server $(cnb-release)/cnbrun
	cp cnbrun/cassandra/onprem/* $(cnb-release)/cnbrun/cassandra/onprem
	cp cnbrun/cassandra/schema.cql $(cnb-release)/cnbrun/cassandra
	cd postprocess && cp plot.json postprocess report.html ../$(cnb-release)/cnbrun
	cp -r postprocess/css $(cnb-release)/cnbrun

	# cnb-ui-demo
	mkdir -p $(cnb-release)/cnb-ui-demo/cassandra/onprem
	cp cnbrun/cassandra/onprem/* $(cnb-release)/cnb-ui-demo/cassandra/onprem
	cp cnbrun/cassandra/schema.cql $(cnb-release)/cnb-ui-demo/cassandra/
	cp cnbrun/services.sh $(cnb-release)/cnb-ui-demo/
	cp -rf cnbrun/services $(cnb-release)/cnb-ui-demo/

	# cleanup
	tar czvf build/$(cnb-release)-Onprem.tar.gz $(cnb-release)
	rm -rf $(cnb-release)

# Clean up binaries
clean:
	-rm cnbrun/autoloader
	-rm cnbrun/cnbrun
	-rm cnbrun/gobench
	-rm postprocess/postprocess
	-rm ../installation/bin/setup-environment
	-rm ../installation/bin/reset-environment
	-rm -rf build
