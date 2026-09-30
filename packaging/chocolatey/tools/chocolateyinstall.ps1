$ErrorActionPreference = 'Stop'
# __URL__ and __CHECKSUM__ are filled in by .github/workflows/chocolatey.yml.
$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

Install-ChocolateyZipPackage `
  -PackageName $env:ChocolateyPackageName `
  -Url64bit '__URL__' `
  -UnzipLocation $toolsDir `
  -ChecksumType64 'sha256' `
  -Checksum64 '__CHECKSUM__'
