# Build the manager binary
FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY fs.go fs.go
COPY handler.go handler.go
COPY main.go main.go
COPY serv.go serv.go
COPY texts.go texts.go
COPY types.go types.go
COPY listenerserv/ listenerserv/

# Build
# NOTE: since this is going to run on Raspberry PI, we build this on ARM
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 GO111MODULE=on go build -a -o listener *.go
RUN chmod +x listener

# Use distroless as minimal base image to package the program binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/listener .
USER nonroot:nonroot

ENV MODE=docker

ENTRYPOINT ["/listener"]