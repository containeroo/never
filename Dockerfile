# syntax=docker/dockerfile:1.25

# Build the manager binary.
FROM golang:1.26 AS prep

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=none
ARG LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}"

ENV CGO_ENABLED=0

WORKDIR /workspace

# Copy the Go module manifests first so dependency downloads can be cached.
COPY go.mod go.mod
COPY go.sum go.sum

# Download modules before copying source files so source changes do not
# invalidate the dependency cache layer.
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

# Copy the Go source and templates.
COPY cmd/ cmd/
COPY internal/ internal/

# Build the binary.
# TARGETARCH defaults to the builder architecture for regular Docker builds,
# but can be set by buildx for cross-platform builds.
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} \
  go build \
  -ldflags="$LDFLAGS" \
  -a \
  -o never \
  cmd/never/main.go

# Create writable runtime directories owned by the root group.
# The setgid bit keeps new files/directories in group 0, which supports
# OpenShift's arbitrary UID model while still running as a non-root user.
RUN install -d -o 0 -g 0 -m 2775 /outfs/work /outfs/tmp

# Use distroless as minimal base image to package the manager binary.
# Refer to https://github.com/GoogleContainerTools/distroless for more details.
FROM gcr.io/distroless/static:nonroot

COPY --from=prep /workspace/never /never
COPY --from=prep /outfs/work /work
COPY --from=prep /outfs/tmp /tmp

ENV HOME=/tmp
WORKDIR /work

# Run as a non-root user by default.
# Use GID 0 so the process can write to root-group-owned writable paths,
# which keeps the image compatible with OpenShift's arbitrary UID model.
USER 65532:0

ENTRYPOINT ["/never"]

