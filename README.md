# phpunit-parallel

[![CI](https://github.com/alexdempster44/phpunit-parallel/actions/workflows/ci.yml/badge.svg)](https://github.com/alexdempster44/phpunit-parallel/actions/workflows/ci.yml)
[![CD](https://github.com/alexdempster44/phpunit-parallel/actions/workflows/cd.yml/badge.svg)](https://github.com/alexdempster44/phpunit-parallel/actions/workflows/cd.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexdempster44/phpunit-parallel)](https://goreportcard.com/report/github.com/alexdempster44/phpunit-parallel)

A CLI tool to run PHPUnit tests in parallel, with a beautiful terminal UI.

## Features

- Run PHPUnit tests in parallel across multiple workers
- Beautiful terminal UI with real-time progress
- TeamCity output format support for CI integration
- Automatic test distribution across workers
- Configurable number of parallel workers (defaults to CPU count)

## Installation

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

# Use TeamCity output format
phpunit-parallel --teamcity
```

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
