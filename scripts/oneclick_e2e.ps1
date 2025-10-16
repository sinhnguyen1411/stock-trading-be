param(
  [string]$ComposeFile = "docker-compose.stack.yml",
  [string]$ConnectUrl = "http://localhost:8083",
  [string]$ConnectorConfig = "connector-mysql-user-outbox.json",
  [string]$DBHost = "localhost",
  [int]$DBPort = 3306,
  [string]$DBUser = "root",
  [SecureString]$DBPassword,
  [string]$DBName = "stock",
  [string]$ConfigPath = "./cmd/server/config/local.yaml"
)

# Resolve DB password from SecureString param or environment variable to avoid plaintext in script.
$pwdPlain = $null
if ($PSBoundParameters.ContainsKey('DBPassword') -and $DBPassword) {
  $ptr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($DBPassword)
  try { $pwdPlain = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($ptr) } finally { [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($ptr) }
} elseif ($env:MYSQL_PASSWORD) {
  $pwdPlain = $env:MYSQL_PASSWORD
} else {
  # Prompt securely if neither provided
  $sec = Read-Host -AsSecureString -Prompt "Enter MySQL password"
  $ptr2 = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($sec)
  try { $pwdPlain = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($ptr2) } finally { [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($ptr2) }
}

powershell -NoProfile -ExecutionPolicy Bypass -File "$PSScriptRoot/dev_up.ps1" -ComposeFile $ComposeFile -ConnectUrl $ConnectUrl -ConnectorConfig $ConnectorConfig -DBHost $DBHost -DBPort $DBPort -DBUser $DBUser -DBPassword $pwdPlain -DBName $DBName
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

powershell -NoProfile -ExecutionPolicy Bypass -File "$PSScriptRoot/e2e_full_mailpit.ps1" -ConfigPath $ConfigPath
exit $LASTEXITCODE
