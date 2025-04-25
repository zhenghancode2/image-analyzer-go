APP_NAME := image-analyzer
BUILD_DIR := build
TAG := latest
DOCKERFILE := Dockerfile

.PHONY: all build run clean docker

all: build

build:
	go mod tidy
	go build -o $(BUILD_DIR)/$(APP_NAME) main.go

run: build
ifeq ($(MODE),analyze)
	./$(BUILD_DIR)/$(APP_NAME) analyze $(IMAGE)
else
	./$(BUILD_DIR)/$(APP_NAME) server
endif

clean:
	rm -rf $(BUILD_DIR)

docker:
	docker build -t $(APP_NAME):$(TAG) -f $(DOCKERFILE) .
