$ErrorActionPreference = "Stop"

$Repo = "HalxDocs/zryl-cli"
$InstallDir = "$env:USERPROFILE\bin"
$Exe = "zryl.exe"
$Dest = "$InstallDir\$Exe"

$Arch = "amd64"
$Uri = "https://github.com/$Repo/releases/latest/download/zryl-windows-$Arch.exe"

Write-Host "Installing zryl..."

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

Invoke-WebRequest $Uri -OutFile $Dest

if ($env:PATH -notlike "*$InstallDir*") {
  setx PATH "$env:PATH;$InstallDir" | Out-Null
}

Write-Host "zryl installed successfully ✅"
Write-Host "Restart your terminal and run:"
Write-Host "  zryl run npm run dev"