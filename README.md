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
gotestchunk test -v --chunks=4 --chunk=2 --packages ./pkg/...
```

### Listing and Chunking Tests

The list command provides several ways to view and chunk tests:

```sh
# List all tests in a package
gotestchunk list ./pkg/example

# List tests in chunk 2 of 4
gotestchunk list --chunks=4 --chunk=2 ./pkg/...

# List package paths for tests
gotestchunk list --format=listPackages ./pkg/...

# Get test pattern for use with go test -run
gotestchunk list --format=runPattern ./pkg/...

# Get package paths for chunk 2 of 4
gotestchunk list --format=listPackages --chunks=4 --chunk=2 ./pkg/...
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
# List all tests in a package
$ gotestchunk list ./pkg/example
TestSimple
TestParallel
TestTableDriven
TestWithSetup

# List tests in chunk 1 of 2
$ gotestchunk list --chunks=2 --chunk=1 ./pkg/example
TestSimple
TestParallel

# List package paths
$ gotestchunk list --format=listPackages ./pkg/example/...
./pkg/example
./pkg/example/sub

# Get test pattern
$ gotestchunk list --format=runPattern ./pkg/example
^(TestSimple|TestParallel|TestTableDriven|TestWithSetup)$
```

### Test Timing Information

To collect test timing information, you can use the `--timing-file` flag:

```sh
# Collect test timing information
gotestchunk test --gotestsum --packages ./pkg/... -- -tags=integration --timing-file=timing.json
```

### Test Distribution with Timing Data

To improve test distribution across chunks, you can use historical timing data:

```sh
# First run - collect timing data
gotestchunk test --write-timing=timing-1.json ./pkg/...

# Use timing data to better distribute tests
gotestchunk test --read-timing="timing-*.json" --chunks=4 --chunk=1 ./pkg/...
```

The tool will:
1. Load and aggregate timing data from all matching files
2. Use the average test duration to distribute tests more evenly across chunks
3. Fall back to equal distribution if no timing data is available

## Features

- Splits tests into equal chunks for parallel execution
- Automatic CI environment detection
- Direct integration with gotestsum
- Supports recursive package listing with `...`
- Handles nested packages
- Shows parallel tests
- Includes tests with setup and cleanup
- Supports all go test flags via `--`
- Multiple output formats for test listing

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
