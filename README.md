# gotestchunk

`gotestchunk` is a command-line interface (CLI) tool written in Go. It is designed to enumerate Go tests in a Go package and split them into chunks. This is particularly useful for CI systems that support parallelism, as it allows for more efficient test execution by distributing the load across multiple shards.

The tool outputs test results in Go's JSON test format, making it compatible with various test output formatters such as [gotestsum](https://github.com/gotestyourself/gotestsum), [gotestfmt](https://github.com/gotestyourself/gotestfmt), [tparse](https://github.com/mfridman/tparse), and [go-junit-report](https://github.com/jstemmer/go-junit-report).

## Installation

To install `gotestchunk`, you can use the `go install` command:

```sh
go install github.com/lox/gotestchunk@latest
```

## Usage

### Running Tests

Basic usage:

```sh
# Run chunk 2 of 4 tests
gotestchunk test --chunks=4 --chunk=2 ./pkg/...

# Pass build tags to go test
gotestchunk test --chunks=4 --chunk=2 ./pkg/... -- -tags=integration,e2e

# Pass timeout and other flags
gotestchunk test --chunks=4 --chunk=2 ./pkg/... -- -timeout=10m -count=1

# Run with verbose output
gotestchunk test -v --chunks=4 --chunk=2 ./pkg/...
```

### Test Output Formatting

gotestchunk outputs test results in Go's JSON test format, which is compatible with various test output formatters. Here are some popular options:

#### gotestsum

```sh
gotestchunk test --chunks=4 --chunk=2 ./pkg/... | gotestsum
```

#### gotestfmt

```sh
gotestchunk test --chunks=4 --chunk=2 ./pkg/... | gotestfmt
```

#### tparse

```sh
gotestchunk test --chunks=4 --chunk=2 ./pkg/... | tparse
```

#### go-junit-report

```sh
gotestchunk test --chunks=4 --chunk=2 ./pkg/... | go-junit-report > report.xml
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
gotestchunk test --packages ./pkg/... -- -tags=integration
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
gotestchunk test --packages ./pkg/... -- -tags=integration --timing-file=timing.json
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
