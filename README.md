# gotestchunk

`gotestchunk` is a command-line interface (CLI) tool written in Go. It is designed to enumerate Go tests in a Go package and split them into chunks. This is particularly useful for CI systems that support parallelism, as it allows for more efficient test execution by distributing the load across multiple shards.

It is designed to integrate neatly with [gotestsum](https://github.com/gotestyourself/gotestsum).

## Installation

To install `gotestchunk`, you can use the `go install` command:

```sh
go install github.com/lox/gotestchunk@latest
```

## Usage

### Running Tests

The simplest way to run chunked tests is with gotestsum:

```sh
# Run chunk 2 of 4 tests with gotestsum
gotestchunk test --gotestsum --chunks=4 --chunk=2 --packages ./pkg/...

# Pass build tags to go test
gotestchunk test --chunks=4 --chunk=2 --packages ./pkg/... -- -tags=integration,e2e

# Pass timeout and other flags
gotestchunk test --chunks=4 --chunk=2 --packages ./pkg/... -- -timeout=10m -count=1

# Run with verbose output
gotestchunk test  -v --chunks=4 --chunk=2 --packages ./pkg/...
```

### Manual Chunking

For more control, you can use the chunking commands separately:

```sh
# Get test pattern for chunk 2 of 4
gotestchunk chunk --chunks=4 --chunk=2 ./pkg/...

# Get package paths for chunk 2 of 4
gotestchunk chunk-packages --chunks=4 --chunk=2 ./pkg/...

# List all tests in a package
gotestchunk list ./pkg/example
```

### CI Environment Support

The tool automatically detects CI environments and their parallelism settings:

- GitLab CI / Knapsack / TravisCI: `CI_NODE_INDEX` / `CI_NODE_TOTAL`
- CircleCI: `CIRCLE_NODE_INDEX` / `CIRCLE_NODE_TOTAL`
- Bitbucket Pipelines: `BITBUCKET_PARALLEL_STEP` / `BITBUCKET_PARALLEL_STEP_COUNT`
- Buildkite: `BUILDKITE_PARALLEL_JOB` / `BUILDKITE_PARALLEL_JOB_COUNT`
- Semaphore: `SEMAPHORE_CURRENT_JOB` / `SEMAPHORE_JOB_COUNT`

When these variables are present, you can omit the `--chunks` and `--chunk` flags:

```sh
# Uses CI environment variables for chunking
gotestchunk test --gotestsum --packages ./pkg/... -- -tags=integration
```

### Example Output

```sh
# Running tests for chunk 2 of 4 with gotestsum
$ gotestchunk test --gotestsum --chunks=4 --chunk=2 --packages ./pkg/...
=== RUN   TestSimple
--- PASS: TestSimple (0.00s)
=== RUN   TestParallel
--- PASS: TestParallel (0.20s)
DONE 2 tests in 0.20s
```

## Features

- Splits tests into equal chunks for parallel execution
- Automatic CI environment detection
- Direct integration with gotestsum
- Supports recursive package listing with `...`
- Handles nested packages
- Shows parallel tests
- Includes tests with setup and cleanup
- Supports all go test flags via `--`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
