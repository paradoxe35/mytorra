# MyTorra Platform - Tilt Development Configuration

load('ext://restart_process', 'docker_build_with_restart')
load('ext://restart_process', 'custom_build_with_restart')

# Create namespace for the application
k8s_yaml('infra/kubernetes/dev/namespace.yaml')

# Docker prune on startup
docker_prune_settings(disable=False, keep_recent=1, interval_hrs=1)

# K8S
allow_k8s_contexts('my-k3s-etiolles')

# =============================================================================
# Infrastructure Services
# =============================================================================

# =============================================================================
# Application Services
# =============================================================================

# Core API Service
custom_build_with_restart(
    'ghcr.io/paradoxe35/mytorra-core-api:dev',
    'docker build -t $EXPECTED_REF -f services/core/Dockerfile.api .',
    entrypoint='/app/core-api',
    deps=['services/core/', 'packages/'],
    live_update=[
        sync('services/core/', '/app/services/core/'),
        sync('packages/', '/app/packages/'),
    ]
)

# Core API Kubernetes manifests
k8s_yaml(['infra/kubernetes/dev/api-deployment.yaml'])

# Configure core-api resource
# k8s_resource(
#     'core-api',
#     port_forwards=['8080:8080'],
#     resource_deps=['postgresql', 'redis', 'rabbitmq'],
#     labels=['core-services']
# )

# =============================================================================
# Development Tools and Utilities
# =============================================================================

# Database migration runner
local_resource(
    'db-migrate',
    cmd='echo "Database migration placeholder - implement when migration tool is ready"',
    deps=['services/core/migrations/'],
    labels=['utilities'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL
)

# API documentation generator
local_resource(
    'generate-docs',
    cmd='make docs',
    deps=['services/core/**/*.go'],
    labels=['utilities'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL
)

# i18n validation
local_resource(
    'validate-i18n',
    cmd='make i18n-validate',
    deps=['services/core/**/*.go', 'packages/i18n/locales/*.json'],
    labels=['utilities'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL
)

# =============================================================================
# Configuration and Help
# =============================================================================

print("""
🚀 MyTorra Development Environment

Manual triggers available:
  - db-migrate: Run database migrations
  - generate-docs: Generate API documentation
  - validate-i18n: Validate i18n keys
""")
