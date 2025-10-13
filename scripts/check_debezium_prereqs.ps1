param(
  [string]$DBHost = "host.docker.internal",
  [int]$DBPort = 3306,
  [string]$DBUser = "root",
  [string]$DBPassword = "Ngdms1107#"
)

$ErrorActionPreference = 'Stop'

function RunMySQL([string]$sql) {
  $args = @('run','--rm','-e',"MYSQL_PWD=$DBPassword",'mysql:8','mysql',
    "--host=$DBHost","--port=$DBPort","--user=$DBUser","-N","-e", $sql)
  $out = & docker @args
  if ($LASTEXITCODE -ne 0) { throw "docker mysql failed" }
  return $out
}

Write-Host "Checking MySQL binlog settings via docker mysql client..."
$vars = RunMySQL "SELECT @@log_bin, @@binlog_format, @@server_id, @@binlog_row_image;"
Write-Host "@@log_bin, @@binlog_format, @@server_id, @@binlog_row_image:" $vars

if ($vars -notmatch '^ON' -and $vars -notmatch '^1') {
  Write-Warning "Binary logging appears disabled (log_bin != ON). Debezium will not stream outbox events."
  Write-Host "Enable in your MySQL config (my.cnf):"
  Write-Host "  server_id=5401"
  Write-Host "  log_bin=binlog"
  Write-Host "  binlog_format=ROW"
  Write-Host "  binlog_row_image=FULL"
  Write-Host "Then restart MySQL and re-run dev_up."
}

