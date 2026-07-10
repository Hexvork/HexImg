#include "heximgbackend.h"

#include <QCoreApplication>
#include <QDir>
#include <QFile>
#include <QFileDialog>
#include <QFileInfo>
#include <QImage>
#include <QPainter>
#include <QRegularExpression>
#include <QSaveFile>
#include <QStandardPaths>
#include <QSvgRenderer>
#include <QTemporaryFile>

#ifdef Q_OS_WIN
#include <objbase.h>
#include <wincodec.h>
#endif

namespace {
const QStringList formats = {
    QStringLiteral("jpg"), QStringLiteral("png"), QStringLiteral("webp"), QStringLiteral("avif"),
    QStringLiteral("heic"), QStringLiteral("gif"), QStringLiteral("ico"), QStringLiteral("svg"),
    QStringLiteral("bmp"), QStringLiteral("tiff"),
};
const QStringList imageExtensions = {
    QStringLiteral("jpg"), QStringLiteral("jpeg"), QStringLiteral("png"), QStringLiteral("webp"),
    QStringLiteral("avif"), QStringLiteral("heic"), QStringLiteral("heif"), QStringLiteral("gif"),
    QStringLiteral("ico"), QStringLiteral("svg"), QStringLiteral("bmp"), QStringLiteral("tif"),
    QStringLiteral("tiff"),
};
}

HexImgBackend::HexImgBackend(QObject *parent)
    : QObject(parent)
{
    m_ffmpegAvailable = !QStandardPaths::findExecutable(QStringLiteral("ffmpeg")).isEmpty();
}

HexImgBackend::~HexImgBackend()
{
    if (m_process) {
        m_process->kill();
        m_process->waitForFinished(1000);
    }
    cleanupPreparedOutput(true);
}

QueueModel *HexImgBackend::queueModel() { return &m_queueModel; }
int HexImgBackend::imageCount() const { return m_queueModel.rowCount(); }
bool HexImgBackend::ffmpegAvailable() const { return m_ffmpegAvailable; }
bool HexImgBackend::converting() const { return m_converting; }
bool HexImgBackend::darkTheme() const { return m_darkTheme; }
QString HexImgBackend::format() const { return m_format; }
int HexImgBackend::quality() const { return m_quality; }
int HexImgBackend::pngCompression() const { return m_pngCompression; }
int HexImgBackend::outputMode() const { return m_outputMode; }
QString HexImgBackend::suffix() const { return m_suffix; }
QString HexImgBackend::folderName() const { return m_folderName; }
QString HexImgBackend::previewOutput() const
{
    if (m_queueModel.items().isEmpty()) return {};
    bool ignored = false;
    ConversionSettings settings{m_format, m_quality, m_pngCompression, m_outputMode, m_suffix, m_folderName};
    return outputFor(m_queueModel.items().first().path, settings, &ignored);
}
QString HexImgBackend::status() const { return m_status; }
QStringList HexImgBackend::logs() const { return m_logs; }

void HexImgBackend::setDarkTheme(bool value)
{
    if (m_darkTheme == value) return;
    m_darkTheme = value;
    emit darkThemeChanged();
}

void HexImgBackend::setFormat(const QString &value)
{
    const QString normalized = value.toLower();
    if (!formats.contains(normalized) || m_format == normalized) return;
    m_format = normalized;
    rebuildModel();
    emit settingsChanged();
}

void HexImgBackend::setQuality(int value)
{
    const int clamped = clampQuality(value);
    if (m_quality == clamped) return;
    m_quality = clamped;
    emit settingsChanged();
}

void HexImgBackend::setPngCompression(int value)
{
    const int clamped = qBound(0, value, 9);
    if (m_pngCompression == clamped) return;
    m_pngCompression = clamped;
    emit settingsChanged();
}

