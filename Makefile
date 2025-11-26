SHELL := /bin/bash

.ONESHELL:
.PHONY: docker linux darwin darwin-zip windows windows-zip test vendor

## version
LSB = $(shell lsb_release -i -s)$(shell lsb_release -r -s)
VER = $(shell ./dist/version.sh)

## declare directory
SD = $(shell pwd)
BD = "$(SD)/build"
CD = "$(SD)/build/coverage"

ARCH ?=amd64
LIN_DIR ?= "openlan-$(VER).$(ARCH)"
WIN_DIR ?= "openceci-windows-$(VER).$(ARCH)"
MAC_DIR ?= "openceci-darwin-$(VER).$(ARCH)"

## declare flags
MOD = github.com/luscis/openlan/pkg/libol
LDFLAGS += -X $(MOD).Date=$(shell date +%F)
LDFLAGS += -X $(MOD).Version=$(VER)
LDFLAGS += -X $(MOD).Commit=$(shell git rev-parse --short HEAD)

build: test pkg

pkg: clean linux-bin windows-gzip darwin-gzip ## build all plaftorm packages

gzip: linux-gzip windows-gzip darwin-gzip ## build all plaftorm gzip

help: ## show make targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);\
	printf " \033[36m%-20s\033[0m  %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## all platform
bin: linux windows darwin ## build all platform binary

## prepare environment
init:
	mkdir -p $(BD)
	gofmt -w -s ./pkg ./cmd

update: ## update source code
	git pull
	git submodule init
	git submodule update --remote --merge

vendor:
	go clean -modcache
	go mod tidy
	go mod vendor -v

config:
	cd $(BD) && mkdir -p config/openlan
	cp -rvf ../docker/docker-compose.yml config/openlan

	mkdir -p config/openlan/etc
	cp -rvf ../dist/rootfs/etc/openlan config/openlan/etc

	cd config && tar -cf ../config.tar openlan && cd ..
	gzip -f config.tar

builder:
	docker run -d -it \
	--env http_proxy="${http_proxy}" --env https_proxy="${https_proxy}" \
	--volume $(SD)/:/opt/openlan --volume $(shell echo ~)/.ssh:/root/.ssh \
	--name openlan_builder debian:bullseye bash
	docker exec openlan_builder bash -c "apt update && apt install -y git lsb-release wget make gcc devscripts"
	docker exec openlan_builder bash -c "apt install -y net-tools make build-essential libnss3-dev pkg-config libevent-dev libunbound-dev bison flex libsystemd-dev libcurl4-nss-dev libpam0g-dev libcap-ng-dev libldns-dev xmlto"
	docker exec openlan_builder badh -c "apt install -y htmldoc libaudit-dev libkrb5-dev libldap2-dev libnss3-tools libselinux1-dev man2html"
	docker exec openlan_builder bash -c "wget https://golang.google.cn/dl/go1.23.0.linux-amd64.tar.gz && tar -xf go1.23.0.linux-amd64.tar.gz -C /usr/local"
	docker exec openlan_builder bash -c "cd /usr/local/bin && ln -s ../go/bin/go . && ln -s ../go/bin/gofmt ."
	docker exec openlan_builder git config --global --add safe.directory /opt/openlan
	docker exec openlan_builder git config --global --add safe.directory /opt/openlan/dist/cert

## build libreswan
# wget http://deb.debian.org/debian/pool/main/libr/libreswan/libreswan_4.10.orig.tar.gz
# tar xvf libreswan_4.10.orig.tar.gz
# cd libreswan-4.10 && make deb

docker-rhel: docker-bin ## build image for redhat
	cp -rf $(SD)/docker/centos $(BD)
	cd $(BD) && \
	sudo docker build -t luscis/openlan:$(VER).$(ARCH).el \
	--build-arg linux_bin=$(LIN_DIR).bin --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" \
	--file centos/Dockerfile .

docker-deb: docker-bin ## build image for debian
	cp -rf $(SD)/docker/debian $(BD)
	cd $(BD) && \
	sudo docker build -t luscis/openlan:$(VER).$(ARCH).deb \
	--build-arg linux_bin=$(LIN_DIR).bin --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" \
	--file debian/Dockerfile .

docker-bin:
	docker exec openlan_builder bash -c "cd /opt/openlan && make linux-bin"

docker-tar:
	docker exec openlan_builder bash -c "cd /opt/openlan && make linux-tar"

docker: docker-deb ## build docker images

docker-builder: builder ## create a builder

docker-compose: ## create a compose files
	rm -rf /tmp/openlan.c && mkdir /tmp/openlan.c && \
	cp -rvf ./dist/rootfs/{var,etc} /tmp/openlan.c && \
	cp -rvf ./docker/docker-compose.yml /tmp/openlan.c  && \
	echo "######## Lunch a openlan cluster #######" && \
	echo "$ cd /tmp/openlan.c" && \
	echo "$ docker-compose up -d"

ceci: linux-ceci darwin-ceci windows-ceci ## build all platform ceci

linux: linux-access linux-switch linux-ceci linux-proxy ## build linux binary
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan ./cmd/main.go

linux-access:
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-access ./cmd/access

linux-switch:
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-switch ./cmd/switch

linux-ceci:
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openceci ./cmd/ceci

linux-proxy:
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-proxy ./cmd/proxy

