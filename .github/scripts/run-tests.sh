#!/bin/bash

# Script to run tests for Go workspace or single module
# Used by CI pipeline for test execution

# Don't exit on error immediately - we want to handle errors ourselves
set -uo pipefail

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

    # Run tests and capture output and exit code
    local test_output
    local test_exit_code=0

    # Change to module directory and run tests
    pushd "$module_path" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        print_error "Failed to enter directory: $module_path"
        popd > /dev/null 2>&1 || true
        return 1
    fi

    # Run the tests
    echo "Running: CGO_ENABLED=1 go test -v -race ./..."
    test_output=$(CGO_ENABLED=1 go test -v -race ./... 2>&1) || test_exit_code=$?

    # Return to original directory
    popd > /dev/null 2>&1

    # Show the output
    echo "$test_output"

    if [ $test_exit_code -eq 0 ]; then
        echo "✅ All tests passed for $module_path"
        return 0
    else
        print_error "Tests failed for $module_path (exit code: $test_exit_code)"
        return 1
    fi
}

# Main execution
main() {
    local failed_modules=()
    local test_count=0

    print_info "Starting test execution..."

    # Check current directory
    echo "Current directory: $(pwd)"

    if [ -f go.work ]; then
        print_info "Running tests in workspace mode..."
        echo "go.work content:"
        cat go.work
        echo ""

        # Extract module paths from go.work
        local modules_found=0
        while IFS= read -r line; do
            if [[ "$line" =~ ^[[:space:]]*\./(.+)$ ]]; then
                module_path="${BASH_REMATCH[1]}"
                ((modules_found++)) || true

                echo ""
                echo "=== Module $modules_found: $module_path ==="

                if [ ! -d "$module_path" ]; then
                    print_error "Directory not found: $module_path"
                    continue
                fi

                if [ ! -f "$module_path/go.mod" ]; then
                    print_warn "No go.mod in $module_path, skipping"
                    continue
                fi

                ((test_count++)) || true
                if ! run_module_tests "$module_path"; then
                    failed_modules+=("$module_path")
                fi
            fi
        done < <(grep -E '^\s*\./.*' go.work || echo "")

        if [ $modules_found -eq 0 ]; then
            print_error "No modules found in go.work file"
            return 1
        fi
    else
        # Single module mode
        print_info "Running tests in single module mode..."
        ((test_count++)) || true

        echo "Running: CGO_ENABLED=1 go test -v -race ./..."
        if ! CGO_ENABLED=1 go test -v -race ./...; then
            failed_modules+=(".")
        fi
    fi

    # Summary
    echo ""
    echo "================================"
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
        if [ $test_count -eq 0 ]; then
            print_error "No tests were run!"
            return 1
        fi
        print_info "✅ All tests passed successfully!"
        return 0
    fi
}

# Trap errors and show where they occurred
trap 'print_error "Error occurred at line $LINENO"' ERR

# Run main function and capture exit code
main "$@"
exit_code=$?

# Ensure we return the correct exit code
exit $exit_code
