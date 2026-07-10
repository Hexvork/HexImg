#include "heximgbackend.h"

#include <QFileInfo>
#include <QImage>
#include <QTemporaryDir>
#include <QtTest>

class HexImgBackendTest final : public QObject
{
    Q_OBJECT

private slots:
    void buildsSanitizedOutputPreview();
    void convertsPngToWebp();
    void convertsAdditionalFormats_data();
    void convertsAdditionalFormats();
    void convertsSvgInput();
};

void HexImgBackendTest::buildsSanitizedOutputPreview()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());

    const QString input = dir.filePath(QStringLiteral("photo.png"));
    QImage image(4, 4, QImage::Format_ARGB32);
    image.fill(QColor(QStringLiteral("#2f81f7")));
    QVERIFY(image.save(input));

    HexImgBackend backend;
    backend.addDroppedPath(input);
    backend.setFormat(QStringLiteral("webp"));
    backend.setOutputMode(1);
    backend.setFolderName(QStringLiteral("Hex:Img"));

    QCOMPARE(backend.imageCount(), 1);
    QCOMPARE(backend.previewOutput(), dir.filePath(QStringLiteral("Hex_Img/photo.webp")));
}

void HexImgBackendTest::convertsPngToWebp()
{
    HexImgBackend backend;
    if (!backend.ffmpegAvailable()) {
        QSKIP("FFmpeg is not available on PATH");
    }

    QTemporaryDir dir;
    QVERIFY(dir.isValid());

    const QString input = dir.filePath(QStringLiteral("source.png"));
    const QString output = dir.filePath(QStringLiteral("source_test.webp"));
    QImage image(8, 8, QImage::Format_ARGB32);
    image.fill(QColor(QStringLiteral("#39d98a")));
    QVERIFY(image.save(input));

    backend.addDroppedPath(input);
    backend.setFormat(QStringLiteral("webp"));
    backend.setSuffix(QStringLiteral("_test"));
    backend.startConversion();

    QVERIFY(backend.converting());
    QTRY_VERIFY_WITH_TIMEOUT(!backend.converting(), 30000);
    QCOMPARE(backend.status(), QStringLiteral("转换完成"));
    QVERIFY2(QFileInfo::exists(output), qPrintable(output));
    QVERIFY(QFileInfo(output).size() > 0);
}

void HexImgBackendTest::convertsAdditionalFormats_data()
{
    QTest::addColumn<QString>("format");
    QTest::newRow("avif") << QStringLiteral("avif");
    QTest::newRow("ico") << QStringLiteral("ico");
    QTest::newRow("gif") << QStringLiteral("gif");
    QTest::newRow("svg") << QStringLiteral("svg");
    QTest::newRow("heic") << QStringLiteral("heic");
}

void HexImgBackendTest::convertsAdditionalFormats()
{
    QFETCH(QString, format);
    HexImgBackend backend;
    if (!backend.ffmpegAvailable()) QSKIP("FFmpeg is not available on PATH");

    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString input = dir.filePath(QStringLiteral("source.png"));
    const QString output = dir.filePath(QStringLiteral("source_test.%1").arg(format));
    QImage image(32, 24, QImage::Format_ARGB32);
    image.fill(QColor(QStringLiteral("#8b5cf6")));
    QVERIFY(image.save(input));

    backend.addDroppedPath(input);
    backend.setFormat(format);
    backend.setSuffix(QStringLiteral("_test"));
    backend.startConversion();

    QTRY_VERIFY_WITH_TIMEOUT(!backend.converting(), 60000);
    QVERIFY2(backend.status() == QStringLiteral("转换完成"), qPrintable(backend.logs().join(QLatin1Char('\n'))));
    QVERIFY2(QFileInfo::exists(output), qPrintable(output));
    QVERIFY(QFileInfo(output).size() > 0);
    if (format == QStringLiteral("svg")) {
        QFile svg(output);
        QVERIFY(svg.open(QIODevice::ReadOnly));
        QVERIFY(svg.readAll().contains("<svg"));
    }
}

void HexImgBackendTest::convertsSvgInput()
{
    HexImgBackend backend;
    if (!backend.ffmpegAvailable()) QSKIP("FFmpeg is not available on PATH");

    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString input = dir.filePath(QStringLiteral("vector.svg"));
    const QString output = dir.filePath(QStringLiteral("vector_test.png"));
    QFile svg(input);
    QVERIFY(svg.open(QIODevice::WriteOnly));
    svg.write("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"48\" height=\"32\"><rect width=\"48\" height=\"32\" fill=\"#22c55e\"/></svg>");
    svg.close();

    backend.addDroppedPath(input);
    backend.setFormat(QStringLiteral("png"));
    backend.setSuffix(QStringLiteral("_test"));
    backend.startConversion();

    QTRY_VERIFY_WITH_TIMEOUT(!backend.converting(), 30000);
    QCOMPARE(backend.status(), QStringLiteral("转换完成"));
    QVERIFY(QFileInfo::exists(output));
    QVERIFY(QFileInfo(output).size() > 0);
}

QTEST_MAIN(HexImgBackendTest)

#include "heximgbackend_test.moc"
