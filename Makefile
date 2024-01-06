SHELL := /bin/bash

.ONESHELL:
.PHONY: docker linux linux-rpm darwin darwin-zip windows windows-zip test vendor

## version
LSB = $(shell lsb_release -i -s)$(shell lsb_release -r -s)
VER = $(shell ./dist/version.sh)
ARCH = $(shell uname -m)

## declare directory
SD = $(shell pwd)
BD = "$(SD)/build"
CD = "$(SD)/build/coverage"
LIN_DIR ?= "openlan-$(LSB)-$(VER).$(ARCH)"
WIN_DIR ?= "openlan-windows-$(VER).$(ARCH)"
MAC_DIR ?= "openlan-darwin-$(VER).$(ARCH)"

## declare flags
MOD = github.com/luscis/openlan/pkg/libol
LDFLAGS += -X $(MOD).Date=$(shell date +%FT%T%z)
LDFLAGS += -X $(MOD).Version=$(VER)

build: test pkg

pkg: clean linux-rpm linux-bin windows-gzip darwin-gzip ## build all plaftorm packages

gzip: linux-gzip windows-gzip darwin-gzip ## build all plaftorm gzip

help: ## show make targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);\
	printf " \033[36m%-20s\033[0m  %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## all platform
bin: linux windows darwin ## build all platform binary

## prepare environment
env: update
	@mkdir -p $(BD)
	@go version
	@gofmt -w -s ./pkg ./cmd

update:
	git submodule init
	git submodule update

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

docker: linux-bin docker-rhel docker-deb ## build docker images

docker-rhel:
	cp -rf $(SD)/docker/centos $(BD)
	cd $(BD) && sudo docker build -t luscis/openlan:$(VER).$(ARCH) --build-arg BIN=$(LIN_DIR).bin -f centos/Dockerfile  .

docker-deb:
	cp -rf $(SD)/docker/debian $(BD)
	cd $(BD) && sudo docker build -t luscis/openlan:$(VER).$(ARCH).deb --build-arg BIN=$(LIN_DIR).bin -f debian/Dockerfile  .

docker-compose:
	rm -rf /tmp/openlan.c && mkdir /tmp/openlan.c && \
	cp -rvf ./dist/rootfs/{var,etc} /tmp/openlan.c && \
	cp -rvf ./docker/docker-compose.yml /tmp/openlan.c  && \
	echo "######## Lunch a openlan cluster #######" && \
	echo "$ cd /tmp/openlan.c" && \
	echo "$ docker-compose up -d"

linux: env ## build linux binary
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openudp ./cmd/openudp
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan ./cmd/main.go
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-proxy ./cmd/proxy
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point ./cmd/point_linux
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-switch ./cmd/switch

rpm: env ## build rpm packages
	mkdir -p ~/rpmbuild/SPECS
	mkdir -p ~/rpmbuild/SOURCES
	sed -e "s/Version:.*/Version:\ $(VER)/" $(SD)/dist/openlan.spec.in > ~/rpmbuild/SPECS/openlan.spec
	@dist/spec.sh
	rpmbuild -ba ~/rpmbuild/SPECS/openlan.spec

linux-gzip: install ## build linux packages
	@rm -rf $(LIN_DIR).tar.gz
	tar -cf $(LIN_DIR).tar $(LIN_DIR) && mv $(LIN_DIR).tar $(BD)
	@rm -rf $(LIN_DIR)
	gzip -f $(BD)/$(LIN_DIR).tar

linux-bin: linux-gzip ## build linux install binary
	@cat $(SD)/dist/rootfs/var/openlan/script/install.sh > $(BD)/$(LIN_DIR).bin && \
	echo "__ARCHIVE_BELOW__:" >> $(BD)/$(LIN_DIR).bin && \
	cat $(BD)/$(LIN_DIR).tar.gz >> $(BD)/$(LIN_DIR).bin && \
	chmod +x $(BD)/$(LIN_DIR).bin && \
	echo "Save to $(BD)/$(LIN_DIR).bin"

