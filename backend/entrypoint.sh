#!/bin/sh

set -e

echo "Waiting for postgres..."
while ! pg_isready -h db -p 5432 -q -U myuser; do
  sleep 1
done

echo "PostgreSQL started"

echo "Running database migrations..."
migrate -database "$DATABASE_URL" -path db/migrations up

exec "$@"