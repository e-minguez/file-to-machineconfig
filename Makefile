# Go parameters
GOCMD=go
DEPCMD=dep
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=file-to-machineconfig
BINARY_UNIX=$(BINARY_NAME)-linux-amd64
BINARY_OSX=$(BINARY_NAME)-darwin-amd64
BINARY_WINDOWS=$(BINARY_NAME)-windows-amd64.exe

all: test build

build-all: build-linux build-osx build-windows

prepare:
	sudo dnf install -y golang git
	mkdir -p "${HOME}/go/bin"
	echo 'export GOPATH=${HOME}/go' >> ${HOME}/.bashrc
	echo 'export PATH=${GOPATH}/bin:${PATH}' >> ${HOME}/.bashrc

build: 
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_OSX)
	rm -f $(BINARY_UNIX).sha256
	rm -f $(BINARY_OSX).sha256

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

deps:
	$(DEPCMD) ensure -v

deps-update:
	$(DEPCMD) ensure -update -v

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
	sha256sum $(BINARY_UNIX) > $(BINARY_UNIX).sha256

build-osx:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_OSX) -v
	sha256sum $(BINARY_OSX) > $(BINARY_OSX).sha256

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_WINDOWS) -v
	sha256sum $(BINARY_WINDOWS) > $(BINARY_WINDOWS).sha256