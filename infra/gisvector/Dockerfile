# Build argument for PostgreSQL version
ARG POSTGRES_VERSION=17-3.4

# Use the PostGIS image as the base
FROM postgis/postgis:${POSTGRES_VERSION}

# Re-declare PGVECTOR_VERSION after FROM to make it available in the build
ARG PGVECTOR_VERSION=v0.8.0

# Install necessary packages including clang-13
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       build-essential \
       libpq-dev \
       wget \
       git \
       postgresql-server-dev-17 \
       clang-13 \
       llvm-13-dev \
    # Clean up to reduce layer size
    && rm -rf /var/lib/apt/lists/* \
    && git clone --branch ${PGVECTOR_VERSION} https://github.com/pgvector/pgvector.git /tmp/pgvector \
    && cd /tmp/pgvector \
    && make \
    && make install \
    # Clean up unnecessary files
    && cd - \
    && apt-get purge -y --auto-remove build-essential postgresql-server-dev-17 libpq-dev wget git clang-13 llvm-13-dev \
    && rm -rf /tmp/pgvector

# Copy initialization scripts
COPY ./docker-entrypoint-initdb.d/ /docker-entrypoint-initdb.d/
