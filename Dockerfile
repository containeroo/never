FROM golang:1.24-alpine AS builder
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -installsuffix nocgo -o /never ./cmd/never

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /never /
USER nonroot:nonroot
ENTRYPOINT ["/never"]
