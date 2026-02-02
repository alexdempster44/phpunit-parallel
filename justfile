# Default recipe
default:
    @just --list

# Build the binary
build:
    go build -o bin/phpunit-parallel .

# Install the binary
install:
    go install .

# Run the program
run *ARGS:
    go run . {{ARGS}}