void HexImgBackend::setOutputMode(int value)
{
    const int clamped = qBound(0, value, 2);
    if (m_outputMode == clamped) return;
    m_outputMode = clamped;
    rebuildModel();
    emit settingsChanged();
}

void HexImgBackend::setSuffix(const QString &value)
{
    if (m_suffix == value) return;
    m_suffix = value;
    rebuildModel();
    emit settingsChanged();
}

void HexImgBackend::setFolderName(const QString &value)
{
    if (m_folderName == value) return;
    m_folderName = value;
    rebuildModel();
    emit settingsChanged();
}

void HexImgBackend::chooseImages()
{
    if (m_converting) return;
    const QStringList paths = QFileDialog::getOpenFileNames(
        nullptr,
        QStringLiteral("选择图片"),
        {},
        QStringLiteral("图片文件 (*.jpg *.jpeg *.png *.webp *.avif *.heic *.heif *.gif *.ico *.svg *.bmp *.tif *.tiff)"));
    for (const QString &path : paths) addPath(path);
    if (!paths.isEmpty()) setStatus(QStringLiteral("已选择 %1 张图片").arg(m_queueModel.rowCount()));
}

void HexImgBackend::addDroppedPath(const QString &urlOrPath)
{
    if (m_converting) return;
    QString path = QUrl(urlOrPath).toLocalFile();
    if (path.isEmpty()) path = urlOrPath;
    addPath(path);
    if (m_queueModel.rowCount() > 0) setStatus(QStringLiteral("已选择 %1 张图片").arg(m_queueModel.rowCount()));
}

void HexImgBackend::addPath(const QString &path)
{
    const QString normalized = normalizePath(path);
    if (normalized.isEmpty() || !isImagePath(normalized)) return;
    for (const ImageItem &item : m_queueModel.items()) {
        if (samePath(item.path, normalized)) return;
    }
    QVector<ImageItem> items = m_queueModel.items();
    ConversionSettings settings{m_format, m_quality, m_pngCompression, m_outputMode, m_suffix, m_folderName};
    bool ignored = false;
    items.append({normalized, fileName(normalized), outputFor(normalized, settings, &ignored)});
    m_queueModel.setItems(items);
    emit queueChanged();
    emit settingsChanged();
}

void HexImgBackend::removeImage(int index)
{
    if (m_converting || index < 0 || index >= m_queueModel.rowCount()) return;
    QVector<ImageItem> items = m_queueModel.items();
    items.removeAt(index);
    m_queueModel.setItems(items);
    emit queueChanged();
    if (items.isEmpty()) setStatus(QStringLiteral("就绪"));
    emit settingsChanged();
}

void HexImgBackend::clearImages()
{
    if (m_converting) return;
    m_queueModel.setItems(QVector<ImageItem>());
    emit queueChanged();
    setStatus(QStringLiteral("就绪"));
    emit settingsChanged();
}

void HexImgBackend::startConversion()
{
    if (m_converting || m_queueModel.items().isEmpty()) return;
    if (!m_ffmpegAvailable) {
        setStatus(QStringLiteral("未检测到 FFmpeg，请先安装并加入 PATH"));
        appendLog(QStringLiteral("错误：未找到 ffmpeg，请先安装 FFmpeg 并加入 PATH"));
        return;
    }
    m_conversionItems = m_queueModel.items();
    m_conversionSettings = {m_format, m_quality, m_pngCompression, m_outputMode, m_suffix, m_folderName};
    m_conversionIndex = -1;
    m_failedCount = 0;
    m_cancelRequested = false;
    m_converting = true;
    m_logs.clear();
    emit logsChanged();
    emit convertingChanged();
    setStatus(QStringLiteral("正在转换..."));
    beginNextConversion();
}

void HexImgBackend::cancelConversion()
{
    if (!m_converting) return;
    m_cancelRequested = true;
    setStatus(QStringLiteral("正在停止转换..."));
    if (m_process) m_process->kill();
}

