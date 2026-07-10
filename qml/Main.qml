import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

ApplicationWindow {
    id: window

    width: 1080
    height: 700
    minimumWidth: 860
    minimumHeight: 560
    visible: true
    title: "HexImg"
    color: backgroundColor

    property color backgroundColor: backend.darkTheme ? "#09090b" : "#fafafa"
    property color foregroundColor: backend.darkTheme ? "#fafafa" : "#18181b"
    property color mutedForeground: backend.darkTheme ? "#a1a1aa" : "#71717a"
    property color cardColor: backend.darkTheme ? "#0f0f12" : "#ffffff"
    property color mutedColor: backend.darkTheme ? "#18181b" : "#f4f4f5"
    property color borderColor: backend.darkTheme ? "#27272a" : "#e4e4e7"
    property color strongBorderColor: backend.darkTheme ? "#3f3f46" : "#d4d4d8"
    property var formats: ["JPG", "PNG", "WebP", "AVIF", "HEIC", "GIF", "ICO", "SVG", "BMP", "TIFF"]

    function formatIndex(value) {
        var normalized = value.toLowerCase()
        var values = ["jpg", "png", "webp", "avif", "heic", "gif", "ico", "svg", "bmp", "tiff"]
        return Math.max(0, values.indexOf(normalized))
    }

    function fileExtension(value) {
        var dot = value.lastIndexOf(".")
        return dot >= 0 ? value.substring(dot + 1) : "FILE"
    }

    function formatHasQuality(value) {
        return value === "jpg" || value === "webp" || value === "avif" || value === "heic"
    }

    header: Rectangle {
        height: 60
        color: window.cardColor
        border.width: 0

        Rectangle {
            anchors.left: parent.left
            anchors.right: parent.right
            anchors.bottom: parent.bottom
            height: 1
            color: window.borderColor
        }

        RowLayout {
            anchors.fill: parent
            anchors.leftMargin: 20
            anchors.rightMargin: 16
            spacing: 12

            ColumnLayout {
                spacing: 0

                Label {
                    text: "HexImg"
                    color: window.foregroundColor
                    font.pixelSize: 20
                    font.weight: Font.DemiBold
                }
                Label {
                    visible: window.width >= 920
                    text: "Qt 6 · QML · FFmpeg"
                    color: window.mutedForeground
                    font.pixelSize: 11
                }
            }

            Item {
                Layout.fillWidth: true
            }

            RowLayout {
                Layout.alignment: Qt.AlignRight | Qt.AlignVCenter
                spacing: 8

                Rectangle {
                    Layout.preferredHeight: 28
                    Layout.preferredWidth: ffmpegStatus.implicitWidth + 20
                    radius: 7
                    color: backend.ffmpegAvailable
                        ? (backend.darkTheme ? "#052e24" : "#ecfdf5")
                        : (backend.darkTheme ? "#450a0a" : "#fef2f2")
                    border.width: 1
                    border.color: backend.ffmpegAvailable
                        ? (backend.darkTheme ? "#065f46" : "#a7f3d0")
                        : (backend.darkTheme ? "#7f1d1d" : "#fecaca")

                    Label {
                        id: ffmpegStatus
                        anchors.centerIn: parent
                        text: backend.ffmpegAvailable ? "FFmpeg 就绪" : "FFmpeg 不可用"
                        color: backend.ffmpegAvailable
                            ? (backend.darkTheme ? "#6ee7b7" : "#047857")
                            : (backend.darkTheme ? "#fca5a5" : "#b91c1c")
                        font.pixelSize: 11
                        font.weight: Font.Medium
                    }
                }

                ShadButton {
                    dark: backend.darkTheme
                    variant: "ghost"
                    text: backend.darkTheme ? "☼" : "☾"
                    Layout.preferredWidth: 32
                    Layout.preferredHeight: 32
                    onClicked: backend.darkTheme = !backend.darkTheme
                    ToolTip.visible: hovered
                    ToolTip.text: backend.darkTheme ? "切换浅色主题" : "切换深色主题"
                }
            }
        }
    }

    ColumnLayout {
        anchors.fill: parent
        anchors.margins: 16
        spacing: 12

        RowLayout {
            Layout.fillWidth: true
            Layout.fillHeight: true
            spacing: 12

            Rectangle {
                Layout.fillWidth: true
                Layout.fillHeight: true
                Layout.minimumWidth: 480
                color: window.cardColor
                radius: 8
                border.width: 1
                border.color: window.borderColor

                ColumnLayout {
                    anchors.fill: parent
                    anchors.margins: 14
                    spacing: 10

                    RowLayout {
                        Layout.fillWidth: true
                        Layout.preferredHeight: 32
                        spacing: 8

                        Label {
                            text: "图片队列"
                            color: window.foregroundColor
                            font.pixelSize: 16
                            font.weight: Font.DemiBold
                        }
                        Label {
                            text: backend.imageCount + " 个文件"
                            color: window.mutedForeground
                            font.pixelSize: 11
                        }
                        Item { Layout.fillWidth: true }
                        ShadButton {
                            dark: backend.darkTheme
                            variant: "outline"
                            text: "清空"
                            enabled: backend.imageCount > 0 && !backend.converting
                            onClicked: backend.clearImages()
                        }
                        ShadButton {
                            dark: backend.darkTheme
                            text: "选择图片"
                            enabled: !backend.converting
                            onClicked: backend.chooseImages()
                        }
                    }

                    Rectangle {
                        id: queueSurface
                        Layout.fillWidth: true
                        Layout.fillHeight: true
                        color: window.backgroundColor
                        radius: 7
                        border.width: 1
                        border.color: window.borderColor

                        Label {
                            anchors.centerIn: parent
                            visible: backend.imageCount === 0
                            text: "暂无图片"
                            color: window.mutedForeground
                            font.pixelSize: 13
                        }

                        ListView {
                            id: imageList
                            anchors.fill: parent
                            anchors.margins: 8
                            visible: backend.imageCount > 0
                            clip: true
                            spacing: 6
                            model: backend.queueModel

                            delegate: Rectangle {
                                required property int index
                                required property string fileName
                                required property string outputPath
                                width: imageList.width
                                height: 58
                                color: removeArea.hovered
                                    ? window.mutedColor
                                    : window.cardColor
                                radius: 7
                                border.width: 1
                                border.color: window.borderColor

                                RowLayout {
                                    anchors.fill: parent
                                    anchors.leftMargin: 10
                                    anchors.rightMargin: 7
                                    spacing: 10

                                    FormatBadge {
                                        format: window.fileExtension(fileName)
                                        Layout.alignment: Qt.AlignVCenter
                                    }

                                    ColumnLayout {
                                        Layout.fillWidth: true
                                        spacing: 2

                                        Label {
                                            Layout.fillWidth: true
                                            text: fileName
                                            color: window.foregroundColor
                                            font.pixelSize: 13
                                            font.weight: Font.Medium
                                            elide: Text.ElideMiddle
                                        }
                                        Label {
                                            Layout.fillWidth: true
                                            text: outputPath
                                            color: window.mutedForeground
                                            font.pixelSize: 10
                                            elide: Text.ElideMiddle
                                        }
                                    }

                                    ShadButton {
                                        id: removeArea
                                        dark: backend.darkTheme
                                        variant: "ghost"
                                        text: "×"
                                        enabled: !backend.converting
                                        Layout.preferredWidth: 30
                                        Layout.preferredHeight: 30
                                        onClicked: backend.removeImage(index)
                                        ToolTip.visible: hovered
                                        ToolTip.text: "移除"
                                    }
                                }
                            }
                        }

                        DropArea {
                            anchors.fill: parent
                            onDropped: function(drop) {
                                for (var i = 0; i < drop.urls.length; ++i) {
                                    backend.addDroppedPath(drop.urls[i].toString())
                                }
                            }
                        }
                    }

                    RowLayout {
                        Layout.fillWidth: true
                        Layout.preferredHeight: 24
                        spacing: 8

                        Label {
                            text: "输出预览"
                            color: window.mutedForeground
                            font.pixelSize: 11
                        }
                        Label {
                            Layout.fillWidth: true
                            text: backend.previewOutput || "—"
                            color: window.foregroundColor
                            font.pixelSize: 11
                            elide: Text.ElideMiddle
                        }
                    }
                }
            }

            Rectangle {
                Layout.fillHeight: true
                Layout.preferredWidth: 320
                Layout.minimumWidth: 300
                color: window.cardColor
                radius: 8
                border.width: 1
                border.color: window.borderColor

                ColumnLayout {
                    anchors.fill: parent
                    anchors.margins: 14
                    spacing: 9

                    Label {
                        text: "转换设置"
                        color: window.foregroundColor
                        font.pixelSize: 16
                        font.weight: Font.DemiBold
                        Layout.preferredHeight: 24
                    }

                    Rectangle {
                        Layout.fillWidth: true
                        Layout.preferredHeight: 1
                        color: window.borderColor
                    }

                    ScrollView {
                        id: settingsScroll
                        Layout.fillWidth: true
                        Layout.preferredHeight: Math.min(settingsColumn.implicitHeight,
                                                         Math.max(140, parent.height - 112))
                        clip: true
                        contentWidth: availableWidth
                        contentHeight: settingsColumn.implicitHeight
                        ScrollBar.horizontal.policy: ScrollBar.AlwaysOff

                        Column {
                            id: settingsColumn
                            width: settingsScroll.availableWidth
                            spacing: 8

                            Label {
                                text: "目标格式"
                                color: window.foregroundColor
                                font.pixelSize: 12
                                font.weight: Font.Medium
                            }
                            ShadComboBox {
                                width: parent.width
                                dark: backend.darkTheme
                                model: window.formats
                                currentIndex: window.formatIndex(backend.format)
                                onActivated: backend.format = currentText.toLowerCase()
                            }

                            Item { width: 1; height: 3 }

                            Column {
                                width: parent.width
                                visible: backend.format === "png" || window.formatHasQuality(backend.format)
                                spacing: 6

                                RowLayout {
                                    width: parent.width
                                    spacing: 8

                                    Label {
                                        text: backend.format === "png" ? "PNG 压缩级别" : "输出质量"
                                        color: window.foregroundColor
                                        font.pixelSize: 12
                                        font.weight: Font.Medium
                                    }
                                    Item { Layout.fillWidth: true }
                                    Label {
                                        id: qualityValue
                                        text: backend.format === "png" ? backend.pngCompression : backend.quality
                                        color: window.mutedForeground
                                        font.pixelSize: 12
                                        font.weight: Font.Medium
                                    }
                                }

                                Slider {
                                    id: qualitySlider
                                    width: parent.width
                                    height: 26
                                    from: 0
                                    to: backend.format === "png" ? 9 : 100
                                    stepSize: 1
                                    value: backend.format === "png" ? backend.pngCompression : backend.quality
                                    onMoved: backend.format === "png"
                                        ? backend.pngCompression = value
                                        : backend.quality = value

                                    background: Rectangle {
                                        x: qualitySlider.leftPadding
                                        y: qualitySlider.topPadding + qualitySlider.availableHeight / 2 - 2
                                        width: qualitySlider.availableWidth
                                        height: 4
                                        radius: 2
                                        color: window.mutedColor

                                        Rectangle {
                                            width: qualitySlider.visualPosition * parent.width
                                            height: parent.height
                                            radius: 2
                                            color: window.foregroundColor
                                        }
                                    }
                                    handle: Rectangle {
                                        x: qualitySlider.leftPadding + qualitySlider.visualPosition * (qualitySlider.availableWidth - width)
                                        y: qualitySlider.topPadding + qualitySlider.availableHeight / 2 - height / 2
                                        width: 14
                                        height: 14
                                        radius: 7
                                        color: window.backgroundColor
                                        border.width: 2
                                        border.color: window.foregroundColor
                                    }
                                }
                            }

                            Label {
                                text: "输出方式"
                                color: window.foregroundColor
                                font.pixelSize: 12
                                font.weight: Font.Medium
                            }
                            ShadComboBox {
                                width: parent.width
                                dark: backend.darkTheme
                                model: ["添加后缀", "当前目录文件夹", "替换源文件"]
                                currentIndex: backend.outputMode
                                onActivated: backend.outputMode = currentIndex
                            }

                            Label {
                                visible: backend.outputMode === 0
                                text: "文件后缀"
                                color: window.foregroundColor
                                font.pixelSize: 12
                                font.weight: Font.Medium
                            }
                            ShadTextField {
                                visible: backend.outputMode === 0
                                width: parent.width
                                dark: backend.darkTheme
                                text: backend.suffix
                                placeholderText: "_HexImg"
                                onTextChanged: if (text !== backend.suffix) backend.suffix = text
                            }

                            Label {
                                visible: backend.outputMode === 1
                                text: "输出文件夹"
                                color: window.foregroundColor
                                font.pixelSize: 12
                                font.weight: Font.Medium
                            }
                            ShadTextField {
                                visible: backend.outputMode === 1
                                width: parent.width
                                dark: backend.darkTheme
                                text: backend.folderName
                                placeholderText: "HexImg"
                                onTextChanged: if (text !== backend.folderName) backend.folderName = text
                            }
                        }
                    }

                    Rectangle {
                        Layout.fillWidth: true
                        Layout.preferredHeight: 1
                        color: window.borderColor
                    }

                    Label {
                        Layout.fillWidth: true
                        Layout.preferredHeight: 18
                        text: backend.status
                        color: window.mutedForeground
                        font.pixelSize: 11
                        elide: Text.ElideRight
                        verticalAlignment: Text.AlignVCenter
                    }

                    RowLayout {
                        Layout.fillWidth: true
                        Layout.preferredHeight: 32
                        spacing: 8

                        ShadButton {
                            Layout.fillWidth: true
                            dark: backend.darkTheme
                            variant: "outline"
                            text: "停止"
                            enabled: backend.converting
                            onClicked: backend.cancelConversion()
                        }
                        ShadButton {
                            Layout.fillWidth: true
                            dark: backend.darkTheme
                            text: backend.converting ? "转换中" : "开始转换"
                            enabled: backend.imageCount > 0 && backend.ffmpegAvailable && !backend.converting
                            onClicked: backend.startConversion()
                        }
                    }
                }
            }
        }

        Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: 112
            color: window.cardColor
            radius: 8
            border.width: 1
            border.color: window.borderColor

            ColumnLayout {
                anchors.fill: parent
                anchors.margins: 11
                spacing: 5

                RowLayout {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 20
                    Label {
                        text: "运行日志"
                        color: window.foregroundColor
                        font.pixelSize: 12
                        font.weight: Font.DemiBold
                    }
                    Item { Layout.fillWidth: true }
                    Label {
                        text: backend.logs.length + " 条"
                        color: window.mutedForeground
                        font.pixelSize: 10
                    }
                }

                ScrollView {
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    clip: true

                    TextArea {
                        readOnly: true
                        text: backend.logs.join("\n")
                        color: window.mutedForeground
                        selectionColor: window.strongBorderColor
                        selectedTextColor: window.foregroundColor
                        font.family: "Consolas"
                        font.pixelSize: 10
                        wrapMode: TextArea.NoWrap
                        leftPadding: 0
                        rightPadding: 0
                        topPadding: 0
                        bottomPadding: 0
                        background: Rectangle { color: "transparent" }
                    }
                }
            }
        }
    }
}
