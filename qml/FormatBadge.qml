import QtQuick
import QtQuick.Controls

Rectangle {
    id: badge

    property string format: "FILE"
    readonly property string label: normalizedFormat(format)

    implicitWidth: Math.max(38, badgeLabel.implicitWidth + 14)
    implicitHeight: 26
    radius: 6
    color: formatColor(label)

    function normalizedFormat(value) {
        var upper = value.toUpperCase()
        if (upper === "JPEG") return "JPG"
        if (upper === "TIF") return "TIFF"
        if (upper === "HEIF") return "HEIC"
        return upper || "FILE"
    }

    function formatColor(value) {
        var colors = {
            "JPG": "#d97706",
            "PNG": "#0284c7",
            "WEBP": "#059669",
            "AVIF": "#7c3aed",
            "HEIC": "#0891b2",
            "GIF": "#db2777",
            "ICO": "#dc2626",
            "SVG": "#ea580c",
            "BMP": "#2563eb",
            "TIFF": "#475569"
        }
        return colors[value] || "#52525b"
    }

    Label {
        id: badgeLabel
        anchors.centerIn: parent
        text: badge.label
        color: "#ffffff"
        font.pixelSize: 9
        font.weight: Font.DemiBold
        font.letterSpacing: 0
    }
}