void HexImgBackend::beginNextConversion()
{
    ++m_conversionIndex;
    if (m_cancelRequested || m_conversionIndex >= m_conversionItems.size()) {
        cleanupPreparedOutput(m_cancelRequested);
        m_converting = false;
        emit convertingChanged();
        if (m_cancelRequested) {
            setStatus(QStringLiteral("已停止"));
        } else if (m_failedCount > 0) {
            setStatus(QStringLiteral("转换完成，%1 个失败").arg(m_failedCount));
        } else {
            setStatus(QStringLiteral("转换完成"));
        }
        return;
    }

    const ImageItem &item = m_conversionItems.at(m_conversionIndex);
    QString error;
    m_preparedOutput = {};
    if (!prepareOutput(item.path, m_conversionSettings, &m_preparedOutput, &error)) {
        ++m_failedCount;
        appendLog(QStringLiteral("[%1/%2] %3 准备失败：%4").arg(m_conversionIndex + 1).arg(m_conversionItems.size()).arg(item.fileName, error));
        beginNextConversion();
        return;
    }

    setStatus(QStringLiteral("正在转换 %1/%2：%3").arg(m_conversionIndex + 1).arg(m_conversionItems.size()).arg(item.fileName));
    m_process = new QProcess(this);
    m_process->setProcessChannelMode(QProcess::MergedChannels);
    connect(m_process, &QProcess::readyReadStandardOutput, this, &HexImgBackend::readProcessOutput);
    connect(m_process, &QProcess::errorOccurred, this, &HexImgBackend::processError);
    connect(m_process, qOverload<int, QProcess::ExitStatus>(&QProcess::finished), this, &HexImgBackend::processFinished);
    m_process->start(QStringLiteral("ffmpeg"), ffmpegArguments(m_preparedOutput.processInput, m_preparedOutput, m_conversionSettings));
}

void HexImgBackend::readProcessOutput()
{
    if (!m_process) return;
    const QString output = QString::fromLocal8Bit(m_process->readAllStandardOutput());
    const QStringList lines = output.split(QRegularExpression(QStringLiteral("[\\r\\n]")), Qt::SkipEmptyParts);
    for (const QString &line : lines) appendLog(line);
}

void HexImgBackend::processError(QProcess::ProcessError error)
{
    if (error != QProcess::FailedToStart || !m_process) return;
    const QString input = m_conversionItems.value(m_conversionIndex).path;
    const QString name = fileName(input);
    const QString message = m_process->errorString();
    QProcess *process = m_process;
    m_process = nullptr;
    process->deleteLater();
    cleanupPreparedOutput(true);
    ++m_failedCount;
    appendLog(QStringLiteral("[%1/%2] %3 启动 FFmpeg 失败：%4").arg(m_conversionIndex + 1).arg(m_conversionItems.size()).arg(name, message));
    beginNextConversion();
}

void HexImgBackend::processFinished(int exitCode, QProcess::ExitStatus exitStatus)
{
    if (!m_process) return;
    readProcessOutput();
    const QString input = m_conversionItems.value(m_conversionIndex).path;
    const QString name = m_conversionItems.value(m_conversionIndex).fileName;
    QProcess *process = m_process;
    m_process = nullptr;
    process->deleteLater();

    if (m_cancelRequested) {
        cleanupPreparedOutput(true);
        beginNextConversion();
        return;
    }

    if (exitStatus != QProcess::NormalExit || exitCode != 0) {
        cleanupPreparedOutput(true);
        ++m_failedCount;
        appendLog(QStringLiteral("[%1/%2] %3 转换失败，FFmpeg 退出码：%4").arg(m_conversionIndex + 1).arg(m_conversionItems.size()).arg(name).arg(exitCode));
        beginNextConversion();
        return;
    }

    QString error;
    if (!postProcessOutput(m_preparedOutput, m_conversionSettings, &error)) {
        cleanupPreparedOutput(true);
        ++m_failedCount;
        appendLog(QStringLiteral("[%1/%2] %3 后处理失败：%4").arg(m_conversionIndex + 1).arg(m_conversionItems.size()).arg(name, error));
        beginNextConversion();
        return;
    }
    if (!finalizeOutput(input, m_preparedOutput, &error)) {
        ++m_failedCount;
        appendLog(QStringLiteral("[%1/%2] %3 保存失败：%4").arg(m_conversionIndex + 1).arg(m_conversionItems.size()).arg(name, error));
    } else {
        appendLog(QStringLiteral("[%1/%2] 完成：%3").arg(m_conversionIndex + 1).arg(m_conversionItems.size()).arg(fileName(m_preparedOutput.output)));
    }
    cleanupPreparedOutput(false);
    beginNextConversion();
}

