#!/bin/bash
set -e

# Script to create multiple PostgreSQL databases
# Usage: Set POSTGRES_MULTIPLE_DATABASES environment variable with comma-separated database names

function create_user_and_database() {
    local database=$1
    echo "Creating user and database '$database'"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        CREATE USER ${database}_user WITH PASSWORD '${database}_password';
        CREATE DATABASE $database;
        GRANT ALL PRIVILEGES ON DATABASE $database TO ${database}_user;
        \c $database;
        GRANT ALL ON SCHEMA public TO ${database}_user;
        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ${database}_user;
        GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${database}_user;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ${database}_user;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ${database}_user;
EOSQL
}

if [ -n "$POSTGRES_MULTIPLE_DATABASES" ]; then
    echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"
    for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
        create_user_and_database $db
    done
    echo "Multiple databases created"
else
    echo "No multiple databases specified"
fi
