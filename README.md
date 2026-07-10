# HexImg

HexImg 是一个基于 Qt 6 + QML 构建的原生图片格式转换工具，使用 FFmpeg 执行实际转换。界面支持批量选择、拖拽导入、实时日志、停止任务和深浅色主题。

## 功能

- 支持 JPG、PNG、WebP、AVIF、HEIC、GIF、ICO、SVG、BMP、TIFF 输出。
- JPG/WebP 支持 0-100 质量控制，PNG 支持 0-9 压缩级别。
- 支持添加后缀、输出到当前目录文件夹、替换源文件三种输出方式。
- 批量处理图片，预览输出路径，并实时显示 FFmpeg 日志。
- 使用 Qt Quick Controls 构建紧凑的 shadcn/ui 风格桌面界面。

## 项目结构

```text
src/
  main.cpp              # Qt 应用入口
  heximgbackend.*       # QML 后端、队列和 FFmpeg 转换流程
  queuemodel.*          # 图片队列模型
qml/
  Main.qml              # Qt Quick/QML 界面
  Shad*.qml             # shadcn/ui 风格基础控件
  FormatBadge.qml       # 队列格式标识
packaging/windows/
  HexImg.iss            # Inno Setup 安装脚本
scripts/
  package-windows.ps1   # CMake、windeployqt 和安装包构建
```

## 本地运行

需要安装 Qt 6.5+、CMake 3.21+、C++ 编译器、FFmpeg，并确保 `ffmpeg` 在 `PATH` 中。HEIC 输出还需要随程序部署 `tools\heif` 中的 libheif/Kvazaar 编码辅助程序。配置 `QT_ROOT_DIR` 指向 Qt 安装目录后运行：

```powershell
cmake -S . -B build -DCMAKE_PREFIX_PATH="$env:QT_ROOT_DIR"
cmake --build build --config Release
ctest --test-dir build --build-config Release --output-on-failure
```

使用 `windeployqt` 部署后，可运行程序及 Qt 运行库应放在同一目录。本地交付文件位于 `exe\HexImg.exe`。

## 构建 Windows 安装包

需要额外安装 Inno Setup 6：

```powershell
.\scripts\package-windows.ps1 -Version 0.2.0 -Arch x64
```

安装包输出到 `exe\HexImg-windows-x64-setup.exe`。

## License

Qt 以 LGPL 版本使用时，请随发行包提供对应许可证和 Qt 版权信息。
