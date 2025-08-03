#!/bin/bash

echo "=== Testing Session Creation ==="

# Проверим структуру таблицы sessions
echo "Checking sessions table structure:"
docker exec -it auth-postgres psql -U postgres -d auth_service -c "\d sessions"

# Попробуем создать сессию вручную
echo -e "\nTesting manual session creation:"
docker exec -it auth-postgres psql -U postgres -d auth_service -c "
INSERT INTO sessions (id, user_id, refresh_token, user_agent, ip_address, is_active, expires_at)
VALUES (
    'test-session-id'::uuid,
    (SELECT id FROM users WHERE email = 'test@example.com'),
    'test-refresh-token',
    'TestAgent',
    '127.0.0.1'::inet,
    true,
    NOW() + INTERVAL '7 days'
)
RETURNING id, created_at;"

# Проверим созданную сессию
echo -e "\nChecking created session:"
docker exec -it auth-postgres psql -U postgres -d auth_service -c "
SELECT id, user_id, refresh_token, user_agent, ip_address, is_active, expires_at, created_at 
FROM sessions 
WHERE refresh_token = 'test-refresh-token';"

# Удалим тестовую сессию
echo -e "\nCleaning up test session:"
docker exec -it auth-postgres psql -U postgres -d auth_service -c "
DELETE FROM sessions WHERE refresh_token = 'test-refresh-token';"
