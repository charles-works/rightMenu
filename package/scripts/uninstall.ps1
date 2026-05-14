$ErrorActionPreference = 'Stop'
$exe = Join-Path $PSScriptRoot '..\rightmenu.exe'
if (-not (Test-Path $exe)) {
  $exe = Join-Path (Get-Location) 'rightmenu.exe'
}
& $exe uninstall
