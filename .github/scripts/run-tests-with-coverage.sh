#!/bin/bash

# Script to run tests with coverage for Go workspace or single module
# Outputs coverage percentage and generates coverage.out file

set -euo pipefail

# Configuration
MIN_COVERAGE="${MIN_COVERAGE:-80}"
COVERAGE_FILES=()
COVERAGE_MODE="atomic"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
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
    local coverage_file=$2

    print_info "Testing module: $module_path"

    # Run tests and capture result
    if (cd "$module_path" && CGO_ENABLED=1 go test -race -coverprofile="../$coverage_file" -covermode=$COVERAGE_MODE ./... 2>&1); then
        echo "✅ Tests passed for $module_path"
    else
        print_warn "Some tests failed for $module_path (continuing)"
    fi

    # Check if coverage file was created and has content
    if [ -f "$coverage_file" ] && [ -s "$coverage_file" ]; then
        # Show module coverage
        MODULE_COV=$(go tool cover -func="$coverage_file" 2>/dev/null | grep -E '^total:' | awk '{print $3}' || echo "0%")
        print_info "Module coverage: $MODULE_COV"
        return 0
    else
        print_warn "No coverage data for $module_path"
        return 1
    fi
}

# Function to merge coverage files
merge_coverage_files() {
    local output_file=$1
    shift
    local files=("$@")

    print_info "Merging ${#files[@]} coverage files..."

    # Check if gocovmerge is available
    if command -v gocovmerge &> /dev/null; then
        gocovmerge "${files[@]}" > "$output_file"
    else
        # Manual merge
        echo "mode: $COVERAGE_MODE" > "$output_file"
        for file in "${files[@]}"; do
            if [ -f "$file" ] && [ -s "$file" ]; then
                grep -v '^mode:' "$file" >> "$output_file" || true
            fi
        done
    fi
}

# Main execution
main() {
    print_info "Starting test coverage calculation..."

    if [ -f go.work ]; then
        print_info "Running tests in workspace mode..."

        # Extract module paths from go.work
        while IFS= read -r line; do
            if [[ "$line" =~ ^[[:space:]]*\./(.+)$ ]]; then
                module_path="${BASH_REMATCH[1]}"

                if [ -d "$module_path" ] && [ -f "$module_path/go.mod" ]; then
                    coverage_file="coverage_$(echo "$module_path" | tr '/' '_').out"

                    if run_module_tests "$module_path" "$coverage_file"; then
                        COVERAGE_FILES+=("$coverage_file")
                    fi
                fi
            fi
        done < <(grep -E '^\s*\./.*' go.work)

        # Merge coverage files if any were generated
        if [ ${#COVERAGE_FILES[@]} -gt 0 ]; then
            merge_coverage_files "coverage.out" "${COVERAGE_FILES[@]}"

            # Clean up individual coverage files
            for file in "${COVERAGE_FILES[@]}"; do
                rm -f "$file"
            done
        else
            print_warn "No coverage files were generated"
            echo "mode: $COVERAGE_MODE" > coverage.out
        fi
    else
        # Single module mode
        print_info "Running tests in single module mode..."
        CGO_ENABLED=1 go test -v -race -coverprofile=coverage.out -covermode=$COVERAGE_MODE ./...
    fi

    # Calculate coverage percentage
    print_info "Calculating total coverage..."
    COVERAGE="0"

    if [ -s coverage.out ]; then
        # Check if coverage.out has actual coverage data (more than just the mode line)
        LINE_COUNT=$(wc -l < coverage.out)
        if [ $LINE_COUNT -gt 1 ]; then
            COVERAGE=$(go tool cover -func=coverage.out 2>/dev/null | grep -E '^total:' | awk '{print $3}' | sed 's/%//' || echo "0")
            if [ -z "$COVERAGE" ] || [ "$COVERAGE" = "" ]; then
                COVERAGE="0"
            fi
        else
            print_warn "Coverage file has no data"
        fi
    else
        print_warn "Coverage file is empty"
    fi

    # Validate coverage is a number
    if ! [[ "$COVERAGE" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
        print_error "Invalid coverage value: $COVERAGE, setting to 0"
        COVERAGE="0"
    fi

    print_info "Total Coverage: ${COVERAGE}%"

    # Output for GitHub Actions
    if [ -n "${GITHUB_OUTPUT:-}" ]; then
        echo "coverage=${COVERAGE}" >> "$GITHUB_OUTPUT"
    fi

    # Show coverage details
    if [ -s coverage.out ] && [ $(wc -l < coverage.out) -gt 1 ]; then
        echo ""
        print_info "Coverage by package (top 20):"
        go tool cover -func=coverage.out | head -20 || true
    fi

    # Check if coverage meets minimum requirement
    echo ""
    if (( $(echo "$COVERAGE < $MIN_COVERAGE" | bc -l) )); then
        print_error "Coverage ${COVERAGE}% is below minimum required ${MIN_COVERAGE}%"
        # Return non-zero to indicate failure, but don't exit
        # This allows the workflow to continue and report the coverage
        return 1
    else
        print_info "✅ Coverage ${COVERAGE}% meets minimum requirement"
        return 0
    fi
}

# Run main function
main "$@"
