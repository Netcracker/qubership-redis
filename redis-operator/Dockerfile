FROM golang:1.24.6-alpine AS builder
ARG GOPROXY=https://proxy.golang.org,direct
ENV CGO_ENABLED=0
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build \
      -o bin/dbaas-redis-operator \
      -trimpath \
      -ldflags "-s -w" \
      ./main.go

FROM alpine:3.22.0

ENV WORKDIR=/opt/operator/ \
    OPERATOR=/usr/local/bin/dbaas-redis-operator \
    USER_UID=1001 \
    USER_NAME=dbaas-redis-operator

WORKDIR ${WORKDIR}

COPY --from=builder /src/bin/dbaas-redis-operator ${OPERATOR}
COPY --from=builder /src/charts/helm/redis-operator/ deployments/charts/redis-operator/
COPY --from=builder /src/charts/helm/redis-operator/ helm-templates/redis-operator/

COPY build/bin /usr/local/bin

RUN apk update \
    && apk add --no-cache curl \
    && rm -rf /var/cache/apk/*

RUN chmod +x /usr/local/bin/entrypoint \
    && chmod +x /usr/local/bin/user_setup \
    && /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
