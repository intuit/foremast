# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/foremast.ai/foremast/foremast-trigger
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager foremast.ai/foremast/foremast-trigger/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:latest
RUN apt-get update
RUN apt-get -y install curl
WORKDIR /root/
COPY --from=builder /go/src/foremast.ai/foremast/foremast-trigger/cmd/manager/requests.csv .
COPY --from=builder /go/src/foremast.ai/foremast/foremast-trigger/manager .
# COPY ./cmd/manager/manager .
ENTRYPOINT ["./manager"]
