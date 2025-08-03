#!/bin/bash

echo "=== Debugging Login Issue ==="

# Проверяем пользователя в базе данных
echo "Checking user in database:"
docker exec -it auth-postgres psql -U postgres -d auth_service -c "
SELECT id, email, password_hash, username, is_active, is_verified, created_at 
FROM users 
WHERE email = 'test@example.com';"

# Проверяем сессии
echo -e "\nChecking sessions:"
docker exec -it auth-postgres psql -U postgres -d auth_service -c "
SELECT id, user_id, ip_address, user_agent, is_active, expires_at, created_at 
FROM sessions 
ORDER BY created_at DESC 
LIMIT 5;"

# Проверяем роли пользователя
echo -e "\nChecking user roles:"
docker exec -it auth-postgres psql -U postgres -d auth_service -c "
SELECT u.email, r.name as role_name 
FROM users u 
LEFT JOIN user_roles ur ON u.id = ur.user_id 
LEFT JOIN roles r ON ur.role_id = r.id 
WHERE u.email = 'test@example.com';"

# Проверяем доступные роли
echo -e "\nChecking available roles:"
docker exec -it auth-postgres psql -U postgres -d auth_service -c "
SELECT id, name, description FROM roles;"