void HexImgBackend::rebuildModel()
{
    QVector<ImageItem> items;
    ConversionSettings settings{m_format, m_quality, m_pngCompression, m_outputMode, m_suffix, m_folderName};
    for (const ImageItem &old : m_queueModel.items()) {
        bool ignored = false;
        items.append({old.path, old.fileName, outputFor(old.path, settings, &ignored)});
    }
    m_queueModel.setItems(items);
    emit queueChanged();
}

void HexImgBackend::setStatus(const QString &value)
{
    if (m_status == value) return;
    m_status = value;
    emit statusChanged();
}

void HexImgBackend::appendLog(const QString &line)
{
    if (line.trimmed().isEmpty()) return;
    m_logs.append(line);
    while (m_logs.size() > 240) m_logs.removeFirst();
    emit logsChanged();
}

bool HexImgBackend::prepareOutput(const QString &input, const ConversionSettings &settings, PreparedOutput *prepared, QString *error)
{
    bool replaceSource = false;
    const QString output = outputFor(input, settings, &replaceSource);
    if (output.isEmpty()) {
        *error = QStringLiteral("输出路径不可用");
        return false;
    }
    if (!QDir().mkpath(parentPath(output))) {
        *error = QStringLiteral("创建输出目录失败");
        return false;
    }
    prepared->output = output;
    prepared->replaceSource = replaceSource;
    prepared->processInput = input;

    if (replaceSource && !samePath(input, output) && QFile::exists(output)) {
        *error = QStringLiteral("输出文件已存在，未覆盖：%1").arg(output);
        return false;
    }

    if (QFileInfo(input).suffix().compare(QStringLiteral("svg"), Qt::CaseInsensitive) == 0) {
        prepared->processInput = reserveTemporaryPath(QDir::tempPath(), QStringLiteral("png"), error);
        prepared->processInputTemporary = !prepared->processInput.isEmpty();
        if (prepared->processInput.isEmpty() || !renderSvgInput(input, prepared->processInput, error)) {
            cleanupPreparedOutput();
            return false;
        }
    }

    const bool requiresPostProcess = settings.format == QStringLiteral("svg") || settings.format == QStringLiteral("heic");
    if (requiresPostProcess) {
        prepared->processOutput = reserveTemporaryPath(parentPath(output), QStringLiteral("png"), error);
        if (prepared->processOutput.isEmpty()) {
            cleanupPreparedOutput();
            return false;
        }
    }

    if (replaceSource) {
        prepared->workOutput = reserveTemporaryPath(parentPath(output), settings.format, error);
        if (prepared->workOutput.isEmpty()) {
            cleanupPreparedOutput();
            return false;
        }
    }

    if (prepared->processOutput.isEmpty()) {
        prepared->processOutput = replaceSource ? prepared->workOutput : output;
    }
    return true;
}

