#!/bin/bash

# Script to run database migrations for services using Atlas
# This script can be used both locally and in CI/CD pipelines

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-mytorra}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-mytorra_dev_password}"
POSTGRES_DB="${POSTGRES_DB:-mytorra_dev}"
POSTGRES_SSLMODE="${POSTGRES_SSLMODE:-disable}"

# Atlas configuration
ATLAS_ENV="${ATLAS_ENV:-bun}"
ATLAS_DEV_URL="${ATLAS_DEV_URL:-docker://postgres/17?search_path=public}"

# Parse command line arguments
COMMAND="${1:-apply}"
SERVICE="${2:-all}"
VERBOSE="${3:-false}"

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if atlas tool is installed
check_atlas_tool() {
    if ! command -v atlas &> /dev/null; then
        print_color "$RED" "❌ Atlas tool is not installed"
        print_color "$YELLOW" "Installing Atlas tool..."

        # Download and install Atlas
        curl -sSf https://atlasgo.sh | sh

        if command -v atlas &> /dev/null; then
            print_color "$GREEN" "✅ Atlas tool installed successfully"
            atlas version
        else
            print_color "$RED" "Failed to install Atlas tool"
            print_color "$YELLOW" "Please install manually from: https://atlasgo.io/getting-started#installation"
            exit 1
        fi
    else
        if [ "$VERBOSE" = "true" ]; then
            print_color "$GREEN" "✅ Atlas version: $(atlas version)"
        fi
    fi
}

# Function to check if psql is installed
check_psql_tool() {
    if ! command -v psql &> /dev/null; then
        print_color "$YELLOW" "⚠️ psql client is not installed"
        print_color "$YELLOW" "Install it using:"
        print_color "$BLUE" "  Ubuntu/Debian: sudo apt-get install postgresql-client"
        print_color "$BLUE" "  MacOS: brew install postgresql"
        print_color "$BLUE" "  RHEL/CentOS: sudo yum install postgresql"
        return 1
    fi
    return 0
}

# Function to check PostgreSQL connection
check_postgres_connection() {
    print_color "$BLUE" "🔍 Checking PostgreSQL connection..."

    # Check if psql is installed first
    if ! check_psql_tool; then
        return 1
    fi

    if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT 1" &> /dev/null; then
        print_color "$GREEN" "✅ PostgreSQL connection successful"
        return 0
    else
        print_color "$RED" "❌ Failed to connect to PostgreSQL"
        print_color "$YELLOW" "Connection details:"
        echo "  Host: $POSTGRES_HOST:$POSTGRES_PORT"
        echo "  User: $POSTGRES_USER"
        echo "  Database: $POSTGRES_DB"
        return 1
    fi
}

# Function to create database for service
create_service_database() {
    local service=$1
    local db_name="${POSTGRES_DB}_${service}"

    print_color "$BLUE" "📦 Creating database for service: $service"

    # Check if database exists
    if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -lqt | cut -d \| -f 1 | grep -qw "$db_name"; then
        print_color "$YELLOW" "Database $db_name already exists"
    else
        PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<EOF
CREATE DATABASE ${db_name};
EOF
        print_color "$GREEN" "✅ Database $db_name created"
    fi

    # Enable required extensions
    print_color "$BLUE" "🔧 Enabling PostgreSQL extensions..."
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$db_name" <<'EOF'
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
EOF
    print_color "$GREEN" "✅ Extensions enabled"
}

