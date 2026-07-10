#pragma once

#include "queuemodel.h"

#include <QProcess>
#include <QTemporaryFile>
#include <QUrl>
#include <QVector>

class HexImgBackend final : public QObject
{
    Q_OBJECT
    Q_PROPERTY(QueueModel *queueModel READ queueModel CONSTANT)
    Q_PROPERTY(int imageCount READ imageCount NOTIFY queueChanged)
    Q_PROPERTY(bool ffmpegAvailable READ ffmpegAvailable CONSTANT)
    Q_PROPERTY(bool converting READ converting NOTIFY convertingChanged)
    Q_PROPERTY(bool darkTheme READ darkTheme WRITE setDarkTheme NOTIFY darkThemeChanged)
    Q_PROPERTY(QString format READ format WRITE setFormat NOTIFY settingsChanged)
    Q_PROPERTY(int quality READ quality WRITE setQuality NOTIFY settingsChanged)
    Q_PROPERTY(int pngCompression READ pngCompression WRITE setPngCompression NOTIFY settingsChanged)
    Q_PROPERTY(int outputMode READ outputMode WRITE setOutputMode NOTIFY settingsChanged)
    Q_PROPERTY(QString suffix READ suffix WRITE setSuffix NOTIFY settingsChanged)
    Q_PROPERTY(QString folderName READ folderName WRITE setFolderName NOTIFY settingsChanged)
    Q_PROPERTY(QString previewOutput READ previewOutput NOTIFY settingsChanged)
    Q_PROPERTY(QString status READ status NOTIFY statusChanged)
    Q_PROPERTY(QStringList logs READ logs NOTIFY logsChanged)

public:
    explicit HexImgBackend(QObject *parent = nullptr);
    ~HexImgBackend() override;

    QueueModel *queueModel();
    int imageCount() const;
    bool ffmpegAvailable() const;
    bool converting() const;
    bool darkTheme() const;
    void setDarkTheme(bool value);
    QString format() const;
    void setFormat(const QString &value);
    int quality() const;
    void setQuality(int value);
    int pngCompression() const;
    void setPngCompression(int value);
    int outputMode() const;
    void setOutputMode(int value);
    QString suffix() const;
    void setSuffix(const QString &value);
    QString folderName() const;
    void setFolderName(const QString &value);
    QString previewOutput() const;
    QString status() const;
    QStringList logs() const;

    Q_INVOKABLE void chooseImages();
    Q_INVOKABLE void addDroppedPath(const QString &urlOrPath);
    Q_INVOKABLE void removeImage(int index);
    Q_INVOKABLE void clearImages();
    Q_INVOKABLE void startConversion();
    Q_INVOKABLE void cancelConversion();

signals:
    void convertingChanged();
    void darkThemeChanged();
    void settingsChanged();
    void statusChanged();
    void logsChanged();
    void queueChanged();

private slots:
    void readProcessOutput();
    void processError(QProcess::ProcessError error);
    void processFinished(int exitCode, QProcess::ExitStatus exitStatus);

private:
    struct ConversionSettings {
        QString format;
        int quality = 85;
        int pngCompression = 6;
        int outputMode = 0;
        QString suffix;
        QString folderName;
    };

    struct PreparedOutput {
        QString output;
        QString processInput;
        QString processOutput;
        QString workOutput;
        bool processInputTemporary = false;
        bool replaceSource = false;
    };

    void rebuildModel();
    void addPath(const QString &path);
    void setStatus(const QString &value);
    void appendLog(const QString &line);
    void beginNextConversion();
    void cleanupPreparedOutput(bool removeFinalOutput = false);
    bool prepareOutput(const QString &input, const ConversionSettings &settings, PreparedOutput *prepared, QString *error);
    bool postProcessOutput(const PreparedOutput &prepared, const ConversionSettings &settings, QString *error);
    bool finalizeOutput(const QString &input, const PreparedOutput &prepared, QString *error);
    bool renderSvgInput(const QString &input, const QString &output, QString *error) const;
    bool writeSvgOutput(const QString &input, const QString &output, QString *error) const;
    bool writeHeicOutput(const QString &input, const QString &output, int quality, QString *error) const;
    QString reserveTemporaryPath(const QString &directory, const QString &extension, QString *error) const;
    QString outputFor(const QString &input, const ConversionSettings &settings, bool *replaceSource) const;
    QStringList ffmpegArguments(const QString &input, const PreparedOutput &prepared, const ConversionSettings &settings) const;
    QString normalizePath(const QString &path) const;
    QString fileName(const QString &path) const;
    QString parentPath(const QString &path) const;
    QString sanitizePathPart(const QString &value, const QString &fallback) const;
    bool isImagePath(const QString &path) const;
    bool samePath(const QString &left, const QString &right) const;
    QString reserveBackupPath(const QString &target, QString *error) const;
    static int clampQuality(int value);
    static int jpegQScale(int value);

    QueueModel m_queueModel;
    QString m_format = QStringLiteral("jpg");
    int m_quality = 85;
    int m_pngCompression = 6;
    int m_outputMode = 0;
    QString m_suffix = QStringLiteral("_HexImg");
    QString m_folderName = QStringLiteral("HexImg");
    bool m_ffmpegAvailable = false;
    bool m_converting = false;
    bool m_darkTheme = true;
    QString m_status = QStringLiteral("就绪");
    QStringList m_logs;

    QProcess *m_process = nullptr;
    QVector<ImageItem> m_conversionItems;
    ConversionSettings m_conversionSettings;
    int m_conversionIndex = -1;
    int m_failedCount = 0;
    PreparedOutput m_preparedOutput;
    bool m_cancelRequested = false;
};
