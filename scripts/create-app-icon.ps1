param(
    [string]$InputPath = "",
    [string]$PngOutput = "",
    [string]$IcoOutput = "",
    [int]$Crop = 60
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
if (-not $InputPath) { $InputPath = Join-Path $repoRoot "logo.png" }
if (-not $PngOutput) { $PngOutput = Join-Path $repoRoot "assets\HexImg.png" }
if (-not $IcoOutput) { $IcoOutput = Join-Path $repoRoot "packaging\windows\HexImg.ico" }

Add-Type -AssemblyName System.Drawing

function New-ScaledBitmap {
    param(
        [System.Drawing.Image]$Source,
        [int]$Size
    )

    $bitmap = [System.Drawing.Bitmap]::new(
        $Size,
        $Size,
        [System.Drawing.Imaging.PixelFormat]::Format32bppArgb
    )
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    try {
        $graphics.Clear([System.Drawing.Color]::Transparent)
        $graphics.CompositingMode = [System.Drawing.Drawing2D.CompositingMode]::SourceCopy
        $graphics.CompositingQuality = [System.Drawing.Drawing2D.CompositingQuality]::HighQuality
        $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        $graphics.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
        $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
        $graphics.DrawImage($Source, 0, 0, $Size, $Size)
    } finally {
        $graphics.Dispose()
    }
    return $bitmap
}

$source = [System.Drawing.Bitmap]::FromFile((Resolve-Path $InputPath))
try {
    $sourceSize = [Math]::Min($source.Width, $source.Height) - (2 * $Crop)
    if ($sourceSize -le 0) { throw "Crop is larger than the source image." }

    $workSize = 2048
    $work = [System.Drawing.Bitmap]::new(
        $workSize,
        $workSize,
        [System.Drawing.Imaging.PixelFormat]::Format32bppArgb
    )
    try {
        $graphics = [System.Drawing.Graphics]::FromImage($work)
        $path = [System.Drawing.Drawing2D.GraphicsPath]::new()
        try {
            $graphics.Clear([System.Drawing.Color]::Transparent)
            $graphics.CompositingMode = [System.Drawing.Drawing2D.CompositingMode]::SourceCopy
            $graphics.CompositingQuality = [System.Drawing.Drawing2D.CompositingQuality]::HighQuality
            $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
            $graphics.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
            $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
            $path.AddEllipse(2, 2, $workSize - 4, $workSize - 4)
            $graphics.SetClip($path)

            $sourceRect = [System.Drawing.Rectangle]::new($Crop, $Crop, $sourceSize, $sourceSize)
            $targetRect = [System.Drawing.Rectangle]::new(0, 0, $workSize, $workSize)
            $graphics.DrawImage($source, $targetRect, $sourceRect, [System.Drawing.GraphicsUnit]::Pixel)
        } finally {
            $path.Dispose()
            $graphics.Dispose()
        }

        $pngDir = Split-Path -Parent $PngOutput
        $icoDir = Split-Path -Parent $IcoOutput
        New-Item -ItemType Directory -Force -Path $pngDir, $icoDir | Out-Null

        $master = New-ScaledBitmap -Source $work -Size 1024
        try {
            $master.Save($PngOutput, [System.Drawing.Imaging.ImageFormat]::Png)

            $sizes = @(16, 24, 32, 48, 64, 128, 256)
            $frames = foreach ($size in $sizes) {
                $frame = New-ScaledBitmap -Source $master -Size $size
                $stream = [System.IO.MemoryStream]::new()
                try {
                    $frame.Save($stream, [System.Drawing.Imaging.ImageFormat]::Png)
                    ,$stream.ToArray()
                } finally {
                    $stream.Dispose()
                    $frame.Dispose()
                }
            }

            $file = [System.IO.File]::Open($IcoOutput, [System.IO.FileMode]::Create)
            $writer = [System.IO.BinaryWriter]::new($file)
            try {
                $writer.Write([uint16]0)
                $writer.Write([uint16]1)
                $writer.Write([uint16]$sizes.Count)

                $offset = 6 + (16 * $sizes.Count)
                for ($i = 0; $i -lt $sizes.Count; $i++) {
                    $dimension = if ($sizes[$i] -eq 256) { 0 } else { $sizes[$i] }
                    $writer.Write([byte]$dimension)
                    $writer.Write([byte]$dimension)
                    $writer.Write([byte]0)
                    $writer.Write([byte]0)
                    $writer.Write([uint16]1)
                    $writer.Write([uint16]32)
                    $writer.Write([uint32]$frames[$i].Length)
                    $writer.Write([uint32]$offset)
                    $offset += $frames[$i].Length
                }

                foreach ($frameBytes in $frames) {
                    $writer.Write($frameBytes)
                }
            } finally {
                $writer.Dispose()
                $file.Dispose()
            }
        } finally {
            $master.Dispose()
        }
    } finally {
        $work.Dispose()
    }
} finally {
    $source.Dispose()
}

