param(
  [string]$ConfigPath = "./cmd/server/config/local.yaml",
  [string]$BaseUrl = "http://127.0.0.1:18080",
  [string]$MailpitUrl = "http://localhost:8025",
  [int]$TimeoutSeconds = 180
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

function Get-MailDetail([string]$MsgId) {
  try { return Invoke-RestMethod -Method GET -Uri "$MailpitUrl/api/v1/message/$MsgId" -TimeoutSec 5 -ErrorAction Stop } catch { return $null }
}

function Find-VerifyLink([string]$content) {
  if (-not $content) { return $null }
  $m = [regex]::Match($content, 'Link:\s*(https?://\S+)', 'IgnoreCase')
  if ($m.Success) { return $m.Groups[1].Value }
  $m2 = [regex]::Match($content, 'https?://\S*?/users/verify\?token=([A-Za-z0-9\-_%]+)')
  if ($m2.Success) { return $m2.Groups[0].Value }
  return $null
}

Write-Host "Starting server in background (full diagram E2E)..."
$repoRoot = (Get-Location).Path
$job = Start-Job -ScriptBlock { param($cfg,$wd) Set-Location $wd; go run main.go server --config $cfg } -ArgumentList $ConfigPath,$repoRoot
Start-Sleep -Seconds 2
if (-not (Wait-Port -HostName '127.0.0.1' -Port 18080 -Seconds 60)) {
  try { Receive-Job -Id $job.Id -Keep | Select-Object -Last 50 } catch {}
  try { Stop-Job -Id $job.Id | Out-Null } catch {}
  throw "HTTP gateway not listening on :18080 within timeout"
}
Write-Host "HTTP gateway is up."

$suffix = (Get-Random -Minimum 100000 -Maximum 999999)
$username = "user$suffix"
$password = "secret12"
$email = "${username}@example.com"

Write-Host "[1] Register user $username"
$payload = @{ username=$username; password=$password; email=$email; name='Full E2E'; cmnd='123456789'; birthday=1714608000; gender=$true; permanent_address='HN'; phone_number='0900000009' } | ConvertTo-Json
$reg = Invoke-RestMethod -Method Post -Uri "$BaseUrl/users" -ContentType 'application/json' -Body $payload -TimeoutSec 20
Write-Host "Register response: $($reg | ConvertTo-Json -Depth 4)"

Write-Host "[2] Trigger ResendVerification"
$resendBody = @{ email=$email } | ConvertTo-Json
try {
  $resendResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/users/verify/resend" -ContentType 'application/json' -Body $resendBody -TimeoutSec 20
  Write-Host "Resend response: $($resendResp | ConvertTo-Json -Depth 4)"
} catch {
  Write-Warning "Resend call failed: $($_.Exception.Message)"
}

Write-Host "[3] Poll Mailpit for two verification emails and pick RESEND"
$deadline = (Get-Date).AddSeconds($TimeoutSeconds)
$regUrl = $null
$resendUrl = $null
while ((Get-Date) -lt $deadline -and (-not $resendUrl)) {
  try {
    $list = Invoke-RestMethod -Method GET -Uri "$MailpitUrl/api/v1/messages?limit=50" -TimeoutSec 5 -ErrorAction Stop
    $msgs = $list.messages
    # Filter messages by recipient
    $mine = @()
    foreach ($m in $msgs) {
      $to = $m.to; $id = $m.id; if (-not $id) { $id = $m.ID }
      if ($to) {
        $toStr = ($to | ConvertTo-Json -Compress)
        if ($toStr -match [regex]::Escape($email)) { $mine += $m }
      }
    }
    if ($mine.Count -gt 0) {
      # sort newest first by Created or ID order
      $sorted = $mine | Sort-Object -Property Created -Descending
      foreach ($m in $sorted) {
        $id = $m.id; if (-not $id) { $id = $m.ID }
        $detail = Get-MailDetail $id
        if ($null -eq $detail) { continue }
        $content = if ($detail.Text) { [string]$detail.Text } else { [string]$detail.HTML }
        $link = Find-VerifyLink $content
        if (-not $link) { continue }
        if ($detail.Subject -match '(?i)resend') {
          $resendUrl = $link; break
        }
        if (-not $regUrl) { $regUrl = $link }
      }
      if (-not $resendUrl -and $regUrl -and $mine.Count -ge 2) {
        # Fallback: if subject didn't include 'resend', assume newest is resend
        $id = ($sorted | Select-Object -First 1).ID
        $detail = Get-MailDetail $id
        $content = if ($detail.Text) { [string]$detail.Text } else { [string]$detail.HTML }
        $resendUrl = Find-VerifyLink $content
      }
    }
  } catch {}
  if (-not $resendUrl) { Start-Sleep -Seconds 2 }
}

if (-not $resendUrl) {
  try { Receive-Job -Id $job.Id -Keep | Select-Object -Last 50 } catch {}
  try { Stop-Job -Id $job.Id | Out-Null } catch {}
  throw "Resend verification email not found in Mailpit within timeout"
}

Write-Host "Picked resend verification URL: $resendUrl"

Write-Host "[4] Verify using RESEND token"
$verify = Invoke-RestMethod -Method Get -Uri $resendUrl -TimeoutSec 20
Write-Host "Verify response: $($verify | ConvertTo-Json -Depth 6)"

if ($regUrl) {
  Write-Host "[4b] (Optional) Try verifying with original REGISTER token (should fail)"
  try {
    $null = Invoke-RestMethod -Method Get -Uri $regUrl -TimeoutSec 10
    Write-Warning "Unexpected: original token still valid"
  } catch {
    Write-Host "Original token rejected as expected: $($_.Exception.Message)"
  }
}

Write-Host "[5] Login after verified"
$loginBody = @{ username=$username; password=$password } | ConvertTo-Json
$login = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/user/login" -ContentType 'application/json' -Body $loginBody -TimeoutSec 20
Write-Host "Login response: $($login | ConvertTo-Json -Depth 6)"

Write-Host "Full diagram E2E (register→resend→verify→login) succeeded for $username"

try { Stop-Job -Id $job.Id | Out-Null } catch {}
Remove-Job -Id $job.Id -ErrorAction SilentlyContinue | Out-Null
