param(
  [string]$ConfigPath = "./cmd/server/config/local.yaml",
  [string]$BaseUrl = "http://127.0.0.1:18080",
  [string]$MailpitUrl = "http://localhost:8025",
  [int]$TimeoutSeconds = 120
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

function Get-MailText([string]$MsgId) {
  try {
    $obj = Invoke-RestMethod -Method GET -Uri "$MailpitUrl/api/v1/message/$MsgId" -TimeoutSec 5 -ErrorAction Stop
    if ($null -ne $obj.Text -and $obj.Text -ne '') { return [string]$obj.Text }
    if ($null -ne $obj.HTML -and $obj.HTML -ne '') { return [string]$obj.HTML }
  } catch { }
  return ""
}

function Find-VerifyLink([string]$content) {
  if (-not $content) { return $null }
  $m = [regex]::Match($content, 'Link:\s*(https?://\S+)', 'IgnoreCase')
  if ($m.Success) { return $m.Groups[1].Value }
  $m2 = [regex]::Match($content, 'https?://\S*?/users/verify\?token=([A-Za-z0-9\-_%]+)')
  if ($m2.Success) { return $m2.Groups[0].Value }
  return $null
}

Write-Host "Starting server in background..."
$repoRoot = (Get-Location).Path
$job = Start-Job -ScriptBlock { param($cfg,$wd) Set-Location $wd; go run main.go server --config $cfg } -ArgumentList $ConfigPath,$repoRoot
Start-Sleep -Seconds 2

if (-not (Wait-Port -HostName '127.0.0.1' -Port 18080 -Seconds 30)) {
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
$payload = @{ username=$username; password=$password; email=$email; name='Mailpit E2E'; cmnd='123456789'; birthday=1714608000; gender=$true; permanent_address='HN'; phone_number='0900000009' } | ConvertTo-Json
$reg = Invoke-RestMethod -Method Post -Uri "$BaseUrl/users" -ContentType 'application/json' -Body $payload -TimeoutSec 20
Write-Host "Register response: $($reg | ConvertTo-Json -Depth 4)"

Write-Host "Polling Mailpit for verification email..."
$deadline = (Get-Date).AddSeconds($TimeoutSeconds)
$verifyUrl = $null
while ((Get-Date) -lt $deadline -and -not $verifyUrl) {
  try {
    $list = Invoke-RestMethod -Method GET -Uri "$MailpitUrl/api/v1/messages?limit=50" -TimeoutSec 5 -ErrorAction Stop
    $msgs = $list.messages
    foreach ($m in $msgs) {
      # Match recipient if available
      $to = $m.to
      $id = $m.id; if (-not $id) { $id = $m.ID }
      $matchRecipient = $false
      if ($to) {
        $toStr = ($to | ConvertTo-Json -Compress)
        if ($toStr -match [regex]::Escape($email)) { $matchRecipient = $true }
      }
      if (-not $matchRecipient) { continue }
      $content = Get-MailText $id
      $verifyUrl = Find-VerifyLink $content
      if ($verifyUrl) { break }
    }
  } catch { }
  if (-not $verifyUrl) { Start-Sleep -Seconds 2 }
}

if (-not $verifyUrl) {
  try { Receive-Job -Id $job.Id -Keep | Select-Object -Last 50 } catch {}
  try { Stop-Job -Id $job.Id | Out-Null } catch {}
  throw "Verification email not found in Mailpit within timeout"
}

Write-Host "Verification URL: $verifyUrl"
Write-Host "Calling verify link..."
$verify = Invoke-RestMethod -Method Get -Uri $verifyUrl -TimeoutSec 20
Write-Host "Verify response: $($verify | ConvertTo-Json -Depth 6)"

Write-Host "Login after verifying..."
$loginBody = @{ username=$username; password=$password } | ConvertTo-Json
$login = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/user/login" -ContentType 'application/json' -Body $loginBody -TimeoutSec 20
Write-Host "Login response: $($login | ConvertTo-Json -Depth 6)"

Write-Host "Mailpit E2E success for $username"

try { Stop-Job -Id $job.Id | Out-Null } catch {}
Remove-Job -Id $job.Id -ErrorAction SilentlyContinue | Out-Null
