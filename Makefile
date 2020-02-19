
DOCKER_USERNAME=devnulled
GITHUB_USERNAME=devnulled
PROJECT_NAME = certsman
BINARY = certsman
GOARCH = amd64
VERSION?=0.05

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

DOCKER_PROJECT := ${DOCKER_USERNAME}/${PROJECT_NAME}
DOCKER_TAG := ${DOCKER_PROJECT}:${VERSION}
GOMOD_NAME = github.com/${GITHUB_USERNAME}/${PROJECT_NAME}GOMOD_NAME = github.com/${GITHUB_USERNAME}/${PROJECT_NAME}

# Symlink into GOPATH

PROJECT_ROOT_PATH=${GOPATH}/github.com/${GITHUB_USERNAME}/${PROJECT_NAME}
BUILD_PATH=${PROJECT_ROOT_PATH}/cmd
CURRENT_DIR=$(shell pwd)
BIN_DIR=${CURRENT_DIR}/bin
BUILD_PATH_LINK=$(shell readlink ${BUILD_PATH})

PKGS = $(shell go list ./... | grep -v /vendor/)
SOURCES = $(shell find . -name '*.go' -not -path "*/vendor/*")

# Setup the -ldflags option for go build here, interpolate the variable values
# LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"
LDFLAGS =

DEPS := \
	go\

# Build the project
all: dependencies clean generate test darwin linux

dependencies:
	@for p in $(DEPS); do \
		$(call isinstalled,$$p) || exit 1; \
	done

linux:
	cd ${BUILD_PATH}; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BIN_DIR}/${BINARY}-linux-${GOARCH} . ; \
	cd - >/dev/null

darwin:
	cd ${BUILD_PATH}; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BIN_DIR}/${BINARY}-darwin-${GOARCH} . ; \
	chmod +x ${BIN_DIR}/${BINARY}-darwin-${GOARCH}; \
	cd - >/dev/null

windows:
	cd ${BUILD_PATH}; \
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BIN_DIR}/${BINARY}-windows-${GOARCH}.exe . ; \
	cd - >/dev/null

test:
	go test $(PKGS)

benchmark: 
	go test -benchmem -count 1000 ${PKGS}

run:
	${BIN_DIR}/${BINARY}-darwin-${GOARCH} -thisArg hello

build-test: test darwin run

docker-build: clean linux
	docker build -t ${DOCKER_TAG} .

docker-pull:
	docker pull ${DOCKER_PROJECT}

docker-push:
	docker push ${DOCKER_PROJECT}

docker-publish: docker-build docker-push

mod-init:
	go mod init ${GOMOD_NAME}

vet:
	go vet $(PKGS)

fmt:
	go fmt $(PKGS)

generate:
	go generate $(PKGS)

clean:
	-rm -f ${BIN_DIR}/${BINARY}-*

basic-load-test:
	hey -n 1000 -m GET http://localhost:8080/certtest/



.PHONY: dependencies linux darwin windows test run docker-build mod-init vet fmt generate clean benchmark

define isinstalled
hash $(1) >/dev/null 2>/dev/null || { echo >&2 "The binary '$(1)' is required but not found. Aborting."; exit 1; }
endef