# Function to run Atlas migrations for a service
run_atlas_migrations() {
    local service=$1
    local command=$2
    local db_name="${POSTGRES_DB}_${service}"
    local service_path="services/${service}"
    local migration_path="${service_path}/migrations"
    local atlas_config="${service_path}/atlas.hcl"

    # Check if Atlas config exists
    if [ ! -f "$atlas_config" ]; then
        print_color "$YELLOW" "⚠️ No atlas.hcl found for service: $service"
        return 0
    fi

    # Check if migration directory exists
    if [ ! -d "$migration_path" ]; then
        print_color "$YELLOW" "⚠️ No migrations directory found for service: $service"
        return 0
    fi

    # Build database URL
    DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${db_name}?sslmode=${POSTGRES_SSLMODE}"

    print_color "$BLUE" "🚀 Running Atlas migrations for service: $service"
    print_color "$BLUE" "   Command: $command"

    if [ "$VERBOSE" = "true" ]; then
        print_color "$BLUE" "   Database: $db_name"
        print_color "$BLUE" "   Config: $atlas_config"
        print_color "$BLUE" "   Migrations: $migration_path"
    fi

    # Change to service directory for Atlas commands
    cd "$service_path"

    # Run Atlas command
    case $command in
        apply)
            print_color "$BLUE" "📤 Applying migrations..."
            atlas migrate apply \
                --env "$ATLAS_ENV" \
                --url "$DATABASE_URL" \
                --dir "file://migrations"
            ;;

        status)
            print_color "$BLUE" "📊 Checking migration status..."
            atlas migrate status \
                --env "$ATLAS_ENV" \
                --url "$DATABASE_URL" \
                --dir "file://migrations"
            ;;

        validate)
            print_color "$BLUE" "✔️ Validating migrations..."
            atlas migrate validate \
                --env "$ATLAS_ENV" \
                --dev-url "$ATLAS_DEV_URL" \
                --dir "file://migrations"
            ;;

        diff)
            print_color "$BLUE" "🔍 Checking for schema differences..."

            # Create temporary env file for Atlas
            cat > .env.temp <<EOF
DB_HOST=$POSTGRES_HOST
DB_PORT=$POSTGRES_PORT
DB_USER=$POSTGRES_USER
DB_PASSWORD=$POSTGRES_PASSWORD
DB_NAME=$db_name
DB_SSL_MODE=$POSTGRES_SSLMODE
EOF

            # Check for pending migrations
            if atlas migrate diff \
                --env "$ATLAS_ENV" \
                --dev-url "$ATLAS_DEV_URL" \
                --to "file://migrations" \
                --format '{{ sql . "  " }}' 2>&1 | tee diff_output.txt; then

                # Check if the diff output contains any SQL statements
                if grep -q -E '(CREATE|ALTER|DROP|INSERT|UPDATE|DELETE)' diff_output.txt; then
                    print_color "$YELLOW" "⚠️ Schema differences detected! New migration may be needed."
                    cat diff_output.txt
                else
                    print_color "$GREEN" "✅ Schema is in sync with migrations"
                fi
            else
                print_color "$GREEN" "✅ No schema differences found"
            fi

            rm -f .env.temp diff_output.txt
            ;;

        new)
            # Create a new migration
            MIGRATION_NAME="${3:-change}"
            print_color "$BLUE" "📝 Creating new migration: $MIGRATION_NAME"

            atlas migrate diff "$MIGRATION_NAME" \
                --env "$ATLAS_ENV" \
                --dev-url "$ATLAS_DEV_URL" \
                --to "file://migrations" \
                --dir "file://migrations"
            ;;

        lint)
            print_color "$BLUE" "🔍 Linting migrations..."
            atlas migrate lint \
                --env "$ATLAS_ENV" \
                --dev-url "$ATLAS_DEV_URL" \
                --dir "file://migrations" \
                --latest 1
            ;;

        hash)
            print_color "$BLUE" "🔐 Updating migration hash sum..."
            atlas migrate hash \
                --env "$ATLAS_ENV" \
                --dir "file://migrations"
            ;;

        *)
            print_color "$RED" "❌ Unknown command: $command"
            cd - > /dev/null
            return 1
            ;;
    esac

    cd - > /dev/null
    print_color "$GREEN" "✅ Atlas migration command completed for $service"

    # Show current schema version
    if [ "$VERBOSE" = "true" ] && [ "$command" = "apply" ]; then
        print_color "$BLUE" "📊 Current schema version:"
        PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$db_name" -c "SELECT * FROM atlas_schema_revisions ORDER BY version DESC LIMIT 5;"
    fi
}

