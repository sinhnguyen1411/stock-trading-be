param(
  [string]$ComposeFile = "docker-compose.stack.yml",
  [string]$ConnectUrl = "http://localhost:8083",
  [string]$ConnectorConfig = "connector-mysql-user-outbox.json",
  [string]$DBHost = "localhost",
  [int]$DBPort = 3306,
  [string]$DBUser = "root",
  [string]$DBPassword = "Ngdms1107#",
  [string]$DBName = "stock"
)

function Require-Cli($name) {
  if (-not (Get-Command $name -ErrorAction SilentlyContinue)) {
    Write-Error "Required CLI not found: $name"
    exit 1
  }
}

Require-Cli docker
Require-Cli curl
Require-Cli mysql

Write-Host "Starting infrastructure via docker compose..."
& docker compose -f $ComposeFile up -d
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Waiting for Kafka Connect at $ConnectUrl ..."
$deadline = (Get-Date).AddSeconds(60)
while ((Get-Date) -lt $deadline) {
  try {
    $resp = Invoke-RestMethod -Method GET -Uri "$ConnectUrl/connectors" -TimeoutSec 2
    Write-Host "Kafka Connect is up."
    break
  } catch {
    Start-Sleep -Seconds 2
  }
}

if ((Get-Date) -ge $deadline) {
  Write-Warning "Kafka Connect did not become ready within timeout. Continuing..."
}

Write-Host "Initializing MySQL schema..."
& $PSScriptRoot\init_db.ps1 -DBHost $DBHost -Port $DBPort -User $DBUser -Password $DBPassword -Database $DBName
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Registering Debezium connector..."
& $PSScriptRoot\register_debezium_connector.ps1 -ConnectUrl $ConnectUrl -ConfigPath $ConnectorConfig
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "All set!"
Write-Host "- Mailpit SMTP: 127.0.0.1:1025, UI: http://localhost:8025"
Write-Host "- HTTP Gateway: http://127.0.0.1:18080"
Write-Host "- Verify URL base: http://127.0.0.1:18080/users/verify?token="
