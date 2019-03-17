# ni-storage

Read more about the [challenge](./docs/challenge.md).

# Why Golang

I chose Golang because:
- there are many concurrent operations in this application and Golang works perfectly in this area
- the result of a build is a single binary, so it is easy to distribute
- perfect tooling (race detector, profiling tools)
- perfect testing abilities from the box
- code is easy to read

## Repository structure

`/api` here is an HTTP API server powered with Chi router 
`/bin` contains actual binary
`/bin/release` contains latest platform specific releases
`/cmd` place for commands that share same underlying code. For now there is just one command for starting storage with `http api` interface
`/config` object that reads configurations from environment variables and command line
`/data` place for a data storage
`/docs` description of a challenge
`/engine` engine interface
`/engine/narwal` engine implementation
`/logger` simple interface that is used for isolation from a particular logger


Requirements: `make`, `docker`, `docker-compose`, `go >= 1.12`, `git`

## HTTP API server

Solution compiles into a single binary `/bin/ni-storage`

    ./bin/ni-storage --help

    Usage of ./bin/ni-storage:
    -data-dir string
            path to folder with data (default "./data"), environment variable: NI_NARWAL_DATA_DIR
    -debug
            debug mode with verbose logging, environment variable: NI_DEBUG 
    -host string
            api-server host (default: 0.0.0.0), environment variable: NI_API_HOST 
    -port int
            api-server port (default: 8500), environment variable: NI_API_PORT 

This server also supports these handlers:

* `/health` for healthcheck
* `/debug` for golang profiler
* `/metrics` for prometheus metrics

Command line arguments have more priority than environment variables.

## Shortcuts
If you are docker user:

Build docker image:

    make docker-build

Start application in docker container:
    
    make docker-start

See docker logs:

    make docker-logs

Stop application:

    make docker-stop

Delete unused images:

    make docker-clean

Make release binaries for Darwin (amd64) and Linux (amd64)

    make release


For local development and usage:

Build:

    make build

Start application:

    make start

Launch tests:

    make test

## Architecture of a storage:

There are 2 main components:

1. Web server based on a standard server from `net/http`. Its router is not enough flexible, so I used router from `go-chi/chi`.
2. NarWAL storage. It has in-memory KV storage and stores all write/delete operations on a disk without any queues and buffers in a JSON format.
If I had more time I would add/change this things:
- records should be `msgpack`/`protobuf`-encoded.
- ability to setup interval for fsync
- log compaction (now it grows without limits)
- for now result of replaying of a log should fit into memory which is bad
- hashsum of each record should be placed in log in order to prevent issues with data corruption
- then I'd replace NarWAL with Redis (for WAL and snapshots) or Badger (for LSM-tree)


I also feel confusion about `PUT /keys` method because it is so unusual way for adding of a single value. I replaced it with `PUT /keys/{id}` because this fits better in my experience.


## Examples:

Create item:

    curl -X PUT "0.0.0.0:8555/keys/time" -H "content-type:application/json" -d "to die"
    "OK"

Create item with expiration time:

    curl -X PUT "0.0.0.0:8555/keys/bear?expire_in=25" -H "content-type:application/json" -d "polar"
    "OK"

Check item:

    curl --head "0.0.0.0:8555/keys/time" -H "content-type:application/json"
    HTTP/1.1 200 OK
    Content-Type: application/json; charset=utf-8
    Date: Sun, 17 Mar 2019 23:44:41 GMT
    Content-Length: 5

Get item:

    curl -X GET "0.0.0.0:8555/keys/time" -H "content-type:application/json"
    "to die"

Get all items:

    curl -X GET "0.0.0.0:8555/keys" -H "content-type:application/json"
    {"bear":{"expiration_time":"2019-03-17T23:43:00.394580122Z","value":"polar","key":"bear"},"time":{"value":"to die","key":"time"}}

Delete item:

    curl -X DELETE  "0.0.0.0:8555/keys/time" -H "content-type:application/json"
    "Accepted"

Delete all items:

    curl -X DELETE  "0.0.0.0:8555/keys" -H "content-type:application/json"
    "Accepted"