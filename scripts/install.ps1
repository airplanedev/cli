#!/usr/bin/env pwsh

# Inspired by:
# - https://github.com/denoland/deno_install/blob/master/install.ps1
#   Copyright 2018 the Deno authors. All rights reserved. MIT license.
# - https://github.com/superfly/flyctl/blob/master/installers/install.ps1

$ErrorActionPreference = 'Stop'

if ($args.Length -eq 1) {
  $Version = $args.Get(0)
}

$AirplaneInstall = $env:AIRPLANE_INSTALL
$BinDir = if ($AirplaneInstall) {
  "$AirplaneInstall\bin"
} else {
  "$Home\.airplane\bin"
}

$AirplaneArchive = "$BinDir\airplane.tar.gz"
$AirplaneExe = "$BinDir\airplane.exe"
$Target = 'windows_x86_64'

# GitHub requires TLS 1.2
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$DownloadURI = if (!$Version) {
  "https://github.com/airplanedev/cli/releases/latest/download/airplane_${Target}.tar.gz"
} else {
  "https://github.com/airplanedev/cli/releases/download/${Version}/airplane_${Target}.tar.gz"
}

if (!(Test-Path $BinDir)) {
  New-Item $BinDir -ItemType Directory | Out-Null
}

Invoke-WebRequest $DownloadURI -OutFile $AirplaneArchive -UseBasicParsing

Push-Location $BinDir
Remove-Item .\airplane.exe -ErrorAction SilentlyContinue
tar -xzf $AirplaneArchive
Pop-Location

Remove-Item $AirplaneArchive

$User = [EnvironmentVariableTarget]::User
$Path = [Environment]::GetEnvironmentVariable('Path', $User)
if (!(";$Path;".ToLower() -like "*;$BinDir;*".ToLower())) {
  [Environment]::SetEnvironmentVariable('Path', "$Path;$BinDir", $User)
  $Env:Path += ";$BinDir"
}

Write-Output "The Airplane CLI was installed successfully to $AirplaneExe"
Write-Output "Run 'airplane --help' to get started."
