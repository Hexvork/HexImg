import QtQuick
import QtQuick.Controls

ComboBox {
    id: control

    property bool dark: true

    implicitHeight: 34
    leftPadding: 11
    rightPadding: 34
    font.pixelSize: 13

    contentItem: Text {
        leftPadding: control.leftPadding
        rightPadding: control.rightPadding
        text: control.displayText
        color: control.dark ? "#fafafa" : "#18181b"
        font: control.font
        verticalAlignment: Text.AlignVCenter
        elide: Text.ElideRight
    }

    indicator: Text {
        x: control.width - width - 11
        y: (control.height - height) / 2 - 1
        text: "⌄"
        color: control.dark ? "#a1a1aa" : "#71717a"
        font.pixelSize: 16
    }

    background: Rectangle {
        radius: 7
        color: control.dark ? "#18181b" : "#ffffff"
        border.width: control.visualFocus ? 2 : 1
        border.color: control.visualFocus
            ? (control.dark ? "#a1a1aa" : "#71717a")
            : (control.dark ? "#3f3f46" : "#e4e4e7")
    }

    delegate: ItemDelegate {
        id: option
        required property int index
        width: control.width - 8
        height: 31
        leftPadding: 10
        rightPadding: 10
        text: control.textAt(index)
        highlighted: control.highlightedIndex === index
        contentItem: Text {
            text: option.text
            color: control.dark ? "#fafafa" : "#18181b"
            font.pixelSize: 13
            verticalAlignment: Text.AlignVCenter
        }
        background: Rectangle {
            radius: 5
            color: option.highlighted
                ? (control.dark ? "#27272a" : "#f4f4f5")
                : "transparent"
        }
    }

    popup: Popup {
        y: control.height + 4
        width: control.width
        implicitHeight: Math.min(contentItem.implicitHeight + 8, 260)
        padding: 4
        contentItem: ListView {
            clip: true
            implicitHeight: contentHeight
            model: control.popup.visible ? control.delegateModel : null
            currentIndex: control.highlightedIndex
            ScrollIndicator.vertical: ScrollIndicator {}
        }
        background: Rectangle {
            radius: 7
            color: control.dark ? "#18181b" : "#ffffff"
            border.width: 1
            border.color: control.dark ? "#3f3f46" : "#e4e4e7"
        }
    }
}