linux-tar: install ## build linux packages
	@rm -rf $(LIN_DIR).tar.gz
	tar -cf $(LIN_DIR).tar $(LIN_DIR) && mv $(LIN_DIR).tar $(BD)
	@rm -rf $(LIN_DIR)
	gzip -f $(BD)/$(LIN_DIR).tar

linux-bin: update linux-tar ## build linux install binary
	@cat $(SD)/dist/rootfs/var/openlan/script/install.sh > $(BD)/$(LIN_DIR).bin && \
	echo "__ARCHIVE_BELOW__:" >> $(BD)/$(LIN_DIR).bin && \
	cat $(BD)/$(LIN_DIR).tar.gz >> $(BD)/$(LIN_DIR).bin && \
	chmod +x $(BD)/$(LIN_DIR).bin && \
	echo "Save to $(LIN_DIR).bin"

install: init linux ## install packages
	@mkdir -p $(LIN_DIR)
	@cp -rf $(SD)/dist/rootfs/{etc,var,usr} $(LIN_DIR)
	@mkdir -p $(LIN_DIR)/var/openlan/{cert,openvpn,access,dhcp}
	@cp -rf $(SD)/dist/cert/openlan/cert $(LIN_DIR)/var/openlan
	@cp -rf $(SD)/dist/cert/openlan/ca/ca.crt $(LIN_DIR)/var/openlan/cert
	@cp -rf $(SD)/pkg/public $(LIN_DIR)/var/openlan
	@mkdir -p $(LIN_DIR)/usr/bin
	@cp -rf $(BD)/{openlan,openlan-switch} $(LIN_DIR)/usr/bin
	@cp -rf $(BD)/{openlan-access,openlan-proxy,openceci} $(LIN_DIR)/usr/bin
	@cp -rf $(BD)/openceci $(LIN_DIR)/usr/bin
	@echo "Installed to $(LIN_DIR)"

## cross build for windows
windows: windows-ceci windows-access ## build windows binary

windows-ceci:
	GOOS=windows GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openceci.exe ./cmd/ceci

windows-access:
	GOOS=windows GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-access.exe ./cmd/access

windows-gzip: init windows ## build windows packages
	@rm -rf $(WIN_DIR) && mkdir -p $(WIN_DIR)
	@cp -rf $(SD)/dist/rootfs/etc/openlan/http/http.yaml.example $(WIN_DIR)/ceci.yaml
	@cp -rf $(BD)/{openlan-access,openceci.exe} $(WIN_DIR)
	tar -cf $(WIN_DIR).tar $(WIN_DIR) && mv $(WIN_DIR).tar $(BD)
	gzip -f $(BD)/$(WIN_DIR).tar && rm -rf $(WIN_DIR)

## cross build for osx
osx: darwin

darwin: darwin-ceci darwin-access ## build darwin binary

darwin-ceci:
	GOOS=darwin GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openceci.dar ./cmd/ceci

darwin-access:
	GOOS=darwin GOARCH=$(ARCH) go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-access.dar ./cmd/access

darwin-gzip: env darwin ## build darwin packages
	@rm -rf $(MAC_DIR) && mkdir -p $(MAC_DIR)
	@cp -rf $(SD)/dist/rootfs/etc/openlan/http/http.yaml.example $(MAC_DIR)/ceci.yaml
	@cp -rf $(BD)/{openlan-access,openceci.dar,openceci.arm64.dar} $(MAC_DIR)
	tar -cf $(MAC_DIR).tar $(MAC_DIR) && mv $(MAC_DIR).tar $(BD)
	gzip -f $(BD)/$(MAC_DIR).tar && rm -rf $(MAC_DIR)

## unit test
test: ## execute unit test
	go clean -testcache
	go test -v -mod=vendor -bench=. github.com/luscis/openlan/pkg/access
	go test -v -mod=vendor -bench=. github.com/luscis/openlan/pkg/libol
	go test -v -mod=vendor -bench=. github.com/luscis/openlan/pkg/models
	go test -v -mod=vendor -bench=. github.com/luscis/openlan/pkg/cache
	go test -v -mod=vendor -bench=. github.com/luscis/openlan/pkg/config
	go test -v -mod=vendor -bench=. github.com/luscis/openlan/pkg/network

## coverage
cover: init ## execute unit test and output coverage
	@rm -rvf $(CD) && mkdir -p $(CD)
	go test -mod=vendor github.com/luscis/openlan/pkg/access -coverprofile=$(CD)/0.out -race -covermode=atomic
	go test -mod=vendor github.com/luscis/openlan/pkg/libol -coverprofile=$(CD)/1.out -race -covermode=atomic
	go test -mod=vendor github.com/luscis/openlan/pkg/models -coverprofile=$(CD)/2.out -race -covermode=atomic
	go test -mod=vendor github.com/luscis/openlan/pkg/cache -coverprofile=$(CD)/3.out -race -covermode=atomic
	go test -mod=vendor github.com/luscis/openlan/pkg/config -coverprofile=$(CD)/4.out -race -covermode=atomic
	go test -mod=vendor github.com/luscis/openlan/pkg/network -coverprofile=$(CD)/5.out -race -covermode=atomic

	@echo 'mode: atomic' > $(SD)/coverage.out && \
	tail -q -n +2 $(CD)/*.out >> $(SD)/coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean: ## clean cache
	rm -rvf ./build