# Function to get all services with Atlas migrations
get_services_with_atlas() {
    local services=()

    if [ -d "services" ]; then
        for service_dir in services/*; do
            if [ -f "$service_dir/atlas.hcl" ] && [ -d "$service_dir/migrations" ]; then
                service_name=$(basename "$service_dir")
                # Check if there are actual migration files
                if ls "$service_dir/migrations"/*.sql &> /dev/null; then
                    services+=("$service_name")
                fi
            fi
        done
    fi

    echo "${services[@]}"
}

# Main execution
main() {
    print_color "$BLUE" "==================================="
    print_color "$BLUE" "   Atlas Migration Runner"
    print_color "$BLUE" "==================================="
    echo

    # Check prerequisites
    check_atlas_tool
    check_postgres_connection || exit 1

    # Determine which services to run migrations for
    if [ "$SERVICE" = "all" ]; then
        SERVICES=($(get_services_with_atlas))
        if [ ${#SERVICES[@]} -eq 0 ]; then
            print_color "$YELLOW" "⚠️ No services with Atlas configuration found"
            exit 0
        fi
        print_color "$BLUE" "📋 Found ${#SERVICES[@]} service(s) with Atlas migrations: ${SERVICES[*]}"
    else
        SERVICES=("$SERVICE")
    fi

    echo

    # Run migrations for each service
    for service in "${SERVICES[@]}"; do
        # Create database if running 'apply' command
        if [ "$COMMAND" = "apply" ] || [ "$COMMAND" = "diff" ] || [ "$COMMAND" = "status" ]; then
            create_service_database "$service"
        fi

        run_atlas_migrations "$service" "$COMMAND"
        echo
    done

    print_color "$GREEN" "==================================="
    print_color "$GREEN" "   Atlas Migration Runner Completed"
    print_color "$GREEN" "==================================="
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [COMMAND] [SERVICE] [VERBOSE/MIGRATION_NAME]

Commands:
  apply     - Apply all pending migrations (default)
  status    - Show current migration status
  validate  - Validate migration files
  diff      - Check for schema differences between code and database
  new       - Create a new migration (requires MIGRATION_NAME as 3rd argument)
  lint      - Lint the latest migration
  hash      - Update migration hash sum

Arguments:
  SERVICE        - Service name or 'all' for all services (default: all)
  VERBOSE        - Show verbose output (true/false, default: false)
  MIGRATION_NAME - Name for new migration (only for 'new' command)

Environment variables:
  POSTGRES_HOST      - PostgreSQL host (default: localhost)
  POSTGRES_PORT      - PostgreSQL port (default: 5432)
  POSTGRES_USER      - PostgreSQL user (default: mytorra)
  POSTGRES_PASSWORD  - PostgreSQL password (default: mytorra_dev_password)
  POSTGRES_DB        - Base database name (default: mytorra_dev)
  POSTGRES_SSLMODE   - SSL mode (default: disable)
  ATLAS_ENV          - Atlas environment to use (default: bun)
  ATLAS_DEV_URL      - Atlas dev database URL (default: docker://postgres/17?search_path=public)

Examples:
  $0 apply                  # Apply all migrations for all services
  $0 apply core             # Apply migrations for core service only
  $0 status core            # Show migration status for core service
  $0 diff core              # Check for schema differences in core service
  $0 new core add_users     # Create new migration 'add_users' for core service
  $0 validate core true     # Validate migrations for core with verbose output
  $0 lint core              # Lint the latest migration for core service

Atlas Configuration:
  Each service must have an atlas.hcl file in its directory
  Migrations should be in the 'migrations' subdirectory

EOF
}

# Handle help flag
if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ] || [ "${1:-}" = "help" ]; then
    show_usage
    exit 0
fi

# Run main function
main
