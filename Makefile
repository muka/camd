.PHONY: build

DOCKER_IMAGE ?= opny/camd

BUILD_PATH ?= ./build

build: build/amd64 build/arm64 build/arm

build/amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH}/camd-amd64 .

build/arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${BUILD_PATH}/camd-arm64 .

build/arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ${BUILD_PATH}/camd-arm .

docker/build/amd64:
	docker build . -t ${DOCKER_IMAGE}-amd64 --build-arg ARCH=amd64

docker/build/arm64:
	docker build . -t ${DOCKER_IMAGE}-arm64 --build-arg ARCH=arm64

docker/build/arm:
	docker build . -t ${DOCKER_IMAGE}-arm --build-arg ARCH=arm

docker/push:
	docker push ${DOCKER_IMAGE}	