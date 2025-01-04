# Загружаем переменные из .env файла
Get-Content .env | ForEach-Object {
    $name, $value = $_.split('=')
    if ($name -and $value) {
        Set-Item -Path "env:$name" -Value $value
    }
}

Write-Host "Current environment variables:"
Write-Host "DB_HOST: $env:DB_HOST"
Write-Host "DB_PORT: $env:DB_PORT"
Write-Host "DB_USER: $env:DB_USER"
Write-Host "DB_NAME: $env:DB_NAME"

# Используем стандартный формат строки подключения PostgreSQL
$env:GOOSE_DRIVER = "postgres"
$env:GOOSE_DBSTRING = "host=127.0.0.1 user=postgres password=postgres dbname=trading_platform sslmode=disable"

Write-Host "`nUsing connection string: $env:GOOSE_DBSTRING"
Write-Host "Running migrations..."
goose -dir db-service/migrations up