void HexImgBackend::cleanupPreparedOutput(bool removeFinalOutput)
{
    if (m_preparedOutput.processInputTemporary && !m_preparedOutput.processInput.isEmpty()) {
        QFile::remove(m_preparedOutput.processInput);
    }
    if (!m_preparedOutput.processOutput.isEmpty() && m_preparedOutput.processOutput != m_preparedOutput.output) {
        QFile::remove(m_preparedOutput.processOutput);
    }
    if (!m_preparedOutput.workOutput.isEmpty() && m_preparedOutput.workOutput != m_preparedOutput.output) {
        QFile::remove(m_preparedOutput.workOutput);
    }
    if (removeFinalOutput && !m_preparedOutput.replaceSource && !m_preparedOutput.output.isEmpty()) {
        QFile::remove(m_preparedOutput.output);
    }
    m_preparedOutput = {};
}

bool HexImgBackend::postProcessOutput(const PreparedOutput &prepared, const ConversionSettings &settings, QString *error)
{
    if (settings.format != QStringLiteral("svg") && settings.format != QStringLiteral("heic")) {
        return true;
    }

    const QString destination = prepared.replaceSource ? prepared.workOutput : prepared.output;
    if (settings.format == QStringLiteral("svg")) {
        return writeSvgOutput(prepared.processOutput, destination, error);
    }
    return writeHeicOutput(prepared.processOutput, destination, settings.quality, error);
}

bool HexImgBackend::finalizeOutput(const QString &input, const PreparedOutput &prepared, QString *error)
{
    if (!prepared.replaceSource) return true;
    if (prepared.workOutput.isEmpty()) {
        *error = QStringLiteral("临时输出路径不可用");
        return false;
    }

    if (samePath(input, prepared.output)) {
        const QString backup = reserveBackupPath(prepared.output, error);
        if (backup.isEmpty()) return false;
        if (!QFile::rename(prepared.output, backup)) {
            *error = QStringLiteral("备份源文件失败");
            return false;
        }
        if (!QFile::rename(prepared.workOutput, prepared.output)) {
            QFile::rename(backup, prepared.output);
            *error = QStringLiteral("替换源文件失败，已恢复原文件");
            return false;
        }
        QFile::remove(backup);
        return true;
    }

    if (QFile::exists(prepared.output) || !QFile::rename(prepared.workOutput, prepared.output)) {
        QFile::remove(prepared.workOutput);
        *error = QStringLiteral("保存替换文件失败，目标文件可能已存在");
        return false;
    }
    if (!QFile::remove(input)) {
        *error = QStringLiteral("删除源文件失败，输出文件已保存");
    }
    return true;
}

bool HexImgBackend::renderSvgInput(const QString &input, const QString &output, QString *error) const
{
    QSvgRenderer renderer(input);
    if (!renderer.isValid()) {
        *error = QStringLiteral("SVG 文件无效或无法解析");
        return false;
    }

    QSize size = renderer.defaultSize();
    if (!size.isValid() || size.isEmpty()) size = QSize(1024, 1024);
    size.scale(QSize(4096, 4096), Qt::KeepAspectRatio);

    QImage image(size, QImage::Format_ARGB32_Premultiplied);
    image.fill(Qt::transparent);
    QPainter painter(&image);
    renderer.render(&painter);
    painter.end();

    if (!image.save(output, "PNG")) {
        *error = QStringLiteral("渲染 SVG 临时图片失败");
        return false;
    }
    return true;
}

bool HexImgBackend::writeSvgOutput(const QString &input, const QString &output, QString *error) const
{
    QImage image(input);
    QFile source(input);
    if (image.isNull() || !source.open(QIODevice::ReadOnly)) {
        *error = QStringLiteral("读取 SVG 中间图片失败");
        return false;
    }

    const QByteArray encoded = source.readAll().toBase64();
    const QByteArray document = QStringLiteral(
        "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
        "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%1\" height=\"%2\" viewBox=\"0 0 %1 %2\">\n"
        "  <image width=\"%1\" height=\"%2\" href=\"data:image/png;base64,%3\"/>\n"
        "</svg>\n")
        .arg(image.width())
        .arg(image.height())
        .arg(QString::fromLatin1(encoded))
        .toUtf8();

    QSaveFile destination(output);
    if (!destination.open(QIODevice::WriteOnly) || destination.write(document) != document.size() || !destination.commit()) {
        *error = QStringLiteral("写入 SVG 文件失败");
        return false;
    }
    return true;
}

