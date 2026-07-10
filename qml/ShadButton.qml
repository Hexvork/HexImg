import QtQuick
import QtQuick.Controls

Button {
    id: control

    property bool dark: true
    property string variant: "default"

    implicitHeight: 32
    implicitWidth: Math.max(64, contentItem.implicitWidth + 24)
    leftPadding: 12
    rightPadding: 12
    spacing: 6
    opacity: enabled ? 1 : 0.5

    function baseColor() {
        if (variant === "default") return dark ? "#fafafa" : "#18181b"
        if (variant === "secondary") return dark ? "#27272a" : "#f4f4f5"
        if (variant === "destructive") return dark ? "#7f1d1d" : "#fee2e2"
        return "transparent"
    }

    function hoverColor() {
        if (variant === "default") return dark ? "#e4e4e7" : "#27272a"
        if (variant === "destructive") return dark ? "#991b1b" : "#fecaca"
        return dark ? "#27272a" : "#f4f4f5"
    }

    function foregroundColor() {
        if (variant === "default") return dark ? "#18181b" : "#fafafa"
        if (variant === "destructive") return dark ? "#fecaca" : "#b91c1c"
        return dark ? "#fafafa" : "#18181b"
    }

    contentItem: Text {
        text: control.text
        color: control.foregroundColor()
        font.pixelSize: 13
        font.weight: Font.Medium
        horizontalAlignment: Text.AlignHCenter
        verticalAlignment: Text.AlignVCenter
        elide: Text.ElideRight
    }

    background: Rectangle {
        radius: 7
        color: control.down || control.hovered ? control.hoverColor() : control.baseColor()
        border.width: control.visualFocus ? 2 : (control.variant === "outline" ? 1 : 0)
        border.color: control.visualFocus
            ? (control.dark ? "#a1a1aa" : "#71717a")
            : (control.dark ? "#3f3f46" : "#e4e4e7")
    }
}
