#!/bin/bash
# Quick test runner for API Gateway
# Usage: ./run_tests.sh [option]
# Options: all, service, server, redis, coverage

set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

run_all_tests() {
    print_header "Running All Tests"
    go test -v ./...
}

run_service_tests() {
    print_header "Running Service Layer Tests"
    go test -v ./internal/service
}

run_server_tests() {
    print_header "Running gRPC Server Tests"
    go test -v ./internal/grpc/server
}

run_redis_tests() {
    print_header "Running Redis Storage Tests"
    go test -v ./internal/storage/redis
}

run_coverage() {
    print_header "Running Tests with Coverage"
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    echo ""
    echo -e "${GREEN}Coverage report generated: coverage.out${NC}"
    echo -e "${GREEN}Open HTML report: go tool cover -html=coverage.out${NC}"
}

case "${1:-all}" in
    all)
        run_all_tests
        ;;
    service)
        run_service_tests
        ;;
    server)
        run_server_tests
        ;;
    redis)
        run_redis_tests
        ;;
    coverage)
        run_coverage
        ;;
    *)
        echo "Unknown option: $1"
        echo "Usage: $0 [all|service|server|redis|coverage]"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}✓ Tests completed successfully!${NC}"