bool HexImgBackend::writeHeicOutput(const QString &input, const QString &output, int quality, QString *error) const
{
    const QString helper = QDir(QCoreApplication::applicationDirPath()).filePath(QStringLiteral("tools/heif/heif-enc.exe"));
    if (QFileInfo::exists(helper)) {
        QFile::remove(output);
        QProcess encoder;
        encoder.setProcessChannelMode(QProcess::MergedChannels);
        encoder.start(
            helper,
            {QStringLiteral("--hevc"), QStringLiteral("--quality"), QString::number(clampQuality(quality)),
             QStringLiteral("--output"), output, input});
        if (!encoder.waitForStarted(5000)) {
            *error = QStringLiteral("启动内置 HEIC 编码器失败：%1").arg(encoder.errorString());
            return false;
        }
        if (!encoder.waitForFinished(300000)) {
            encoder.kill();
            encoder.waitForFinished(3000);
            *error = QStringLiteral("HEIC 编码超时");
            QFile::remove(output);
            return false;
        }
        if (encoder.exitStatus() != QProcess::NormalExit || encoder.exitCode() != 0 || QFileInfo(output).size() <= 0) {
            const QString details = QString::fromLocal8Bit(encoder.readAll()).trimmed();
            *error = QStringLiteral("内置 HEIC 编码器失败：%1").arg(details.isEmpty() ? QString::number(encoder.exitCode()) : details);
            QFile::remove(output);
            return false;
        }
        return true;
    }

#ifdef Q_OS_WIN
    QImage image(input);
    if (image.isNull()) {
        *error = QStringLiteral("读取 HEIC 中间图片失败");
        return false;
    }
    image = image.convertToFormat(QImage::Format_BGR888);

    const HRESULT initializeResult = CoInitializeEx(nullptr, COINIT_MULTITHREADED);
    const bool shouldUninitialize = initializeResult == S_OK || initializeResult == S_FALSE;
    if (FAILED(initializeResult) && initializeResult != RPC_E_CHANGED_MODE) {
        *error = QStringLiteral("初始化 Windows 图像编码服务失败");
        return false;
    }

    IWICImagingFactory *factory = nullptr;
    IWICStream *stream = nullptr;
    IWICBitmapEncoder *encoder = nullptr;
    IWICBitmapFrameEncode *frame = nullptr;
    IPropertyBag2 *properties = nullptr;
    HRESULT result = S_OK;
    QString step;

    QFile::remove(output);
    do {
        result = CoCreateInstance(CLSID_WICImagingFactory, nullptr, CLSCTX_INPROC_SERVER, IID_IWICImagingFactory, reinterpret_cast<void **>(&factory));
        if (FAILED(result)) { step = QStringLiteral("创建 Windows 图像工厂失败"); break; }
        result = factory->CreateStream(&stream);
        if (FAILED(result)) { step = QStringLiteral("创建 HEIC 输出流失败"); break; }
        const std::wstring outputPath = QDir::toNativeSeparators(output).toStdWString();
        result = stream->InitializeFromFilename(outputPath.c_str(), GENERIC_WRITE);
        if (FAILED(result)) { step = QStringLiteral("打开 HEIC 输出文件失败"); break; }
        result = factory->CreateEncoder(GUID_ContainerFormatHeif, nullptr, &encoder);
        if (FAILED(result)) { step = QStringLiteral("Windows HEIF 图像扩展不可用"); break; }
        result = encoder->Initialize(stream, WICBitmapEncoderNoCache);
        if (FAILED(result)) { step = QStringLiteral("初始化 HEIC 编码器失败"); break; }
        result = encoder->CreateNewFrame(&frame, &properties);
        if (FAILED(result)) { step = QStringLiteral("创建 HEIC 图像帧失败"); break; }

        if (properties) {
            PROPBAG2 option = {};
            option.pstrName = const_cast<wchar_t *>(L"ImageQuality");
            VARIANT value;
            VariantInit(&value);
            value.vt = VT_R4;
            value.fltVal = static_cast<float>(clampQuality(quality)) / 100.0f;
            properties->Write(1, &option, &value);
            VariantClear(&value);
        }

        result = frame->Initialize(properties);
        if (FAILED(result)) { step = QStringLiteral("初始化 HEIC 图像帧失败"); break; }
        result = frame->SetSize(static_cast<UINT>(image.width()), static_cast<UINT>(image.height()));
        if (FAILED(result)) { step = QStringLiteral("设置 HEIC 图像尺寸失败"); break; }
        WICPixelFormatGUID pixelFormat = GUID_WICPixelFormat24bppBGR;
        result = frame->SetPixelFormat(&pixelFormat);
        if (FAILED(result) || pixelFormat != GUID_WICPixelFormat24bppBGR) {
            step = QStringLiteral("HEIC 编码器不支持当前像素格式");
            break;
        }
        result = frame->WritePixels(
            static_cast<UINT>(image.height()),
            static_cast<UINT>(image.bytesPerLine()),
            static_cast<UINT>(image.sizeInBytes()),
            image.bits());
        if (FAILED(result)) { step = QStringLiteral("写入 HEIC 像素失败"); break; }
        result = frame->Commit();
        if (FAILED(result)) { step = QStringLiteral("提交 HEIC 图像帧失败"); break; }
        result = encoder->Commit();
        if (FAILED(result)) { step = QStringLiteral("提交 HEIC 文件失败"); break; }
    } while (false);

    if (properties) properties->Release();
    if (frame) frame->Release();
    if (encoder) encoder->Release();
    if (stream) stream->Release();
    if (factory) factory->Release();
    if (shouldUninitialize) CoUninitialize();

    if (FAILED(result)) {
        QFile::remove(output);
        *error = QStringLiteral("%1（错误 0x%2）").arg(step).arg(static_cast<quint32>(result), 8, 16, QLatin1Char('0'));
        return false;
    }
    return true;
#else
    Q_UNUSED(input)
    Q_UNUSED(output)
    Q_UNUSED(quality)
    *error = QStringLiteral("当前平台不支持 HEIC 编码");
    return false;
#endif
}

