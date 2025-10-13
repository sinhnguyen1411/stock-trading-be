param(
  [string]$ConnectUrl = "http://localhost:8083",
  [string]$ConfigPath = "connector-mysql-user-outbox.json"
)

if (!(Test-Path -Path $ConfigPath)) {
  Write-Error "Config file not found: ${ConfigPath}"
  exit 1
}

$jsonText = Get-Content -Raw -Path $ConfigPath
try {
  $json = $jsonText | ConvertFrom-Json -ErrorAction Stop
} catch {
  Write-Error "Invalid JSON in ${ConfigPath}: $_"
  exit 1
}

if (-not $json.name) {
  Write-Error "JSON must contain 'name' field."
  exit 1
}

$name = [string]$json.name
$cfg  = $json.config | ConvertTo-Json -Compress

try {
  $exists = Invoke-RestMethod -Method GET -Uri "$ConnectUrl/connectors/$name" -ErrorAction Stop
  Write-Host "Connector '$name' exists. Updating config..."
  $resp = Invoke-RestMethod -Method PUT -Uri "$ConnectUrl/connectors/$name/config" -ContentType 'application/json' -Body $cfg -ErrorAction Stop
  Write-Host "Updated connector '$name'"
} catch {
  Write-Host "Connector '$name' not found. Creating..."
  $body = @{ name = $name; config = $json.config } | ConvertTo-Json -Depth 8
  try {
    $resp = Invoke-RestMethod -Method POST -Uri "$ConnectUrl/connectors" -ContentType 'application/json' -Body $body -ErrorAction Stop
    Write-Host "Created connector '$name'"
  } catch {
    Write-Error "Failed to create connector: $_"
    exit 1
  }
}

Write-Host "Done."
