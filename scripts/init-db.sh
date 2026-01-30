#!/bin/bash
# Database initialization script
# This script runs after the schema is applied

set -e

echo "🚀 Initializing Protobank database..."

# Wait for PostgreSQL to be ready
until pg_isready -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-protobank}"; do
  echo "⏳ Waiting for PostgreSQL to be ready..."
  sleep 2
done

echo "✅ PostgreSQL is ready!"

# Check if database is already initialized
if psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-protobank}" -tAc "SELECT 1 FROM categories LIMIT 1" >/dev/null 2>&1; then
  echo "✅ Database already initialized with seed data"
  exit 0
fi

echo "📊 Database initialized successfully!"
echo ""
echo "🔐 Database credentials:"
echo "  - Database: ${POSTGRES_DB:-protobank}"
echo "  - User: ${POSTGRES_USER:-postgres}"
echo "  - Password: ${POSTGRES_PASSWORD:-postgres}"
echo "  - Host: localhost"
echo "  - Port: ${POSTGRES_PORT:-5432}"
echo ""
echo "🎉 Protobank database is ready for development!"
