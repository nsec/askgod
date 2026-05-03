FROM golang:latest AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
RUN apt-get update && apt-get install -y make
COPY . .
RUN make linux

FROM debian:bookworm-slim
COPY --from=builder /src/bin/linux/askgod-server /askgod-server
ENTRYPOINT ["/askgod-server"]
