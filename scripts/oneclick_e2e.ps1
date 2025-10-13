param(
  [string]$ComposeFile = "docker-compose.stack.yml",
  [string]$ConnectUrl = "http://localhost:8083",
  [string]$ConnectorConfig = "connector-mysql-user-outbox.json",
  [string]$DBHost = "localhost",
  [int]$DBPort = 3306,
  [string]$DBUser = "root",
  [string]$DBPassword = "Ngdms1107#",
  [string]$DBName = "stock",
  [string]$ConfigPath = "./cmd/server/config/local.yaml"
)

powershell -NoProfile -ExecutionPolicy Bypass -File "$PSScriptRoot/dev_up.ps1" -ComposeFile $ComposeFile -ConnectUrl $ConnectUrl -ConnectorConfig $ConnectorConfig -DBHost $DBHost -DBPort $DBPort -DBUser $DBUser -DBPassword $DBPassword -DBName $DBName
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

powershell -NoProfile -ExecutionPolicy Bypass -File "$PSScriptRoot/e2e_full_mailpit.ps1" -ConfigPath $ConfigPath
exit $LASTEXITCODE

