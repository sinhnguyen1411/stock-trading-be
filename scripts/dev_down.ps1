param(
  [string]$ComposeFile = "docker-compose.stack.yml"
)

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
  Write-Error "docker is required"
  exit 1
}

Write-Host "Stopping infrastructure via docker compose..."
& docker compose -f $ComposeFile down
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Done."

