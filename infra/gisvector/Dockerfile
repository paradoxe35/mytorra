# Use the PostGIS image as the base
FROM postgis/postgis:17-3.4

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
    && git clone --branch v0.8.0 https://github.com/pgvector/pgvector.git /tmp/pgvector \
    && cd /tmp/pgvector \
    && make \
    && make install \
    # Clean up unnecessary files
    && cd - \
    && apt-get purge -y --auto-remove build-essential postgresql-server-dev-17 libpq-dev wget git clang-13 llvm-13-dev \
    && rm -rf /tmp/pgvector

# Copy initialization scripts
COPY ./docker-entrypoint-initdb.d/ /docker-entrypoint-initdb.d/
