$ErrorActionPreference = "Stop"

$arch = "amd64"
$installDir = "$env:USERPROFILE\bin"
$exe = "zryl.exe"
$dest = "$installDir\$exe"

$uri = "https://github.com/HalxDocs/zryl-cl i/releases/latest/download/zryl-windows-amd64.exe"

Write-Host "Installing zryl..."

New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Invoke-WebRequest $uri -OutFile $dest

if ($env:PATH -notlike "*$installDir*") {
  setx PATH "$env:PATH;$installDir"
}

Write-Host "zryl installed successfully ✅"
Write-Host "Restart your terminal and run: zryl run npm run dev"