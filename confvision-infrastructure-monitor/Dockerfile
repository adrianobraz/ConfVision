FROM golang:1.25-bookworm AS build

WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/monitor ./cmd/monitor

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/monitor /monitor

ENV MONITOR_BEHIND_PROXY=1
ENV MONITOR_LISTEN=0.0.0.0:8090
ENV MONITOR_HOST_ROOT=/host

EXPOSE 8090

ENTRYPOINT ["/monitor"]
