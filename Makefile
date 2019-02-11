BINARY_NAME=image-caching-test
DOCKERIMAGE_NAME=image-caching-test
DOCKERIMAGE_TAG=dev

all: build docker

build:
	GOOS=linux go build -v -o ./bin/${BINARY_NAME} ./cmd/main.go

docker:
	docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} .

clean:
	rm -rf ./bin