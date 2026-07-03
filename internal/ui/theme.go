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
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 20
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func darkColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return BackgroundColor(true)
	case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameOverlayBackground:
		return CardColor(true)
	case theme.ColorNameForeground:
		return TextColor(true)
	case theme.ColorNamePrimary, theme.ColorNameFocus:
		return AccentColor(true)
	case theme.ColorNameHover, theme.ColorNamePressed, theme.ColorNameSelection:
		return SelectionColor(true)
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return MutedTextColor(true)
	case theme.ColorNameSeparator, theme.ColorNameShadow:
		return hexColor(0x3A, 0x3A, 0x3A)
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func lightColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return BackgroundColor(false)
	case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameOverlayBackground:
		return CardColor(false)
	case theme.ColorNameForeground:
		return TextColor(false)
	case theme.ColorNamePrimary, theme.ColorNameFocus:
		return AccentColor(false)
	case theme.ColorNameHover, theme.ColorNamePressed, theme.ColorNameSelection:
		return SelectionColor(false)
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return MutedTextColor(false)
	case theme.ColorNameSeparator, theme.ColorNameShadow:
		return hexColor(0xDD, 0xDD, 0xDD)
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func hexColor(r, g, b uint8) color.Color {
	return color.NRGBA{R: r, G: g, B: b, A: 0xFF}
}

func BackgroundColor(dark bool) color.Color {
	if dark {
		return hexColor(0x11, 0x18, 0x27)
	}
	return hexColor(0xF7, 0xF8, 0xFA)
}

func CardColor(dark bool) color.Color {
	if dark {
		return hexColor(0x1F, 0x29, 0x37)
	}
	return hexColor(0xFF, 0xFF, 0xFF)
}

func AccentColor(dark bool) color.Color {
	if dark {
		return hexColor(0x38, 0xBD, 0xF8)
	}
	return hexColor(0x25, 0x63, 0xEB)
}

func TextColor(dark bool) color.Color {
	if dark {
		return hexColor(0xF9, 0xFA, 0xFB)
	}
	return hexColor(0x11, 0x18, 0x27)
}

func MutedTextColor(dark bool) color.Color {
	if dark {
		return hexColor(0x9C, 0xA3, 0xAF)
	}
	return hexColor(0x6B, 0x72, 0x80)
}

func SelectionColor(dark bool) color.Color {
	if dark {
		return hexColor(0x1E, 0x3A, 0x8A)
	}
	return hexColor(0xDB, 0xEA, 0xFE)
}
