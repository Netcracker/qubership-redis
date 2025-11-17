FROM golang:1.24.6-alpine AS builder

ARG GOPATH="/home/netcrk/go"
ENV GOPATH=${GOPATH}

WORKDIR /src

COPY source/go.mod source/go.sum ./
RUN go mod download

COPY source ./
RUN mkdir -p bin \
  && go build \
       -o ./bin/monitoring-watcher \
       -gcflags "all=-trimpath=${GOPATH}" \
       -asmflags "all=-trimpath=${GOPATH}"

FROM redis:7.4.7-alpine AS redis
FROM telegraf:1.36.2-alpine

USER root
COPY health.sh /health.sh
COPY get_metrics.sh /opt/telegraf/scripts/
RUN chown 1001:1001 /opt/telegraf/scripts/get_metrics.sh

#timeout tool has issues with alpine 10+
#fix for "timeout: unrecognized option: t"
RUN apk update \
    && apk add --no-cache coreutils tini gcompat \
    && apk upgrade --no-cache

COPY --from=redis /usr/local/bin/redis-server /usr/bin/redis-server
COPY --from=redis /usr/local/bin/redis-cli /usr/bin/redis-cli

# Fixed "not found" issue
# ldd /bin/...
#RUN ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY --from=builder /src/bin/monitoring-watcher /bin/monitoring-watcher
USER 1001
ENTRYPOINT /bin/monitoring-watcher

#ENTRYPOINT ["/sbin/tini", "--"]
#CMD ["telegraf"]
