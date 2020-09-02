FROM golang:1.13-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers
# DEBUG
RUN apk add busybox-extras

# Get and build ipld-btc-indexer
ADD . /go/src/github.com/vulcanize/ipld-btc-indexer
WORKDIR /go/src/github.com/vulcanize/ipld-btc-indexer
RUN GO111MODULE=on GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o ipld-btc-indexer .

# Build migration tool
WORKDIR /
RUN go get -u -d github.com/pressly/goose/cmd/goose
WORKDIR /go/src/github.com/pressly/goose/cmd/goose
RUN GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -tags='no_mysql no_sqlite' -o goose .

WORKDIR /go/src/github.com/vulcanize/ipld-btc-indexer

# app container
FROM alpine

ARG USER
ARG CONFIG_FILE
ARG EXPOSE_PORT_1
ARG EXPOSE_PORT_2

RUN adduser -Du 5000 $USER
WORKDIR /app
RUN chown $USER /app
USER $USER

# chown first so dir is writable
# note: using $USER is merged, but not in the stable release yet
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/ipld-btc-indexer/$CONFIG_FILE config.toml
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/ipld-btc-indexer/startup_script.sh .

# keep binaries immutable
COPY --from=builder /go/src/github.com/vulcanize/ipld-btc-indexer/ipld-btc-indexer ipld-btc-indexer
COPY --from=builder /go/src/github.com/pressly/goose/cmd/goose/goose goose
COPY --from=builder /go/src/github.com/vulcanize/ipld-btc-indexer/db/migrations migrations/vulcanizedb
COPY --from=builder /go/src/github.com/vulcanize/ipld-btc-indexer/environments environments

EXPOSE $EXPOSE_PORT_1
EXPOSE $EXPOSE_PORT_2

ENTRYPOINT ["/app/startup_script.sh"]
