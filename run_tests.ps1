# Quick test runner for API Gateway (Windows)
# Usage: .\run_tests.ps1 -Target all|service|server|redis|coverage

param(
    [ValidateSet("all", "service", "server", "redis", "coverage")]
    [string]$Target = "all"
)

$ErrorActionPreference = "Stop"

function Print-Header {
    param([string]$Text)
    Write-Host "========================================" -ForegroundColor Blue
    Write-Host $Text -ForegroundColor Blue
    Write-Host "========================================" -ForegroundColor Blue
}

function Run-AllTests {
    Print-Header "Running All Tests"
    go test -v ./...
}

function Run-ServiceTests {
    Print-Header "Running Service Layer Tests"
    go test -v ./internal/service
}

function Run-ServerTests {
    Print-Header "Running gRPC Server Tests"
    go test -v ./internal/grpc/server
}

function Run-RedisTests {
    Print-Header "Running Redis Storage Tests"
    go test -v ./internal/storage/redis
}

function Run-Coverage {
    Print-Header "Running Tests with Coverage"
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    Write-Host ""
    Write-Host "Coverage report generated: coverage.out" -ForegroundColor Green
    Write-Host "Open HTML report with: go tool cover -html=coverage.out" -ForegroundColor Green
}

switch ($Target) {
    "all" { Run-AllTests }
    "service" { Run-ServiceTests }
    "server" { Run-ServerTests }
    "redis" { Run-RedisTests }
    "coverage" { Run-Coverage }
    default {
        Write-Host "Unknown option: $Target" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "✓ Tests completed successfully!" -ForegroundColor Green
