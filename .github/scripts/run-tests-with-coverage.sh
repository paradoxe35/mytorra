#!/bin/bash

# Script to run tests with coverage for Go workspace or single module
# Outputs coverage percentage and generates coverage.out file

set -uo pipefail

# Configuration
MIN_COVERAGE="${MIN_COVERAGE:-80}"
COVERAGE_FILES=()
COVERAGE_MODE="atomic"

# Check if CGO should be enabled (based on gcc availability)
CGO_ENABLED=1
if ! command -v gcc &> /dev/null; then
    print_warn "gcc not found, disabling CGO for tests"
    CGO_ENABLED=0
fi

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

# Function to check if module has tests
has_test_files() {
    local module_path=$1
    find "$module_path" -name "*_test.go" -type f 2>/dev/null | grep -q .
}

# Function to run tests for a single module
run_module_tests() {
    local module_path=$1
    local coverage_file=$2

    print_info "Testing module: $module_path"

    # Check if module has any test files
    if ! has_test_files "$module_path"; then
        print_info "No test files in $module_path - skipping"
        return 1
    fi

    # Run tests and capture result
    local test_output
    local test_exit_code=0

    # Get absolute path for coverage file
    local abs_coverage_file="$(pwd)/$coverage_file"

    pushd "$module_path" > /dev/null 2>&1
    # Use -count=1 to avoid cached results that don't generate coverage files
    # Note: -race flag requires CGO, so only use it when CGO is enabled
    if [ $CGO_ENABLED -eq 1 ]; then
        test_output=$(CGO_ENABLED=1 go test -count=1 -race -coverprofile="$abs_coverage_file" -covermode=$COVERAGE_MODE ./... 2>&1) || test_exit_code=$?
    else
        test_output=$(CGO_ENABLED=0 go test -count=1 -coverprofile="$abs_coverage_file" -covermode=$COVERAGE_MODE ./... 2>&1) || test_exit_code=$?
    fi
    popd > /dev/null 2>&1

    # Show test output
    echo "$test_output"

    if [ $test_exit_code -eq 0 ]; then
        echo "✅ Tests passed for $module_path"
    else
        print_warn "Some tests failed for $module_path (continuing)"
    fi

    # Check if coverage file was created and has actual coverage data
    # Coverage file is in the parent directory (current directory after popd)
    if [ -f "$coverage_file" ] && [ -s "$coverage_file" ]; then
        # Check if the file has actual coverage data
        local line_count=$(wc -l < "$coverage_file")
        if [ $line_count -gt 1 ]; then
            # Calculate module coverage
            local module_cov=$(go tool cover -func="$coverage_file" 2>/dev/null | grep -E '^total:' | awk '{print $3}' | sed 's/%//' || echo "0")

            if [ -n "$module_cov" ] && [ "$module_cov" != "0" ] && [ "$module_cov" != "0.0" ]; then
                print_info "Module coverage: ${module_cov}%"
                return 0
            else
                print_warn "Module has tests but 0% coverage: $module_path"
                rm -f "$coverage_file"
                return 1
            fi
        else
            print_info "Coverage file has no data for $module_path"
            rm -f "$coverage_file"
            return 1
        fi
    else
        print_warn "No coverage file generated for $module_path (looked for $coverage_file in $(pwd))"
        # Debug: list files to see what's there
        ls -la coverage_*.out 2>/dev/null || true
        return 1
    fi
}

