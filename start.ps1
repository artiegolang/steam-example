# Запускаем Docker Compose
Write-Host "Starting Docker containers..."
docker-compose up -d

# Ждем 5 секунд, чтобы база данных успела запуститься
Write-Host "Waiting for database to start..."
Start-Sleep -Seconds 5

# Запускаем миграции
Write-Host "Running migrations..."
.\run-migrations.ps1

Write-Host "Setup complete! Your development environment is ready."
Write-Host "PostgreSQL is available at localhost:5432"
Write-Host "pgAdmin is available at http://localhost:5050"
Write-Host "pgAdmin credentials: admin@admin.com / admin"