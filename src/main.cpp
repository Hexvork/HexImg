#include "heximgbackend.h"

#include <QApplication>
#include <QIcon>
#include <QQmlApplicationEngine>
#include <QQmlContext>
#include <QQuickStyle>

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    app.setOrganizationName(QStringLiteral("Hexvork"));
    app.setOrganizationDomain(QStringLiteral("hexvork.com"));
    app.setApplicationName(QStringLiteral("HexImg"));
    app.setWindowIcon(QIcon(QStringLiteral(":/assets/HexImg.png")));

    QQuickStyle::setStyle(QStringLiteral("Basic"));

    HexImgBackend backend;
    QQmlApplicationEngine engine;
    engine.rootContext()->setContextProperty(QStringLiteral("backend"), &backend);
    engine.load(QUrl(QStringLiteral("qrc:/qml/Main.qml")));

    if (engine.rootObjects().isEmpty()) {
        return 1;
    }
    return app.exec();
}
