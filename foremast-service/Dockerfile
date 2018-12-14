
# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/foremast.ai/foremast/foremast-service
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o service foremast.ai/foremast/foremast-service/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:latest
WORKDIR /app/
COPY --from=builder /go/src/foremast.ai/foremast/foremast-service/service .
ENTRYPOINT ["./service"]