install: env linux ## install packages
	@mkdir -p $(LIN_DIR)
	@cp -rf $(SD)/dist/rootfs/{etc,var,usr} $(LIN_DIR)
	@mkdir -p $(LIN_DIR)/var/openlan/{cert,openvpn,point,l2tp,dhcp}
	@cp -rf $(SD)/dist/cert/openlan/cert $(LIN_DIR)/var/openlan
	@cp -rf $(SD)/dist/cert/openlan/ca/ca.crt $(LIN_DIR)/var/openlan/cert
	@cp -rf $(SD)/pkg/public $(LIN_DIR)/var/openlan
	@mkdir -p $(LIN_DIR)/usr/bin
	@cp -rf $(BD)/{openudp,openlan} $(LIN_DIR)/usr/bin
	@cp -rf $(BD)/{openlan-point,openlan-proxy,openlan-switch} $(LIN_DIR)/usr/bin

## cross build for windows
windows: ## build windows binary
	GOOS=windows go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.exe ./cmd/point_windows
	GOOS=windows go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-proxy.exe ./cmd/proxy
	GOOS=windows go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan.exe ./cmd/main.go

windows-gzip: env windows ## build windows packages
	@rm -rf $(WIN_DIR) && mkdir -p $(WIN_DIR)
	@rm -rf $(WIN_DIR).tar.gz

	@cp -rf $(SD)/dist/rootfs/etc/openlan/point.json.example $(WIN_DIR)/point.json
	@cp -rf $(BD)/openlan-point.exe $(WIN_DIR)
	@cp -rf $(SD)/dist/rootfs/etc/openlan/proxy.json.example $(WIN_DIR)/proxy.json
	@cp -rf $(BD)/openlan-proxy.exe $(WIN_DIR)

	tar -cf $(WIN_DIR).tar $(WIN_DIR) && mv $(WIN_DIR).tar $(BD)
	@rm -rf $(WIN_DIR)
	gzip -f $(BD)/$(WIN_DIR).tar

windows-syso: ## build windows syso
	rsrc -manifest ./cmd/point_windows/main.manifest -ico ./cmd/point_windows/main.ico  -o ./cmd/point_windows/main.syso

## cross build for osx
osx: darwin

darwin: env ## build darwin binary
	GOOS=darwin go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.din ./cmd/point_darwin
	GOOS=darwin go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan.din ./cmd/main.go
	GOOS=darwin go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-proxy.din ./cmd/proxy

darwin-gzip: env darwin ## build darwin packages
	@rm -rf $(MAC_DIR) && mkdir -p $(MAC_DIR)
	@rm -rf $(MAC_DIR).tar.gz

	@cp -rf $(SD)/dist/rootfs/etc/openlan/point.json.example $(MAC_DIR)/point.json
	@cp -rf $(BD)/openlan-point.din $(MAC_DIR)

	tar -cf $(MAC_DIR).tar $(MAC_DIR) && mv $(MAC_DIR).tar $(BD)
	@rm -rf $(MAC_DIR)
	gzip -f $(BD)/$(MAC_DIR).tar

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
cover: env ## execute unit test and output coverage
	@rm -rvf $(CD) && mkdir -p $(CD)
	@go test -mod=vendor github.com/luscis/openlan/pkg/access -coverprofile=$(CD)/0.out -race -covermode=atomic
	@go test -mod=vendor github.com/luscis/openlan/pkg/libol -coverprofile=$(CD)/1.out -race -covermode=atomic
	@go test -mod=vendor github.com/luscis/openlan/pkg/models -coverprofile=$(CD)/2.out -race -covermode=atomic
	@go test -mod=vendor github.com/luscis/openlan/pkg/cache -coverprofile=$(CD)/3.out -race -covermode=atomic
	@go test -mod=vendor github.com/luscis/openlan/pkg/config -coverprofile=$(CD)/4.out -race -covermode=atomic
	@go test -mod=vendor github.com/luscis/openlan/pkg/network -coverprofile=$(CD)/5.out -race -covermode=atomic

	@echo 'mode: atomic' > $(SD)/coverage.out && \
	tail -q -n +2 $(CD)/*.out >> $(SD)/coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean: ## clean cache
	rm -rvf ./build
