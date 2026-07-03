# HexImg

HexImg 是一个使用 Go + Fyne 编写的原生跨平台 FFmpeg 图形化工具。项目不使用 Electron 或其他 Web 套壳，界面配色采用 fluent dark/light 风格，并通过 `internal/fas` 调用 Font Awesome Solid 图标资源。

## 功能

- 原生 Fyne 桌面界面，支持深色/浅色主题。
- 选择输入/输出文件，配置格式、质量、预设、缩放和额外 FFmpeg 参数。
- 实时生成 FFmpeg 命令预览。
- 调用系统 `ffmpeg` 执行任务并输出运行日志。
- GitHub Actions 自动构建并发布 release 产物。

## 本地运行

需要先安装 Go 和 FFmpeg，并确保 `ffmpeg` 在 `PATH` 中。

```powershell
go mod tidy
go run ./cmd/heximg
```

## 构建

```powershell
go build -trimpath -ldflags "-s -w" -o dist/HexImg.exe ./cmd/heximg
```

Windows 安装包可使用 Inno Setup：

```powershell
.\scripts\package-windows.ps1 -Version 0.1.0 -Arch x64
```

## Release

`.github/workflows/release.yml` 支持：

- `workflow_dispatch` 手动输入版本号发布。
- 推送 `v*` tag 自动发布。

目标平台：

- Windows x64
- Windows ARM64
- macOS Apple Silicon
- macOS Intel
- Linux x86_64
- Linux ARM64

Windows 产物包含原生安装 `.exe`，同时保留直接运行的二进制 `.exe`。

## 图标授权

界面图标路径数据来自 Font Awesome Free Solid，遵循 Font Awesome Free 授权条款。
