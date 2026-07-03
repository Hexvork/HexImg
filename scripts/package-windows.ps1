param(
    [string]$Version = "0.1.0",
    [ValidateSet("x64", "arm64")]
    [string]$Arch = "x64"
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$distDir = Join-Path $repoRoot "dist\windows-$Arch"
New-Item -ItemType Directory -Force -Path $distDir | Out-Null

if ($Arch -eq "x64") {
    $goArch = "amd64"
    $defaultCC = "gcc"
    $compilerCandidates = @(
        "C:\msys64\ucrt64\bin\gcc.exe",
        "C:\msys64\mingw64\bin\gcc.exe"
    )
} else {
    $goArch = "arm64"
    $defaultCC = "clang"
    $compilerCandidates = @(
        "C:\msys64\clangarm64\bin\clang.exe"
    )
}

$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = $goArch
$env:CC = $defaultCC

if (-not (Get-Command $env:CC -ErrorAction SilentlyContinue)) {
    foreach ($candidate in $compilerCandidates) {
        if (Test-Path $candidate) {
            $compilerDir = Split-Path -Parent $candidate
            $env:PATH = "$compilerDir;$env:PATH"
            break
        }
    }
}

if ($Arch -eq "arm64") {
    $env:CGO_CFLAGS = "-DWINBOOL=BOOL -Wno-strict-prototypes"
    if (Test-Path "C:\msys64\clangarm64\lib") {
        $env:CGO_LDFLAGS = "-LC:\msys64\clangarm64\lib"
    }
}

if (-not (Get-Command $env:CC -ErrorAction SilentlyContinue)) {
    throw "Missing $($env:CC). Install the C compiler for Windows $Arch and add it to PATH."
}

try {
    & $env:CC --version | Out-Null
} catch {
    throw "$($env:CC) is not executable on this host. Install a compiler that can run here for Windows $Arch."
}

$ldflags = "-s -w -H=windowsgui -X main.version=$Version"
$goArgs = @(
    "build",
    "-trimpath",
    "-ldflags",
    $ldflags,
    "-o",
    (Join-Path $distDir "HexImg.exe"),
    "./cmd/heximg"
)
& go @goArgs

$iscc = Get-Command iscc -ErrorAction SilentlyContinue
if (-not $iscc) {
    Write-Warning "Inno Setup iscc was not found. HexImg.exe was built, installer packaging skipped."
    exit 0
}

$installer = Join-Path $repoRoot "packaging\windows\HexImg.iss"
$outDir = Join-Path $repoRoot "dist"
& $iscc.Source "/DAppVersion=$Version" "/DAppArch=$Arch" "/DSourceDir=$distDir" "/DOutputDir=$outDir" $installer
