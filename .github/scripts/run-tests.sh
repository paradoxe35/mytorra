#!/bin/bash

# Script to run tests for Go workspace or single module
# Used by CI pipeline for test execution

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to run tests for a single module
run_module_tests() {
    local module_path=$1

    print_info "Testing module: $module_path"

    if (cd "$module_path" && CGO_ENABLED=1 go test -v -race ./... 2>&1); then
        echo "✅ All tests passed for $module_path"
        return 0
    else
        print_error "Tests failed for $module_path"
        return 1
    fi
}

# Main execution
main() {
    local failed_modules=()
    local test_count=0

    print_info "Starting test execution..."

    if [ -f go.work ]; then
        print_info "Running tests in workspace mode..."

        # Extract module paths from go.work
        while IFS= read -r line; do
            if [[ "$line" =~ ^[[:space:]]*\./(.+)$ ]]; then
                module_path="${BASH_REMATCH[1]}"

                if [ -d "$module_path" ] && [ -f "$module_path/go.mod" ]; then
                    ((test_count++))
                    if ! run_module_tests "$module_path"; then
                        failed_modules+=("$module_path")
                    fi
                fi
            fi
        done < <(grep -E '^\s*\./.*' go.work)
    else
        # Single module mode
        print_info "Running tests in single module mode..."
        ((test_count++))
        if ! CGO_ENABLED=1 go test -v -race ./...; then
            failed_modules+=(".")
        fi
    fi

    # Summary
    echo ""
    print_info "Test Summary:"
    echo "  Total modules tested: $test_count"
    echo "  Failed modules: ${#failed_modules[@]}"

    if [ ${#failed_modules[@]} -gt 0 ]; then
        print_error "The following modules have failing tests:"
        for module in "${failed_modules[@]}"; do
            echo "  - $module"
        done
        return 1
    else
        print_info "✅ All tests passed successfully!"
        return 0
    fi
}

# Run main function
main "$@"
