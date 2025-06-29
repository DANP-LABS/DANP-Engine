# DANP Engine Makefile
PROG_CLIENT=bin/DANP-MCP-CLIENT
PROG_SERVER=bin/DANP-MCP-SERVER
SRCS_CLIENT=./cmd/DANP-MCP-CLIENT
SRCS_SERVER=./cmd/DANP-MCP-SERVER

# Version info
COMMIT_HASH=$(shell git rev-parse --short HEAD || echo "GitNotFound")
BUILD_DATE=$(shell date -u '+%Y-%m-%d %H:%M:%S')
BUILD_FLAGS=-ldflags "-s -w -X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\""

# Default target
all: build-client build-server

# Create bin directory
$(shell mkdir -p bin)

# Build targets
build-client:
	go build ${BUILD_FLAGS} -o ${PROG_CLIENT} ${SRCS_CLIENT}

build-server:
	go build ${BUILD_FLAGS} -o ${PROG_SERVER} ${SRCS_SERVER}

# Cross-compilation targets
build-all: build-linux build-windows build-darwin build-arm

build-linux:
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-linux ${SRCS_CLIENT}
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-linux ${SRCS_SERVER}

build-windows:
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-windows.exe ${SRCS_CLIENT}
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-windows.exe ${SRCS_SERVER}

build-darwin:
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-darwin ${SRCS_CLIENT}
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-darwin ${SRCS_SERVER}

build-arm:
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-arm ${SRCS_CLIENT}
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-arm ${SRCS_SERVER}

# Development targets
run-client:
	go run ${SRCS_CLIENT}

run-server:
	go run ${SRCS_SERVER}

# Cleanup
clean:
	rm -f ${PROG_CLIENT}* ${PROG_SERVER}* bin/*.exe

.PHONY: all build-client build-server build-all build-linux build-windows build-darwin build-arm run-client run-server clean
