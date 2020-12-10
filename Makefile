# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Update gomod
update-gomod:
	go mod download
	go mod tidy 
	go mod verify

# Build this
build: #test
	go build -a -o listener *.go

test:
	go test ./...