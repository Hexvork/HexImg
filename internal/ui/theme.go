package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type FluentTheme struct {
	dark bool
}

func NewFluentTheme(dark bool) fyne.Theme {
	return &FluentTheme{dark: dark}
}

func (t *FluentTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.dark {
		return darkColor(name)
	}
	return lightColor(name)
}

func (t *FluentTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *FluentTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *FluentTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 6
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func darkColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return BackgroundColor(true)
	case theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		return CardColor(true)
	case theme.ColorNameButton:
		return ButtonColor(true)
	case theme.ColorNameInputBackground:
		return InputColor(true)
	case theme.ColorNameForeground:
		return TextColor(true)
	case theme.ColorNamePrimary, theme.ColorNameFocus:
		return AccentColor(true)
	case theme.ColorNameHover, theme.ColorNamePressed:
		return HoverColor(true)
	case theme.ColorNameSelection:
		return SelectionColor(true)
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return MutedTextColor(true)
	case theme.ColorNameSeparator, theme.ColorNameInputBorder, theme.ColorNameShadow:
		return SeparatorColor(true)
	case theme.ColorNameScrollBar:
		return scrollBarColor(true)
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func lightColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return BackgroundColor(false)
	case theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		return CardColor(false)
	case theme.ColorNameButton:
		return ButtonColor(false)
	case theme.ColorNameInputBackground:
		return InputColor(false)
	case theme.ColorNameForeground:
		return TextColor(false)
	case theme.ColorNamePrimary, theme.ColorNameFocus:
		return AccentColor(false)
	case theme.ColorNameHover, theme.ColorNamePressed:
		return HoverColor(false)
	case theme.ColorNameSelection:
		return SelectionColor(false)
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return MutedTextColor(false)
	case theme.ColorNameSeparator, theme.ColorNameInputBorder, theme.ColorNameShadow:
		return SeparatorColor(false)
	case theme.ColorNameScrollBar:
		return scrollBarColor(false)
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func hexColor(r, g, b uint8) color.Color {
	return color.NRGBA{R: r, G: g, B: b, A: 0xFF}
}

func BackgroundColor(dark bool) color.Color {
	if dark {
		return hexColor(0x0D, 0x11, 0x17)
	}
	return hexColor(0xF5, 0xF7, 0xFA)
}

func CardColor(dark bool) color.Color {
	if dark {
		return hexColor(0x16, 0x1B, 0x22)
	}
	return hexColor(0xFF, 0xFF, 0xFF)
}

func ButtonColor(dark bool) color.Color {
	if dark {
		return hexColor(0x21, 0x26, 0x2D)
	}
	return hexColor(0xEE, 0xF1, 0xF6)
}

func InputColor(dark bool) color.Color {
	if dark {
		return hexColor(0x0D, 0x11, 0x17)
	}
	return hexColor(0xFF, 0xFF, 0xFF)
}

func AccentColor(dark bool) color.Color {
	if dark {
		return hexColor(0x4C, 0x8D, 0xFF)
	}
	return hexColor(0x25, 0x63, 0xEB)
}

func TextColor(dark bool) color.Color {
	if dark {
		return hexColor(0xE6, 0xED, 0xF3)
	}
	return hexColor(0x1F, 0x24, 0x30)
}

func MutedTextColor(dark bool) color.Color {
	if dark {
		return hexColor(0x8B, 0x94, 0x9E)
	}
	return hexColor(0x6B, 0x72, 0x80)
}

func HoverColor(dark bool) color.Color {
	if dark {
		return hexColor(0x30, 0x36, 0x3D)
	}
	return hexColor(0xE6, 0xEC, 0xF4)
}

func SelectionColor(dark bool) color.Color {
	if dark {
		return hexColor(0x1C, 0x2B, 0x45)
	}
	return hexColor(0xDC, 0xE9, 0xFF)
}

func SeparatorColor(dark bool) color.Color {
	if dark {
		return hexColor(0x2A, 0x2F, 0x37)
	}
	return hexColor(0xE3, 0xE8, 0xEF)
}

func scrollBarColor(dark bool) color.Color {
	if dark {
		return color.NRGBA{R: 0x8B, G: 0x94, B: 0x9E, A: 0x66}
	}
	return color.NRGBA{R: 0x6B, G: 0x72, B: 0x80, A: 0x66}
}
