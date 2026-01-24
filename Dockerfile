# syntax=docker/dockerfile:1.21
FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=dirty
ARG LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}"
ENV CGO_ENABLED=0

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="${LDFLAGS}" -o /out/never ./cmd/never/main.go

# Prepare world-writable dirs for arbitrary UID (OpenShift)
RUN mkdir -p /outfs/work /outfs/tmp && chmod 0777 /outfs/work /outfs/tmp

# Final: FROM scratch
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/never /never
COPY --from=builder /outfs/work /work
COPY --from=builder /outfs/tmp  /tmp
ENV HOME=/tmp
WORKDIR /work
USER 65532:65532
ENTRYPOINT ["/never"]



