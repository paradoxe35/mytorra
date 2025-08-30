#!/bin/bash

# Script to dynamically detect existing services
# This prevents building non-existent services and makes the CI more efficient

set -euo pipefail

SERVICES_DIR="services"
OUTPUT_FORMAT="${1:-json}"  # json, array, or list

if [ ! -d "$SERVICES_DIR" ]; then
    echo "❌ Services directory not found: $SERVICES_DIR"
    exit 1
fi

# Find all service directories that contain Go modules
EXISTING_SERVICES=()

echo "🔍 Scanning for existing services in $SERVICES_DIR..."

for service_dir in "$SERVICES_DIR"/*; do
    if [ -d "$service_dir" ]; then
        service_name=$(basename "$service_dir")

        # Check if it's a valid Go service (has go.mod and at least one cmd directory)
        if [ -f "$service_dir/go.mod" ] && [ -d "$service_dir/cmd" ]; then
            # Check what components exist
            components=()

            if [ -d "$service_dir/cmd/api" ]; then
                components+=("api")
            fi

            if [ -d "$service_dir/cmd/worker" ]; then
                components+=("worker")
            fi

            if [ ${#components[@]} -gt 0 ]; then
                EXISTING_SERVICES+=("$service_name")
                echo "  ✅ $service_name (${components[*]})"
            else
                echo "  ⚠️ $service_name (no cmd/api or cmd/worker found)"
            fi
        else
            echo "  ❌ $service_name (missing go.mod or cmd directory)"
        fi
    fi
done

if [ ${#EXISTING_SERVICES[@]} -eq 0 ]; then
    echo "⚠️ No valid services found!"
    case $OUTPUT_FORMAT in
        json)
            echo "[]"
            ;;
        array)
            echo ""
            ;;
        list)
            echo ""
            ;;
    esac
    exit 0
fi

echo ""
echo "📋 Found ${#EXISTING_SERVICES[@]} valid service(s): ${EXISTING_SERVICES[*]}"

# Output in requested format
case $OUTPUT_FORMAT in
    json)
        # Output as JSON array for GitHub Actions matrix
        printf '['
        for i in "${!EXISTING_SERVICES[@]}"; do
            printf '"%s"' "${EXISTING_SERVICES[$i]}"
            if [ $i -lt $((${#EXISTING_SERVICES[@]} - 1)) ]; then
                printf ','
            fi
        done
        printf ']'
        ;;
    array)
        # Output as space-separated array
        echo "${EXISTING_SERVICES[*]}"
        ;;
    list)
        # Output as newline-separated list
        printf '%s\n' "${EXISTING_SERVICES[@]}"
        ;;
    *)
        echo "❌ Unknown output format: $OUTPUT_FORMAT"
        echo "Available formats: json, array, list"
        exit 1
        ;;
esac
