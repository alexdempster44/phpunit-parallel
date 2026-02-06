# phpunit-parallel

[![License](https://img.shields.io/github/license/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel)
[![PHP Version](https://img.shields.io/packagist/dependency-v/alexdempster44/phpunit-parallel/php)](https://packagist.org/packages/alexdempster44/phpunit-parallel)
[![CI](https://img.shields.io/github/actions/workflow/status/alexdempster44/phpunit-parallel/ci.yml?label=CI&logo=github)](https://github.com/alexdempster44/phpunit-parallel/actions/workflows/ci.yml)
[![CD](https://img.shields.io/github/actions/workflow/status/alexdempster44/phpunit-parallel/cd.yml?label=CD&logo=github)](https://github.com/alexdempster44/phpunit-parallel/actions/workflows/cd.yml)
[![GitHub issues](https://img.shields.io/github/issues/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel/issues)
[![GitHub pulls](https://img.shields.io/github/issues-pr/alexdempster44/phpunit-parallel)](https://github.com/alexdempster44/phpunit-parallel/pulls)

A CLI tool to run PHPUnit tests in parallel, with a beautiful terminal UI.

## Features

- Run PHPUnit tests in parallel across multiple workers
- Beautiful terminal UI with real-time progress
- TeamCity output format support for CI integration
- Automatic test distribution across workers
- Configurable number of parallel workers (defaults to CPU count)

## Installation

### Using Composer (Recommended)

```bash
composer require --dev alexdempster44/phpunit-parallel
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
