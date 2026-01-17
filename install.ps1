$arch = "amd64"
$url = "https://github.com/HalxDocs/zryl/releases/latest/download/zryl-windows-amd64.exe"

$installDir = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$dest = "$installDir\zryl.exe"
Invoke-WebRequest $url -OutFile $dest

if ($env:PATH -notlike "*$installDir*") {
  setx PATH "$env:PATH;$installDir"
}

Write-Host "zryl installed successfully 🎉"
Write-Host "Restart your terminal and run: zryl run npm run dev"
