param(
    [Parameter(Mandatory = $true)]
    [string]$OutputDir,
    [string]$WorkDir = "",
    [string]$Version = "1.23.1",
    [string]$MsysRoot = "C:\msys64"
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
if (-not $WorkDir) {
    $WorkDir = Join-Path $repoRoot "build\heif-tools"
}

$ucrtRoot = Join-Path $MsysRoot "ucrt64"
$cmake = Join-Path $ucrtRoot "bin\cmake.exe"
$ninja = Join-Path $ucrtRoot "bin\ninja.exe"
if (-not (Test-Path $cmake) -or -not (Test-Path $ninja)) {
    throw "MSYS2 UCRT64 CMake and Ninja are required to build HEIC tools."
}

$sourceDir = Join-Path $WorkDir "source"
$buildDir = Join-Path $WorkDir "build"
if (-not (Test-Path (Join-Path $sourceDir "CMakeLists.txt"))) {
    New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null
    & git clone --depth 1 --branch "v$Version" https://github.com/strukturag/libheif.git $sourceDir
    if ($LASTEXITCODE -ne 0) { throw "Failed to download libheif source." }
}

$env:PATH = "$(Join-Path $ucrtRoot 'bin');$(Join-Path $MsysRoot 'usr\bin');$env:PATH"
$configureArgs = @(
    "-S", $sourceDir,
    "-B", $buildDir,
    "-G", "Ninja",
    "-DCMAKE_BUILD_TYPE=Release",
    "-DBUILD_SHARED_LIBS=ON",
    "-DENABLE_PLUGIN_LOADING=OFF",
    "-DWITH_LIBDE265=OFF",
    "-DWITH_X265=OFF",
    "-DWITH_KVAZAAR=ON",
    "-DWITH_KVAZAAR_PLUGIN=OFF",
    "-DWITH_UVG266=OFF",
    "-DWITH_VVDEC=OFF",
    "-DWITH_VVENC=OFF",
    "-DWITH_X264=OFF",
    "-DWITH_OpenH264_DECODER=OFF",
    "-DWITH_AOM_DECODER=OFF",
    "-DWITH_AOM_ENCODER=OFF",
    "-DWITH_DAV1D=OFF",
    "-DWITH_SvtEnc=OFF",
    "-DWITH_RAV1E=OFF",
    "-DWITH_JPEG_DECODER=OFF",
    "-DWITH_JPEG_ENCODER=OFF",
    "-DWITH_OpenJPEG_ENCODER=OFF",
    "-DWITH_OpenJPEG_DECODER=OFF",
    "-DWITH_FFMPEG_DECODER=OFF",
    "-DWITH_OPENJPH_ENCODER=OFF",
    "-DWITH_UNCOMPRESSED_CODEC=OFF",
    "-DWITH_LIBSHARPYUV=OFF",
    "-DWITH_EXAMPLES=ON",
    "-DWITH_EXAMPLE_HEIF_THUMB=OFF",
    "-DWITH_EXAMPLE_HEIF_VIEW=OFF",
    "-DWITH_GDK_PIXBUF=OFF",
    "-DBUILD_DEVELOPMENT_TOOLS=OFF",
    "-DBUILD_DOCUMENTATION=OFF",
    "-DBUILD_TESTING=OFF",
    "-DCMAKE_DISABLE_FIND_PACKAGE_TIFF=ON",
    "-DCMAKE_DISABLE_FIND_PACKAGE_JPEG=ON",
    "-DCMAKE_DISABLE_FIND_PACKAGE_WEBP=ON"
)

& $cmake @configureArgs
if ($LASTEXITCODE -ne 0) { throw "Failed to configure minimal libheif." }
& $cmake --build $buildDir --target heif-enc --parallel
if ($LASTEXITCODE -ne 0) { throw "Failed to build minimal heif-enc." }

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
$files = @(
    (Join-Path $buildDir "examples\heif-enc.exe"),
    (Join-Path $buildDir "libheif\libheif.dll"),
    (Join-Path $ucrtRoot "bin\libgcc_s_seh-1.dll"),
    (Join-Path $ucrtRoot "bin\libstdc++-6.dll"),
    (Join-Path $ucrtRoot "bin\libwinpthread-1.dll"),
    (Join-Path $ucrtRoot "bin\libpng16-16.dll"),
    (Join-Path $ucrtRoot "bin\zlib1.dll"),
    (Join-Path $ucrtRoot "bin\libkvazaar-7.dll"),
    (Join-Path $ucrtRoot "bin\libcryptopp.dll")
)
foreach ($file in $files) {
    if (-not (Test-Path $file)) { throw "Missing HEIC runtime dependency: $file" }
    Copy-Item -LiteralPath $file -Destination $OutputDir -Force
}

$licenseDir = Join-Path $OutputDir "licenses"
New-Item -ItemType Directory -Force -Path $licenseDir | Out-Null
Copy-Item -LiteralPath (Join-Path $sourceDir "COPYING") -Destination (Join-Path $licenseDir "libheif.txt") -Force
$licensePackages = @("kvazaar", "crypto++", "libpng", "zlib", "gcc-libs")
foreach ($package in $licensePackages) {
    $packageDir = Join-Path $ucrtRoot "share\licenses\$package"
    if (Test-Path $packageDir) {
        Copy-Item -LiteralPath $packageDir -Destination (Join-Path $licenseDir $package) -Recurse -Force
    }
}
