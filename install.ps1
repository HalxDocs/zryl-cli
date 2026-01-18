$ErrorActionPreference = "Stop"

$repo = "HalxDocs/zryl-cli"
$installDir = "$env:USERPROFILE\bin"
$exeName = "zryl.exe"
$dest = "$installDir\$exeName"

$asset = "zryl-windows-amd64.exe"
$uri = "https://github.com/$repo/releases/latest/download/$asset"

Write-Host "Installing zryl..." -ForegroundColor Cyan

# Create bin dir
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

# Download binary
Invoke-WebRequest $uri -OutFile $dest

# Add to PATH if needed
if ($env:PATH -notlike "*$installDir*") {
  setx PATH "$env:PATH;$installDir" | Out-Null
}

Write-Host "zryl installed successfully ✅" -ForegroundColor Green
Write-Host "Restart your terminal and run:" -ForegroundColor Yellow
Write-Host "  zryl run npm run dev"