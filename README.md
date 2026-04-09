# phpunit-parallel

[![License](https://img.shields.io/github/license/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel/releases)
[![Homebrew](https://img.shields.io/badge/homebrew-alexdempster44%2Ftap-orange)](https://github.com/alexdempster44/homebrew-tap)
[![Go Version](https://img.shields.io/github/go-mod/go-version/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel)
[![CI](https://img.shields.io/github/actions/workflow/status/alexdempster44/phpunit-parallel/ci.yml?label=CI&logo=github)](https://github.com/alexdempster44/phpunit-parallel/actions/workflows/ci.yml)
[![CD](https://img.shields.io/github/actions/workflow/status/alexdempster44/phpunit-parallel/cd.yml?label=CD&logo=github)](https://github.com/alexdempster44/phpunit-parallel/actions/workflows/cd.yml)
[![GitHub issues](https://img.shields.io/github/issues/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel/issues)
[![GitHub pulls](https://img.shields.io/github/issues-pr/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel/pulls)

A CLI tool to run PHPUnit tests in parallel, with a beautiful terminal UI.

## Features

- **Parallel test execution** - Distributes tests across multiple workers using round-robin scheduling
- **Interactive terminal UI** - Real-time progress bars, per-worker status, and ETA powered by Bubble Tea
- **Rerun failed tests** - Press `r` to rerun only failures, or `a` to rerun all tests, without leaving the UI
- **Deprecation tracking** - Parses and displays PHPUnit deprecation warnings with source locations
- **Lifecycle hooks** - Run custom commands before/after test runs and before/after each worker
- **Docker support** - Run each worker in its own container using hooks and the `{}` placeholder
- **TeamCity output** - Native TeamCity format for CI integration (JetBrains TeamCity, GitHub Actions, etc.)
- **Runner configuration file** - Define reusable settings in `phpunit-parallel.xml`
- **PHPUnit compatible** - Reads your existing `phpunit.xml` / `phpunit.xml.dist` and preserves all settings
- **Test filtering** - Filter tests with `--filter`, `--group`, and `--exclude-group`
- **Copy errors to clipboard** - Copy individual or all errors/deprecations with keyboard shortcuts
- **IDE integration** - Included PHP wrapper script for PhpStorm interop
- **Cross-platform** - Pre-built binaries for Linux and macOS (amd64/arm64)

## Installation

### Using Homebrew (Recommended)

```bash
brew install alexdempster44/tap/phpunit-parallel
```

### Using Go

```bash
go install github.com/alexdempster44/phpunit-parallel@latest
```

### From Releases

Download the pre-built binary for your platform from the [releases page](https://github.com/alexdempster44/phpunit-parallel/releases).

## Usage

```bash
# Run with default settings (uses phpunit.xml in current directory)
phpunit-parallel

# Specify number of workers
phpunit-parallel -w 4

# Specify PHPUnit configuration file
phpunit-parallel -c phpunit.xml.dist

# Filter tests
phpunit-parallel --filter testLogin
phpunit-parallel --group integration
phpunit-parallel --exclude-group slow

# Use TeamCity output format (for CI)
phpunit-parallel --teamcity
```

### CLI Flags

| Flag                 | Short | Default              | Description                                        |
| -------------------- | ----- | -------------------- | -------------------------------------------------- |
| `--workers`          | `-w`  | CPU count - 2        | Number of parallel workers                         |
| `--configuration`    | `-c`  | `phpunit.xml`        | PHPUnit configuration file                         |
| `--runner-config`    |       |                      | Runner configuration file (`phpunit-parallel.xml`) |
| `--teamcity`         |       | `false`              | Output in TeamCity format                          |
| `--filter`           |       |                      | Filter which tests to run (passed to PHPUnit)      |
| `--group`            |       |                      | Only run tests from the specified group(s)         |
| `--exclude-group`    |       |                      | Exclude tests from the specified group(s)          |
| `--test-suffix`      |       | `Test.php`           | File suffix for test discovery                     |
| `--config-build-dir` |       | `.phpunit-parallel`  | Directory for generated per-worker configs         |
| `--before`           |       |                      | Command to run once before all workers start       |
| `--before-worker`    |       |                      | Command to run before each worker starts           |
| `--run-worker`       |       | `vendor/bin/phpunit` | Command to run PHPUnit for each worker             |
| `--after-worker`     |       |                      | Command to run after each worker completes         |
| `--after`            |       |                      | Command to run once after all workers complete     |
| `--version`          |       |                      | Display version                                    |

### Environment Variables

| Variable                   | Description                                                        |
| -------------------------- | ------------------------------------------------------------------ |
| `PHPUNIT_PARALLEL_WORKERS` | Override the number of workers (lower precedence than `--workers`) |

The following environment variables are exposed to hooks and workers:

| Variable       | Description                                              |
| -------------- | -------------------------------------------------------- |
| `PARALLEL`     | Always `1` when running under phpunit-parallel           |
| `PROJECT`      | Project directory name                                   |
| `RUNNER_PID`   | PID of the runner process                                |
| `WORKER_ID`    | Worker identifier (per-worker hooks and run-worker only) |
| `WORKER_COUNT` | Total number of workers                                  |

## Terminal UI

The interactive TUI features a three-panel layout:

- **Top-left** - Running tests panel (switches to a summary after completion)
- **Bottom-left** - Per-worker progress bars with percentage and test counts
- **Right** - Error and deprecation panel with expandable details

### Keyboard Shortcuts

| Key               | Action                                     |
| ----------------- | ------------------------------------------ |
| `Tab`             | Switch between panels                      |
| `Up` / `k`        | Navigate up                                |
| `Down` / `j`      | Navigate down                              |
| `Enter` / `Space` | Toggle expand/collapse                     |
| `Right` / `l`     | Expand error details                       |
| `Left` / `h`      | Collapse error details                     |
| `PgUp` / `PgDn`   | Page up/down                               |
| `c`               | Copy selected error to clipboard           |
| `C`               | Copy all errors/deprecations (opens modal) |
| `r`               | Rerun failed tests (after completion)      |
| `a`               | Rerun all tests (after completion)         |
| `q` / `Ctrl+C`    | Quit                                       |

When piped or redirected, output falls back to a plain-text format automatically.

## Runner Configuration File

For reusable or complex setups, create a `phpunit-parallel.xml` file:

```xml
<runner>
    <workers>4</workers>
    <configuration>phpunit.xml</configuration>
    <config-build-dir>.phpunit-parallel</config-build-dir>
    <test-suffix>Test.php</test-suffix>
    <before>echo "Starting test run"</before>
    <before-worker>echo "Setting up worker $WORKER_ID"</before-worker>
    <run-worker>vendor/bin/phpunit</run-worker>
    <after-worker>echo "Cleaning up worker $WORKER_ID"</after-worker>
    <after>echo "Test run complete"</after>
</runner>
```

Use it with:

```bash
phpunit-parallel --runner-config phpunit-parallel.xml
```

If `phpunit-parallel.xml` exists in the current directory, it is loaded automatically.

### Docker Example

Run each worker in its own Docker container:

```xml
<runner>
    <workers>4</workers>
    <before-worker>docker run -d --name "phpunit-worker-$WORKER_ID" -v "$(pwd)":/app -w /app php:8.3-cli sleep infinity</before-worker>
    <run-worker>docker exec "phpunit-worker-$WORKER_ID" php vendor/bin/phpunit {}</run-worker>
    <after-worker>docker rm -f "phpunit-worker-$WORKER_ID"</after-worker>
</runner>
```

The `{}` placeholder in `run-worker` is replaced with the generated PHPUnit arguments (configuration path, flags, filters, etc.).

### Docker Compose Example

For projects using Docker Compose, use a unique project name per worker to isolate containers. Include `--rmi local` in the `after-worker` hook to remove locally-built images — without it, each run leaves behind orphaned images that accumulate over time:

```xml
<runner>
    <workers>4</workers>
    <before>docker compose -f docker-compose.yml -f docker-compose.testing.yml build testing</before>
    <run-worker>docker compose -f docker-compose.yml -f docker-compose.testing.yml -p $PROJECT-test-$RUNNER_PID-$WORKER_ID run --rm testing php vendor/bin/phpunit {}</run-worker>
    <after-worker>docker compose -f docker-compose.yml -f docker-compose.testing.yml -p $PROJECT-test-$RUNNER_PID-$WORKER_ID down -v --remove-orphans --rmi local</after-worker>
</runner>
```

## Output Modes

| Mode                | When                    | Description                                                 |
| ------------------- | ----------------------- | ----------------------------------------------------------- |
| **Interactive TUI** | Terminal (default)      | Full UI with progress bars, panels, and keyboard navigation |
| **Plain text**      | Piped/redirected output | Structured `[worker N]` and `[summary]` lines               |
| **TeamCity**        | `--teamcity` flag       | TeamCity service messages for CI integration                |

## Building from Source

```bash
# Clone the repository
git clone https://github.com/alexdempster44/phpunit-parallel.git
cd phpunit-parallel

# Build
go build -o bin/phpunit-parallel .

# Or using just
just build
```

## License

This is free and unencumbered software released into the public domain. See the [LICENSE](LICENSE) file for details.