# Function to merge coverage files
merge_coverage_files() {
    local output_file=$1
    shift
    local files=("$@")

    if [ ${#files[@]} -eq 0 ]; then
        print_warn "No coverage files to merge"
        return 1
    fi

    print_info "Merging ${#files[@]} coverage files..."

    # Start with mode line
    echo "mode: $COVERAGE_MODE" > "$output_file"

    # Append all coverage data
    for file in "${files[@]}"; do
        if [ -f "$file" ] && [ -s "$file" ]; then
            # Skip the mode line and append the rest
            tail -n +2 "$file" >> "$output_file"
        fi
    done

    return 0
}

# Main execution
main() {
    print_info "Starting test coverage calculation..."

    local modules_with_tests=0
    local modules_with_coverage=0

    if [ -f go.work ]; then
        print_info "Running tests in workspace mode..."

        # Extract module paths from go.work
        while IFS= read -r line; do
            if [[ "$line" =~ ^[[:space:]]*\./(.+)$ ]]; then
                module_path="${BASH_REMATCH[1]}"

                if [ -d "$module_path" ] && [ -f "$module_path/go.mod" ]; then
                    if has_test_files "$module_path"; then
                        ((modules_with_tests++)) || true
                        coverage_file="coverage_$(echo "$module_path" | tr '/' '_').out"

                        if run_module_tests "$module_path" "$coverage_file"; then
                            COVERAGE_FILES+=("$coverage_file")
                            ((modules_with_coverage++)) || true
                        fi
                    else
                        print_info "Skipping $module_path (no test files)"
                    fi
                fi
            fi
        done < <(grep -E '^\s*\./.*' go.work)

        print_info "Found $modules_with_tests modules with tests, $modules_with_coverage with coverage"

        # Merge coverage files if any were generated
        if [ ${#COVERAGE_FILES[@]} -gt 0 ]; then
            if merge_coverage_files "coverage.out" "${COVERAGE_FILES[@]}"; then
                # Clean up individual coverage files
                for file in "${COVERAGE_FILES[@]}"; do
                    rm -f "$file"
                done
            else
                print_error "Failed to merge coverage files"
                echo "mode: $COVERAGE_MODE" > coverage.out
            fi
        else
            print_warn "No modules generated coverage data"
            echo "mode: $COVERAGE_MODE" > coverage.out
        fi
    else
        # Single module mode
        print_info "Running tests in single module mode..."
        if has_test_files "."; then
            if [ $CGO_ENABLED -eq 1 ]; then
                CGO_ENABLED=1 go test -v -race -coverprofile=coverage.out -covermode=$COVERAGE_MODE ./...
            else
                CGO_ENABLED=0 go test -v -coverprofile=coverage.out -covermode=$COVERAGE_MODE ./...
            fi
        else
            print_warn "No test files found"
            echo "mode: $COVERAGE_MODE" > coverage.out
        fi
    fi

    # Calculate coverage percentage
    print_info "Calculating total coverage..."
    COVERAGE="0"

    if [ -s coverage.out ]; then
        # Check if coverage.out has actual coverage data
        LINE_COUNT=$(wc -l < coverage.out)
        if [ $LINE_COUNT -gt 1 ]; then
            # Calculate total coverage
            TOTAL_COV=$(go tool cover -func=coverage.out 2>/dev/null | grep -E '^total:' | awk '{print $3}' | sed 's/%//' || echo "")

            if [ -n "$TOTAL_COV" ] && [ "$TOTAL_COV" != "" ]; then
                COVERAGE="$TOTAL_COV"
            else
                # Try alternative: sum up coverage from individual packages
                local sum=0
                local count=0
                while read -r line; do
                    if [[ "$line" =~ ([0-9]+\.[0-9]+)% ]]; then
                        local pkg_cov="${BASH_REMATCH[1]}"
                        sum=$(echo "$sum + $pkg_cov" | bc)
                        ((count++)) || true
                    fi
                done < <(go tool cover -func=coverage.out 2>/dev/null | grep -v '^total:' || echo "")

                if [ $count -gt 0 ]; then
                    COVERAGE=$(echo "scale=1; $sum / $count" | bc)
                fi
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
        print_info "Coverage by package:"
        go tool cover -func=coverage.out 2>/dev/null | head -30 || true
    fi

    # Check if coverage meets minimum requirement
    echo ""
    if [ "$modules_with_coverage" -eq 0 ]; then
        print_warn "No modules have test coverage"
        # Don't fail if no modules have tests
        return 0
    fi

    if (( $(echo "$COVERAGE < $MIN_COVERAGE" | bc -l) )); then
        print_error "Coverage ${COVERAGE}% is below minimum required ${MIN_COVERAGE}%"
        # Return non-zero to indicate failure, but don't exit
        return 1
    else
        print_info "✅ Coverage ${COVERAGE}% meets minimum requirement"
        return 0
    fi
}

# Run main function
main "$@"
