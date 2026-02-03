# Multi-stage: build Go + static, run single binary. See docs/plan-from-scratch.md.
# Build from repo root: docker build -f docker/Dockerfile .
FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 go build -o /studio-photos ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=build /studio-photos .
EXPOSE 8080
ENTRYPOINT ["./studio-photos"]
