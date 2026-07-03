package ui

import (
	"fmt"
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
		return 12
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
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
		return hexColor(0x1F, 0x1F, 0x1F)
	}
	return hexColor(0xF3, 0xF3, 0xF3)
}

func CardColor(dark bool) color.Color {
	if dark {
		return hexColor(0x2B, 0x2B, 0x2B)
	}
	return hexColor(0xFF, 0xFF, 0xFF)
}

func AccentColor(dark bool) color.Color {
	if dark {
		return hexColor(0x60, 0xCD, 0xFF)
	}
	return hexColor(0x00, 0x78, 0xD4)
}

func TextColor(dark bool) color.Color {
	if dark {
		return hexColor(0xFF, 0xFF, 0xFF)
	}
	return hexColor(0x1A, 0x1A, 0x1A)
}

func InvertedTextColor(dark bool) color.Color {
	if dark {
		return hexColor(0x1A, 0x1A, 0x1A)
	}
	return hexColor(0xFF, 0xFF, 0xFF)
}

func MutedTextColor(dark bool) color.Color {
	if dark {
		return hexColor(0x9F, 0xA7, 0xAE)
	}
	return hexColor(0x6B, 0x72, 0x80)
}

func AccentTextColor(dark bool) color.Color {
	if dark {
		return hexColor(0x1A, 0x1A, 0x1A)
	}
	return hexColor(0xFF, 0xFF, 0xFF)
}

func ButtonColor(dark bool) color.Color {
	if dark {
		return hexColor(0x2F, 0x2F, 0x2F)
	}
	return hexColor(0xFF, 0xFF, 0xFF)
}

func SelectionColor(dark bool) color.Color {
	if dark {
		return hexColor(0x32, 0x4D, 0x5D)
	}
	return hexColor(0xD8, 0xEA, 0xF8)
}

func ColorHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}
