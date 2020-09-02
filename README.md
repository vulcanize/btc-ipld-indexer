# ipld-btc-indexer

[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanize/ipld-btc-indexer)](https://goreportcard.com/report/github.com/vulcanize/ipld-btc-indexer)

>  ipld-btc-indexer is used to extract, transform, and load all Bitcoin IPLD data into an IPFS-backing Postgres datastore while generating useful secondary indexes around the data in other Postgres tables

## Table of Contents
1. [Background](#background)
1. [Install](#install)
1. [Usage](#usage)
1. [Contributing](#contributing)
1. [License](#license)

## Background
ipld-btc-indexer is a collection of interfaces that are used to extract, transform, store, and index
all Bitcoin IPLD data in Postgres. The raw data indexed by ipld-btc-indexer serves as the basis for more specific watchers and applications.

## Dependencies
Minimal build dependencies
* Go (1.13)
* Git
* GCC compiler
* This repository

Potential external dependencies
* Goose
* Postgres
* Bitcoin full node

## Install
1. [Goose](#goose)
1. [Postgres](#postgres)
1. [Bitcoin](#bitcoin)
1. [Indexer](#indexer)

### Goose
[goose](https://github.com/pressly/goose) is used for migration management. While it is not necessary to use `goose` for manual setup, it
is required for running the automated tests and is used by the `make migrate` command.

### Postgres
1. [Install Postgres](https://wiki.postgresql.org/wiki/Detailed_installation_guides)
1. Create a superuser for yourself and make sure `psql --list` works without prompting for a password.
1. `createdb vulcanize_public`
1. `cd $GOPATH/src/github.com/vulcanize/ipld-btc-indexer`
1.  Run the migrations: `make migrate HOST_NAME=localhost NAME=vulcanize_public PORT=5432`
    - There are optional vars `USER=username:password` if the database user is not the default user `postgres` and/or a password is present
    - To rollback a single step: `make rollback NAME=vulcanize_public`
    - To rollback to a certain migration: `make rollback_to MIGRATION=n NAME=vulcanize_public`
    - To see status of migrations: `make migration_status NAME=vulcanize_public`

    * See below for configuring additional environments
    
In some cases (such as recent Ubuntu systems), it may be necessary to overcome failures of password authentication from
localhost. To allow access on Ubuntu, set localhost connections via hostname, ipv4, and ipv6 from peer/md5 to trust in: /etc/postgresql/<version>/pg_hba.conf

(It should be noted that trusted auth should only be enabled on systems without sensitive data in them: development and local test databases)

### Bitcoin
ipld-btc-indexer operates through the universally exposed bitcoin JSON-RPC interfaces.
Any of the standard full nodes can be used (e.g. bitcoind, btcd) as the data source.

Point at a remote node or set one up locally using the instructions for [bitcoind](https://github.com/bitcoin/bitcoin) and [btcd](https://github.com/btcsuite/btcd).

The default http url is "127.0.0.1:8332". We will use the http endpoint as both the `bitcoin.wsPath` and `bitcoin.httpPath`
(bitcoind does not support websocket endpoints, the watcher currently uses a "subscription" wrapper around the http endpoints)

### Indexer
Finally, setup the indexer process itself.

Start by downloading ipld-btc-indexer and moving into the repo:

`GO111MODULE=off go get -d github.com/vulcanize/ipld-btc-indexer`

`cd $GOPATH/src/github.com/vulcanize/ipld-btc-indexer`

Then, build the binary:

`make build`

## Usage
After building the binary, three commands are available

* Sync: Streams raw chain data at the head, transforms it into IPLD objects, and indexes the resulting set of CIDs in Postgres with useful metadata.

`./ipld-btc-indexer sync --config=<the name of your config file.toml>`

* Backfill: Automatically searches for and detects gaps in the DB; syncs the data to fill these gaps.

`./ipld-btc-indexer backfill --config=<the name of your config file.toml>`

* Resync: Manually define block ranges within which to (re)fill data over HTTP; can be ran in parallel with non-overlapping regions to scale historical data processing

`./ipld-btc-indexer resync --config=<the name of your config file.toml>`


### Configuration

Below is the set of parameters for the ipld-btc-indexer command, in .toml form, with the respective environmental variables commented to the side.
The corresponding CLI flags can be found with the `./ipld-btc-indexer {command} --help` command.

```toml
[database]
    name     = "vulcanize_public" # $DATABASE_NAME
    hostname = "localhost" # $DATABASE_HOSTNAME
    port     = 5432 # $DATABASE_PORT
    user     = "postgres" # $DATABASE_USER
    password = "" # $DATABASE_PASSWORD

[log]
    level = "info" # $LOGRUS_LEVEL

[sync]
    workers = 4 # $SYNC_WORKERS

[backfill]
    frequency = 15 # $BACKFILL_FREQUENCY
    batchSize = 2 # $BACKFILL_BATCH_SIZE
    workers = 4 # $BACKFILL_WORKERS
    timeout = 300 # $HTTP_TIMEOUT
    validationLevel = 1 # $BACKFILL_VALIDATION_LEVEL

[resync]
    type = "full" # $RESYNC_TYPE
    start = 0 # $RESYNC_START
    stop = 0 # $RESYNC_STOP
    batchSize = 2 # $RESYNC_BATCH_SIZE
    workers = 4 # $RESYNC_WORKERS
    timeout = 300 # $HTTP_TIMEOUT
    clearOldCache = false # $RESYNC_CLEAR_OLD_CACHE
    resetValidation = false # $RESYNC_RESET_VALIDATION

[bitcoin]
    wsPath  = "127.0.0.1:8332" # $BTC_WS_PATH
    httpPath = "127.0.0.1:8332" # $BTC_HTTP_PATH
    pass = "password" # $BTC_NODE_PASSWORD
    user = "username" # $BTC_NODE_USER
    nodeID = "ocd0" # $BTC_NODE_ID
    clientName = "Omnicore" # $BTC_CLIENT_NAME
    genesisBlock = "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f" # $BTC_GENESIS_BLOCK
    networkID = "0xD9B4BEF9" # $BTC_NETWORK_ID
```

`sync`, `backfill`, and `resync` parameters are only applicable to their respective commands.

`backfill` and `resync` require only an `bitcoin.httpPath` while `sync` requires only an `bitcoin.wsPath`.

### Exposing the data
* Use [ipld-btc-server](https://github.com/vulcanize/ipld-btc-server) to expose standard btc JSON RPC endpoints as well as unique ones
* Use [Postgraphile](https://www.graphile.org/postgraphile/) to expose GraphQL endpoints on top of the Postgres tables

e.g.

`postgraphile --plugins @graphile/pg-pubsub --subscriptions --simple-subscriptions -c postgres://localhost:5432/vulcanize_public?sslmode=disable -s public,btc -a -j`


This will stand up a Postgraphile server on the public and btc schemas- exposing GraphQL endpoints for all of the tables contained under those schemas.
All of their data can then be queried with standard [GraphQL](https://graphql.org) queries.

* Use PG-IPFS to expose the raw IPLD data. More information on how to stand up an IPFS node on top
of Postgres can be found [here](./documentation/ipfs.md)

### Testing
`make test` will run the unit tests  
`make test` setups a clean `vulcanize_testing` db

## Contributing
Contributions are welcome!

VulcanizeDB follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/1/4/code-of-conduct).

## License
[AGPL-3.0](LICENSE) © Vulcanize Inc

