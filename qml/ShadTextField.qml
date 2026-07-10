import QtQuick
import QtQuick.Controls

TextField {
    id: control

    property bool dark: true

    implicitHeight: 34
    leftPadding: 11
    rightPadding: 11
    color: dark ? "#fafafa" : "#18181b"
    placeholderTextColor: dark ? "#71717a" : "#a1a1aa"
    selectionColor: dark ? "#3f3f46" : "#d4d4d8"
    selectedTextColor: color
    font.pixelSize: 13

    background: Rectangle {
        radius: 7
        color: control.dark ? "#18181b" : "#ffffff"
        border.width: control.activeFocus ? 2 : 1
        border.color: control.activeFocus
            ? (control.dark ? "#a1a1aa" : "#71717a")
            : (control.dark ? "#3f3f46" : "#e4e4e7")
    }
}
