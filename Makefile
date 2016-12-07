#
# Copyright (C) 2011 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

SHELL := /bin/bash
GOPATH := $(shell pwd)/_gopath
ORG := github.com/jimmidyson
REPOPATH ?= $(ORG)/kube-client-gen
PACKAGES ?= $(shell glide novendor)

.PHONY: build
build: build/kube-client-gen

build/kube-client-gen: gopath $(shell find -name *.go)
	cd $(GOPATH)/src/$(REPOPATH) && CGO_ENABLED=0 go build -o build/generate -ldflags="-s -w -extldflags '-static'" .

.PHONY: all
all: test build build-plugins

.PHONY: test
test: gopath
	cd $(GOPATH)/src/$(REPOPATH) && go test -race -v $(PACKAGES)

.PHONY: gopath
gopath: $(GOPATH)/src/$(ORG)

$(GOPATH)/src/$(ORG):
	mkdir -p $(GOPATH)/src/$(ORG)
	ln -s -f $(shell pwd) $(GOPATH)/src/$(ORG)
	ln -s -f $(shell pwd)/vendor/golang.org $(GOPATH)/src
	ln -s -f $(shell pwd)/vendor/github.com/* $(GOPATH)/src/github.com/

.PHONY: clean
clean:
	rm -rf $(GOPATH) build
