#!/bin/bash

# Script to run database migrations for services
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
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-torra_dev_password}"
POSTGRES_DB="${POSTGRES_DB:-torra_dev}"
POSTGRES_SSLMODE="${POSTGRES_SSLMODE:-disable}"

# Parse command line arguments
COMMAND="${1:-up}"
SERVICE="${2:-all}"
VERBOSE="${3:-false}"

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if migrate tool is installed
check_migrate_tool() {
    if ! command -v migrate &> /dev/null; then
        print_color "$RED" "❌ migrate tool is not installed"
        print_color "$YELLOW" "Installing migrate tool..."

        # Detect OS and architecture
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)

        if [ "$ARCH" = "x86_64" ]; then
            ARCH="amd64"
        elif [ "$ARCH" = "aarch64" ]; then
            ARCH="arm64"
        fi

        # Download and install migrate
        MIGRATE_VERSION="v4.17.0"
        MIGRATE_URL="https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.${OS}-${ARCH}.tar.gz"

        curl -L "$MIGRATE_URL" | tar xvz
        sudo mv migrate /usr/local/bin/ 2>/dev/null || mv migrate ~/bin/

        if command -v migrate &> /dev/null; then
            print_color "$GREEN" "✅ migrate tool installed successfully"
        else
            print_color "$RED" "Failed to install migrate tool"
            exit 1
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

# Function to run migrations for a service
run_migrations() {
    local service=$1
    local command=$2
    local db_name="${POSTGRES_DB}_${service}"
    local migration_path="services/${service}/migrations"

    # Check if migration directory exists
    if [ ! -d "$migration_path" ]; then
        print_color "$YELLOW" "⚠️ No migrations found for service: $service"
        return 0
    fi

    # Check if there are migration files
    if [ -z "$(ls -A $migration_path/*.sql 2>/dev/null)" ]; then
        print_color "$YELLOW" "⚠️ No migration files found in $migration_path"
        return 0
    fi

    # Build database URL
    DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${db_name}?sslmode=${POSTGRES_SSLMODE}"

    print_color "$BLUE" "🚀 Running migrations for service: $service"
    print_color "$BLUE" "   Command: $command"

    if [ "$VERBOSE" = "true" ]; then
        print_color "$BLUE" "   Database: $db_name"
        print_color "$BLUE" "   Path: $migration_path"
    fi

    # Run migration command
    case $command in
        up)
            migrate -path "$migration_path" -database "$DATABASE_URL" up
            ;;
        down)
            migrate -path "$migration_path" -database "$DATABASE_URL" down 1
            ;;
        drop)
            print_color "$YELLOW" "⚠️ This will drop all tables. Are you sure? (y/N)"
            read -r confirmation
            if [ "$confirmation" = "y" ] || [ "$confirmation" = "Y" ]; then
                migrate -path "$migration_path" -database "$DATABASE_URL" drop -f
            else
                print_color "$YELLOW" "Cancelled"
                return 0
            fi
            ;;
        force)
            VERSION="${3:-}"
            if [ -z "$VERSION" ]; then
                print_color "$RED" "❌ Version number required for force command"
                return 1
            fi
            migrate -path "$migration_path" -database "$DATABASE_URL" force "$VERSION"
            ;;
        goto)
            VERSION="${3:-}"
            if [ -z "$VERSION" ]; then
                print_color "$RED" "❌ Version number required for goto command"
                return 1
            fi
            migrate -path "$migration_path" -database "$DATABASE_URL" goto "$VERSION"
            ;;
        version)
            migrate -path "$migration_path" -database "$DATABASE_URL" version
            ;;
        *)
            print_color "$RED" "❌ Unknown command: $command"
            return 1
            ;;
    esac

    print_color "$GREEN" "✅ Migration completed for $service"

    # Show current schema version
    if [ "$VERBOSE" = "true" ]; then
        print_color "$BLUE" "📊 Current schema version:"
        PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$db_name" -c "SELECT * FROM schema_migrations;"
    fi
}

# Function to get all services with migrations
get_services_with_migrations() {
    local services=()

    if [ -d "services" ]; then
        for service_dir in services/*; do
            if [ -d "$service_dir/migrations" ]; then
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
    print_color "$BLUE" "   Database Migration Runner"
    print_color "$BLUE" "==================================="
    echo

    # Check prerequisites
    check_migrate_tool
    check_postgres_connection || exit 1

    # Determine which services to run migrations for
    if [ "$SERVICE" = "all" ]; then
        SERVICES=($(get_services_with_migrations))
        if [ ${#SERVICES[@]} -eq 0 ]; then
            print_color "$YELLOW" "⚠️ No services with migrations found"
            exit 0
        fi
        print_color "$BLUE" "📋 Found ${#SERVICES[@]} service(s) with migrations: ${SERVICES[*]}"
    else
        SERVICES=("$SERVICE")
    fi

    echo

    # Run migrations for each service
    for service in "${SERVICES[@]}"; do
        # Create database if running 'up' command
        if [ "$COMMAND" = "up" ]; then
            create_service_database "$service"
        fi

        run_migrations "$service" "$COMMAND"
        echo
    done

    print_color "$GREEN" "==================================="
    print_color "$GREEN" "   Migration Runner Completed"
    print_color "$GREEN" "==================================="
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [COMMAND] [SERVICE] [VERBOSE]

Commands:
  up        - Apply all migrations
  down      - Rollback one migration
  drop      - Drop all tables (requires confirmation)
  force     - Force set migration version (requires VERSION)
  goto      - Go to specific migration version (requires VERSION)
  version   - Show current migration version

Arguments:
  SERVICE   - Service name or 'all' for all services (default: all)
  VERBOSE   - Show verbose output (true/false, default: false)

Environment variables:
  POSTGRES_HOST      - PostgreSQL host (default: localhost)
  POSTGRES_PORT      - PostgreSQL port (default: 5432)
  POSTGRES_USER      - PostgreSQL user (default: mytorra)
  POSTGRES_PASSWORD  - PostgreSQL password (default: mytorra_dev_password)
  POSTGRES_DB        - Base database name (default: mytorra_dev)
  POSTGRES_SSLMODE   - SSL mode (default: disable)

Examples:
  $0 up                     # Apply all migrations for all services
  $0 up core                # Apply migrations for core service only
  $0 down core true         # Rollback one migration for core service with verbose output
  $0 goto core 3 true       # Go to migration version 3 for core service
  $0 version core           # Show current version for core service

EOF
}

# Handle help flag
if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ] || [ "${1:-}" = "help" ]; then
    show_usage
    exit 0
fi

# Run main function
main
