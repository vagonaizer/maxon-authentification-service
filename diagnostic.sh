#!/bin/bash

echo "=== Debugging Database Connection ==="

# Проверяем переменные окружения
echo "Environment variables:"
echo "DB_HOST: ${DB_HOST:-localhost}"
echo "DB_PORT: ${DB_PORT:-5433}"
echo "DB_USER: ${DB_USER:-postgres}"
echo "DB_NAME: ${DB_NAME:-auth_service}"

# Проверяем Docker контейнеры
echo -e "\n=== Docker containers ==="
docker ps | grep -E "(postgres|redis|kafka)"

# Проверяем подключение к PostgreSQL
echo -e "\n=== Testing PostgreSQL connection ==="
if docker exec auth-postgres pg_isready -U postgres -d auth_service; then
    echo "✅ PostgreSQL is ready"
else
    echo "❌ PostgreSQL is not ready"
    echo "PostgreSQL logs:"
    docker logs auth-postgres --tail 20
    exit 1
fi

# Проверяем существование базы данных
echo -e "\n=== Checking database ==="
docker exec -it auth-postgres psql -U postgres -c "\l" | grep auth_service

# Проверяем таблицы
echo -e "\n=== Checking tables ==="
docker exec -it auth-postgres psql -U postgres -d auth_service -c "\dt"

# Проверяем миграции
echo -e "\n=== Checking migrations ==="
docker exec -it auth-postgres psql -U postgres -d auth_service -c "SELECT version FROM schema_migrations ORDER BY version;"

echo -e "\n=== Debug completed ==="
