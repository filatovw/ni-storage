# ni-storage

Read more about the [challenge](./docs/challenge.md).

# Why Golang

I chose Golang because:
- there are many concurrent operations in this application and Golang works perfectly in this area
- the result of a build is a single binary, so it is easy to distribute
- perfect tooling (race detector, profiling tools)
- perfect testing abilities from the box (unittests, benchmarks, property-based testing)
- code is easy to read

## TODO:

- web server with basic handlers
- `/debug` handler
- `/metrics` handler
- engine adapter 
- simple persistent storage
- third-party storage (pg/redis/badger)
- documentate solution