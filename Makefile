.PHONY: build release try

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
REPO_PATH:=github.com/byrnedo/dockdash
GO_IMAGE:=golang:1.5.1
RELEASE_VERSION:=

build:
	go get github.com/tools/godep && godep go build -v -o build/dockdash

prepare-release:
	bash _make_scripts/create_release_artifacts.sh

release:
	bash _make_scripts/release.sh $(RELEASE_VERSION)

d-build:
	mkdir -p $(ROOT_DIR)/build && \
	docker run --rm -it\
		-v "$(ROOT_DIR)":/usr/src/$(REPO_PATH)\
		-w /usr/src/$(REPO_PATH) $(GO_IMAGE) \
		go get github.com/tools/godep && \
		godep go build -v -o build/dockdash

try:
	docker run --rm -it\
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v "$(ROOT_DIR)":/usr/src/$(REPO_PATH) \
		-w /usr/src/$(REPO_PATH) $(GO_IMAGE) \
		go get -d -v && go build -v -o /tmp/dockdash && /tmp/dockdash