QString HexImgBackend::reserveTemporaryPath(const QString &directory, const QString &extension, QString *error) const
{
    QTemporaryFile temp(QDir(directory).filePath(QStringLiteral(".heximg-XXXXXX.%1").arg(extension)));
    temp.setAutoRemove(false);
    if (!temp.open()) {
        *error = QStringLiteral("创建临时文件失败");
        return {};
    }
    const QString path = temp.fileName();
    temp.close();
    QFile::remove(path);
    return path;
}

QString HexImgBackend::outputFor(const QString &input, const ConversionSettings &settings, bool *replaceSource) const
{
    if (input.isEmpty()) return {};
    const QFileInfo info(input);
    const QString base = info.completeBaseName();
    const QString extension = settings.format;
    *replaceSource = settings.outputMode == 2;
    if (settings.outputMode == 1) {
        const QString folder = QDir(info.absolutePath()).filePath(sanitizePathPart(settings.folderName, QStringLiteral("HexImg")));
        return QDir(folder).filePath(base + QLatin1Char('.') + extension);
    }
    if (settings.outputMode == 2) {
        return QDir(info.absolutePath()).filePath(base + QLatin1Char('.') + extension);
    }
    const QString suffix = settings.suffix.trimmed().isEmpty() ? QStringLiteral("_HexImg") : settings.suffix.trimmed();
    return QDir(info.absolutePath()).filePath(base + suffix + QLatin1Char('.') + extension);
}

