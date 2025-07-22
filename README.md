# Cache Service

As per the task this project implements a simple in-memory cache service with an HTTP interface. Items are stored for 30 minutes by default and the service is designed to handle a high volume of requests by sharding the cache and supporting concurrent access. The TTL can be modified through an env variable `CACHE_TTL`. There are some other env variable that can control various configuration to run the cache. There is a docker-compose.yaml file provided that demonstrates how this env variables can be passed. 

## Building

```
go build -o bin/cache-service ./cmd/cache-service
```

or 
```
make build
```
This produces a binary in the bin directory (in the project root) called cache-service.

## Running 

### Running directly:
Run the compiled binary and specify the port to expose:

```
PORT=8080 ./cache-service
```

Alternatively it can be run directly from source:

```
go run ./cmd/cache-service
```

or 

```
make run
```
### Running in docker:
We can also run the project using the provided docker-compose file. 

Running `docker compose up` command in the root directory of the project will start the docker container at port 8080 and then the service can be accessed through http://localhost:8080

Alternatively we can also run the project using make file convinience commands.
- The `make up` command builds and starts the cache service container and exposes it on port 8080 
- The `make down` command stops the container and thus the service.

Once the project is running we can run the example HTTP requests are included in `requests/test_api.http` for use with tools like the JetBrains HTTP client.

## Testing
Run unit tests for all packages:

```
go test ./...
```

We can use the included Makefile to run tests with additional options:

```
make test       # run tests
make test-race  # run tests with the race detector
make coverage   # generate coverage report
```

The service's HTTP interface can also be exercised manually:

```
curl -X POST http://localhost:8080/api/v1/cache/mykey -d 'some value'
curl http://localhost:8080/api/v1/cache/mykey
```
You can import `requests/test_api.http` into JetBrains IDEs or other clients to run the same requests.

Note: The port 8080 is just for demonstration (it can be set as an env variable) and any valid port can be used.

## Performance

Benchmarks were run with Go 1.24.3. Results:

```
$ go test -bench . ./internal/cache ./internal/evictors ./internal/server
BenchmarkCacheReadWrite-5                4827630               244.6 ns/op
BenchmarkCacheSet-5                      1000000              1544 ns/op
BenchmarkCacheGetMiss-5                  2222994               459.4 ns/op
BenchmarkCacheParallelReads-5            5918631               240.0 ns/op
BenchmarkCacheParallelWrites-5           1000000              2101 ns/op
BenchmarkLRUEvictorSet-5         1659502               798.3 ns/op
BenchmarkLRUEvictorEvict-5      68147101                16.44 ns/op
BenchmarkHandleGet-5      438638              2449 ns/op
BenchmarkHandleSet-5      257946              4896 ns/op
```

These numbers show that read operations take only a few hundred nanoseconds and writes complete in a few microseconds. Parallel benchmarks demonstrate the cache can sustain millions of operations per second while the LRU evictor keeps eviction overhead extremely low (~16ns per call).


## Areas of improvement

- Some TODO comments in the code highligh the improvements that could be made in those places. 
For example when trying to get a shard we fallback to the first shard, this is not a good appraoch and may result in hotspots.

- The main function in the main.go file needs to be refactored to make it cleaner. Right now it has too many responsibilities.

- The docs folder in the internal folder, not a good place to keep it there. Ideally it should be in the root directory. 

- Tests only cover about 70% of the code and miss out some error paths that might have critical bugs and race conditions. 

