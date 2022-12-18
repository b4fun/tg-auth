FROM golang:1.19 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy source files
COPY exports.go exports.go
COPY cmd/ cmd/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o server cmd/server/*.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workspace/server .
USER nonroot:nonroot

ENTRYPOINT [ "/server" ]