QStringList HexImgBackend::ffmpegArguments(const QString &input, const PreparedOutput &prepared, const ConversionSettings &settings) const
{
    QStringList args = {QStringLiteral("-hide_banner"), QStringLiteral("-y"), QStringLiteral("-i"), input, QStringLiteral("-frames:v"), QStringLiteral("1")};
    if (settings.format == QStringLiteral("jpg")) args << QStringLiteral("-q:v") << QString::number(jpegQScale(settings.quality));
    if (settings.format == QStringLiteral("webp")) args << QStringLiteral("-compression_level") << QStringLiteral("6") << QStringLiteral("-quality") << QString::number(clampQuality(settings.quality));
    if (settings.format == QStringLiteral("png")) args << QStringLiteral("-compression_level") << QString::number(qBound(0, settings.pngCompression, 9));
    if (settings.format == QStringLiteral("avif")) {
        const int crf = (100 - clampQuality(settings.quality)) * 63 / 100;
        args << QStringLiteral("-c:v") << QStringLiteral("libaom-av1")
             << QStringLiteral("-still-picture") << QStringLiteral("1")
             << QStringLiteral("-crf") << QString::number(crf)
             << QStringLiteral("-cpu-used") << QStringLiteral("6")
             << QStringLiteral("-pix_fmt") << QStringLiteral("yuv420p");
    }
    if (settings.format == QStringLiteral("ico")) {
        args << QStringLiteral("-vf") << QStringLiteral("scale=256:256:force_original_aspect_ratio=decrease,pad=256:256:(ow-iw)/2:(oh-ih)/2:color=0x00000000")
             << QStringLiteral("-c:v") << QStringLiteral("png")
             << QStringLiteral("-pix_fmt") << QStringLiteral("rgba")
             << QStringLiteral("-f") << QStringLiteral("ico");
    }
    if (settings.format == QStringLiteral("svg") || settings.format == QStringLiteral("heic")) {
        args << QStringLiteral("-compression_level") << QStringLiteral("6");
    }
    args << prepared.processOutput;
    return args;
}

QString HexImgBackend::normalizePath(const QString &path) const
{
    const QFileInfo info(path);
    return info.exists() ? info.absoluteFilePath() : path;
}

QString HexImgBackend::fileName(const QString &path) const { return QFileInfo(path).fileName(); }
QString HexImgBackend::parentPath(const QString &path) const { return QFileInfo(path).absolutePath(); }

QString HexImgBackend::sanitizePathPart(const QString &value, const QString &fallback) const
{
    QString result = value.trimmed();
    if (result.isEmpty()) return fallback;
    for (const QChar &invalid : QStringLiteral("\\/:*?\"<>|")) result.replace(invalid, QLatin1Char('_'));
    return result;
}

bool HexImgBackend::isImagePath(const QString &path) const { return imageExtensions.contains(QFileInfo(path).suffix().toLower()); }

bool HexImgBackend::samePath(const QString &left, const QString &right) const
{
    return QFileInfo(left).absoluteFilePath().compare(QFileInfo(right).absoluteFilePath(), Qt::CaseInsensitive) == 0;
}

QString HexImgBackend::reserveBackupPath(const QString &target, QString *error) const
{
    QTemporaryFile temp(QDir(parentPath(target)).filePath(QStringLiteral(".heximg-backup-XXXXXX.bak")));
    temp.setAutoRemove(false);
    if (!temp.open()) {
        *error = QStringLiteral("创建源文件备份路径失败");
        return {};
    }
    const QString path = temp.fileName();
    temp.close();
    QFile::remove(path);
    return path;
}

int HexImgBackend::clampQuality(int value) { return qBound(0, value, 100); }
int HexImgBackend::jpegQScale(int value) { return 31 - static_cast<int>(clampQuality(value) * 29.0 / 100.0); }
