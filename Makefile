GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool

BINARY_NAME=golisp

all: test build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test:
	$(GOTEST) -v ./...
cover:
	$(GOTEST) -coverprofile=cover.out ./...
	$(GOTOOL) cover -html=cover.out -o cover.html
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
