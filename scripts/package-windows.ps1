param(
    [string]$Version = "0.1.0",
    [ValidateSet("x64", "arm64")]
    [string]$Arch = "x64"
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$distDir = Join-Path $repoRoot "dist\windows-$Arch"
New-Item -ItemType Directory -Force -Path $distDir | Out-Null

$goArch = if ($Arch -eq "x64") { "amd64" } else { "arm64" }
$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = $goArch

go build -trimpath -ldflags "-s -w -X main.version=$Version" -o (Join-Path $distDir "HexImg.exe") ./cmd/heximg

$iscc = Get-Command iscc -ErrorAction SilentlyContinue
if (-not $iscc) {
    Write-Warning "未找到 Inno Setup iscc，已生成 HexImg.exe，跳过安装包。"
    exit 0
}

$installer = Join-Path $repoRoot "packaging\windows\HexImg.iss"
$outDir = Join-Path $repoRoot "dist"
& $iscc.Source "/DAppVersion=$Version" "/DAppArch=$Arch" "/DSourceDir=$distDir" "/DOutputDir=$outDir" $installer
