#!/usr/bin/env bash

# Bash strict mode - exit on error, undefined vars, and pipe failures
set -euo pipefail

# Script directory for relative path resolution
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to project root directory
cd "${PROJECT_ROOT}"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_and_run() {
    log_info "Running: $*"
    "$@"
}

# Error handler
error_exit() {
    log_error "Script failed at line $1"
    exit 1
}

# Trap errors
trap 'error_exit $LINENO' ERR

# Main function
main() {
    log_info "Starting bact workflow development script..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check if the cmd/bact directory exists
    if [[ ! -d "cmd/bact" ]]; then
        log_error "cmd/bact directory not found. Are you in the correct project root?"
        exit 1
    fi

    # Check if the example workflow file exists
    local workflow_file="./examples/workflows/hello.yaml"
    if [[ ! -f "${workflow_file}" ]]; then
        log_warn "Example workflow file '${workflow_file}' not found"
        log_info "Proceeding anyway - bact will handle the missing file"
    fi

    # Run the bact workflow command
    log_and_run go run ./cmd/bact workflow run -f "${workflow_file}"

    log_info "bact workflow execution completed successfully"
}

# Run main function
main "$@"
