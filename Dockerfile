# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w" \
      -trimpath \
      -o /server \
      ./cmd/server

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /server /server

EXPOSE 80 2002
USER nonroot:nonroot

ENTRYPOINT ["/server"]
