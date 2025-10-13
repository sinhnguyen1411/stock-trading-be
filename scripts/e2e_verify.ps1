param(
  [string]$ConfigPath = "./cmd/server/config/local.yaml",
  [string]$DBHost = "localhost",
  [int]$DBPort = 3306,
  [string]$DBUser = "root",
  [string]$DBPassword = "Ngdms1107#",
  [string]$DBName = "stock",
  [string]$BaseUrl = "http://127.0.0.1:18080",
  [int]$TimeoutSeconds = 60
)

$ErrorActionPreference = 'Stop'

function Wait-Port {
  param([string]$HostName,[int]$Port,[int]$Seconds)
  $deadline = (Get-Date).AddSeconds($Seconds)
  while ((Get-Date) -lt $deadline) {
    try {
      $tcp = New-Object System.Net.Sockets.TcpClient
      $iar = $tcp.BeginConnect($HostName,$Port,$null,$null)
      if ($iar.AsyncWaitHandle.WaitOne(500)) {
        $tcp.EndConnect($iar); $tcp.Close(); return $true
      }
    } catch { }
    Start-Sleep -Milliseconds 500
  }
  return $false
}

function Docker-MySQL([string]$sql) {
  $args = @('run','--rm','-e',"MYSQL_PWD=$DBPassword",'mysql:8','mysql',
    "--host=host.docker.internal","--port=$DBPort","--user=$DBUser","-N","-e", $sql)
  $output = & docker @args
  if ($LASTEXITCODE -ne 0) { throw "docker mysql failed: $output" }
  return $output
}

Write-Host "Starting server in background..."
$job = Start-Job -ScriptBlock { param($cfg) go run main.go server --config $cfg } -ArgumentList $ConfigPath
Start-Sleep -Seconds 2

if (-not (Wait-Port -HostName '127.0.0.1' -Port 18080 -Seconds $TimeoutSeconds)) {
  try { Receive-Job -Id $job.Id -Keep | Select-Object -Last 50 } catch {}
  Stop-Job -Id $job.Id -Force | Out-Null
  throw "HTTP gateway not listening on :18080 within timeout"
}
Write-Host "HTTP gateway is up."

$suffix = (Get-Random -Minimum 100000 -Maximum 999999)
$username = "user$suffix"
$password = "secret12"
$email = "${username}@example.com"

Write-Host "Registering user $username ..."
$payload = @{ username=$username; password=$password; email=$email; name='E2E Tester'; cmnd='123456789'; birthday=1714608000; gender=$true; permanent_address='HN'; phone_number='0900000009' } | ConvertTo-Json
$reg = Invoke-RestMethod -Method Post -Uri "$BaseUrl/users" -ContentType 'application/json' -Body $payload -TimeoutSec 20
Write-Host "Register response: $($reg | ConvertTo-Json -Depth 4)"

Start-Sleep -Seconds 2
Write-Host "Fetching latest verification token from DB..."
$token = (Docker-MySQL "SELECT t.token FROM ${DBName}.user_verification_tokens t JOIN ${DBName}.users u ON u.id=t.user_id WHERE u.username='${username}' ORDER BY t.id DESC LIMIT 1;").Trim()
if (-not $token) { throw "Token not found for user $username" }
Write-Host "Token: $token"

Write-Host "Verifying user via API..."
$verify = Invoke-RestMethod -Method Get -Uri "$BaseUrl/users/verify?token=$token" -TimeoutSec 20
Write-Host "Verify response: $($verify | ConvertTo-Json -Depth 6)"

Write-Host "Login after verifying..."
$loginBody = @{ username=$username; password=$password } | ConvertTo-Json
$login = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/user/login" -ContentType 'application/json' -Body $loginBody -TimeoutSec 20
Write-Host "Login response: $($login | ConvertTo-Json -Depth 6)"

Write-Host "E2E success for $username"

try { Stop-Job -Id $job.Id -Force | Out-Null } catch {}
Remove-Job -Id $job.Id -Force -ErrorAction SilentlyContinue | Out-Null
