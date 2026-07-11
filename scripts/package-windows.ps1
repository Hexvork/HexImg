param(
    [string]$Version = "0.1.0",
    [ValidateSet("x64", "arm64")]
    [string]$Arch = "x64",
    [string]$MsysRoot = ""
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$exeDir = Join-Path $repoRoot "exe"
$buildDir = Join-Path $repoRoot "build\qt-release-$Arch"
$deployDir = Join-Path $buildDir "deploy"
New-Item -ItemType Directory -Force -Path $exeDir | Out-Null
& (Join-Path $PSScriptRoot "create-app-icon.ps1")

$cmake = Get-Command cmake -ErrorAction SilentlyContinue
if (-not $cmake) {
    throw "cmake was not found. Install CMake 3.21 or newer."
}

$qtPrefix = $env:QT_ROOT_DIR
if (-not $qtPrefix) {
    $qtCommand = Get-Command qtpaths6 -ErrorAction SilentlyContinue
    if (-not $qtCommand) { $qtCommand = Get-Command qtpaths -ErrorAction SilentlyContinue }
    if ($qtCommand) {
        $qtPrefix = (& $qtCommand.Source --install-prefix).Trim()
    }
}

if (-not $qtPrefix) {
    throw "Qt 6 was not found. Set QT_ROOT_DIR or add qtpaths to PATH."
}

$configureArgs = @(
    "-S", $repoRoot,
    "-B", $buildDir,
    "-DCMAKE_BUILD_TYPE=Release",
    "-DCMAKE_PREFIX_PATH=$qtPrefix"
)
if ($Arch -eq "arm64") {
    $configureArgs += @("-A", "ARM64")
}
& $cmake.Source @configureArgs
if ($LASTEXITCODE -ne 0) { throw "CMake configure failed." }

& $cmake.Source --build $buildDir --config Release --parallel
if ($LASTEXITCODE -ne 0) { throw "Qt build failed." }

$binaryCandidates = @(
    (Join-Path $buildDir "Release\HexImg.exe"),
    (Join-Path $buildDir "HexImg.exe")
)
$binary = $binaryCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $binary) {
    throw "Qt executable was not found in $buildDir"
}

$binaryDir = Split-Path -Parent $binary
$heifToolsDir = Join-Path $binaryDir "tools\heif"
if ($Arch -eq "x64") {
    $heifArgs = @{
        OutputDir = $heifToolsDir
        WorkDir = Join-Path $buildDir "heif-tools"
    }
    if ($MsysRoot) { $heifArgs.MsysRoot = $MsysRoot }
    & (Join-Path $PSScriptRoot "build-heif-tools.ps1") @heifArgs
}

$canRunTarget = $Arch -eq "x64" -or $env:PROCESSOR_ARCHITECTURE -eq "ARM64"
if ($canRunTarget) {
    $ctestPath = Join-Path (Split-Path -Parent $cmake.Source) "ctest.exe"
    if (-not (Test-Path $ctestPath)) { throw "ctest was not found next to cmake." }
    & $ctestPath --test-dir $buildDir --build-config Release --output-on-failure
    if ($LASTEXITCODE -ne 0) { throw "Qt tests failed." }
}

New-Item -ItemType Directory -Force -Path $deployDir | Out-Null
Copy-Item $binary (Join-Path $deployDir "HexImg.exe") -Force
if (Test-Path $heifToolsDir) {
    $deployedHeifDir = Join-Path $deployDir "tools\heif"
    New-Item -ItemType Directory -Force -Path $deployedHeifDir | Out-Null
    Copy-Item -Path (Join-Path $heifToolsDir "*") -Destination $deployedHeifDir -Recurse -Force
}

$windeployqtPath = $null
$windeployqt = Get-Command windeployqt6 -ErrorAction SilentlyContinue
if (-not $windeployqt) { $windeployqt = Get-Command windeployqt -ErrorAction SilentlyContinue }
if ($windeployqt) {
    $windeployqtPath = $windeployqt.Source
}
if (-not $windeployqtPath) {
    $knownWindeployqt = Join-Path $qtPrefix "bin\windeployqt.exe"
    if (Test-Path $knownWindeployqt) { $windeployqtPath = $knownWindeployqt }
}
if (-not $windeployqtPath) { throw "windeployqt was not found in the Qt installation." }

& $windeployqtPath --release --no-translations --qmldir (Join-Path $repoRoot "qml") (Join-Path $deployDir "HexImg.exe")
if ($LASTEXITCODE -ne 0) { throw "windeployqt failed." }

$iscc = Get-Command iscc -ErrorAction SilentlyContinue
if (-not $iscc) {
    $knownIscc = @(
        "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
        "C:\Program Files\Inno Setup 6\ISCC.exe"
    ) | Where-Object { Test-Path $_ } | Select-Object -First 1

    if (-not $knownIscc) {
        throw "Inno Setup iscc was not found. Install Inno Setup 6 and retry."
    }
    $isccPath = $knownIscc
} else {
    $isccPath = $iscc.Source
}

$installer = Join-Path $repoRoot "packaging\windows\HexImg.iss"
& $isccPath "/DAppVersion=$Version" "/DAppArch=$Arch" "/DSourceDir=$deployDir" "/DOutputDir=$exeDir" $installer
