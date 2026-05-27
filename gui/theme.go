package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

var ForceLightMode bool = false

type USSTheme struct{}

var _ fyne.Theme = (*USSTheme)(nil)

// Dark colors
var colorBgDark = color.NRGBA{R: 0x0d, G: 0x11, B: 0x17, A: 0xff}
var colorSurfaceDark = color.NRGBA{R: 0x16, G: 0x1e, B: 0x2e, A: 0xff}
var colorTextDark = color.NRGBA{R: 0xe2, G: 0xe8, B: 0xf0, A: 0xff}
var colorSubtextDark = color.NRGBA{R: 0x64, G: 0x74, B: 0x8b, A: 0xff}
var colorBorderDark = color.NRGBA{R: 0x1e, G: 0x2d, B: 0x45, A: 0xff}

// Light colors
var colorBgLight = color.NRGBA{R: 0xf3, G: 0xf4, B: 0xf6, A: 0xff}
var colorSurfaceLight = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
var colorTextLight = color.NRGBA{R: 0x1f, G: 0x29, B: 0x37, A: 0xff}
var colorSubtextLight = color.NRGBA{R: 0x6b, G: 0x72, B: 0x80, A: 0xff}
var colorBorderLight = color.NRGBA{R: 0xe5, G: 0xe7, B: 0xeb, A: 0xff}

// Shared
var colorPrimary = color.NRGBA{R: 0x00, G: 0xc8, B: 0xff, A: 0xff}
var colorPrimaryDarker = color.NRGBA{R: 0x02, G: 0x84, B: 0xc7, A: 0xff} // Sky-Blue-600 – angenehmes Mittelblau für Light Mode

func (t *USSTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	isDark := variant == theme.VariantDark
	if ForceLightMode {
		isDark = false
	} else {
		isDark = true // force dark by default unless OS tells us otherwise? No, we force it by our bool. Let's just use ForceLightMode
	}

	switch name {
	case theme.ColorNameBackground:
		if isDark { return colorBgDark } else { return colorBgLight }
	case theme.ColorNameButton:
		if isDark { return colorSurfaceDark } else { return colorSurfaceLight }
	case theme.ColorNameDisabledButton:
		if isDark { return colorBorderDark } else { return colorBorderLight }
	case theme.ColorNameForeground:
		if isDark { return colorTextDark } else { return colorTextLight }
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		if isDark { return colorSubtextDark } else { return colorSubtextLight }
	case theme.ColorNamePrimary:
		if isDark { return colorPrimary } else { return colorPrimaryDarker }
	case theme.ColorNameFocus:
		// In Light Mode: helles, transparentes Blau statt vollem Primärfarbton,
		// damit der Select-Button-Hintergrund nicht übermäßig schwer wirkt.
		if isDark { return colorPrimary } else { return color.NRGBA{R: 0x02, G: 0x84, B: 0xc7, A: 0x22} }
	case theme.ColorNameHover:
		if isDark { return color.NRGBA{R: 0x1e, G: 0x2d, B: 0x45, A: 0xff} } else { return color.NRGBA{R: 0xe5, G: 0xe7, B: 0xeb, A: 0xff} }
	case theme.ColorNameSelection:
		if isDark {
			return color.NRGBA{R: 0x00, G: 0xc8, B: 0xff, A: 0x30}
		} else {
			return color.NRGBA{R: 0x02, G: 0x84, B: 0xc7, A: 0x28}
		}
	case theme.ColorNameShadow:
		if isDark { return color.NRGBA{A: 0x44} } else { return color.NRGBA{A: 0x11} }
	case theme.ColorNameInputBackground, theme.ColorNameHeaderBackground, theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		if isDark { return colorSurfaceDark } else { return colorSurfaceLight }
	case theme.ColorNameSeparator, theme.ColorNameScrollBar:
		if isDark { return colorBorderDark } else { return colorBorderLight }
	case theme.ColorNameSuccess:
		if isDark {
			return color.NRGBA{R: 0x00, G: 0xff, B: 0x44, A: 0x77} // Translucent Neon for Dark Mode
		} else {
			return color.NRGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff} // Solid deep Emerald for Light Mode
		}
	case theme.ColorNameWarning:
		if isDark {
			return color.NRGBA{R: 0xff, G: 0x88, B: 0x00, A: 0x77}
		} else {
			return color.NRGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff}
		}
	case theme.ColorNameError:
		if isDark {
			return color.NRGBA{R: 0xff, G: 0x00, B: 0x33, A: 0x77} // Translucent Red for Dark Mode
		} else {
			return color.NRGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff} // Solid deep Red for Light Mode
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *USSTheme) Font(style fyne.TextStyle) fyne.Resource { return theme.DefaultTheme().Font(style) }
func (t *USSTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(name) }
func (t *USSTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding: return 8
	case theme.SizeNameInnerPadding: return 6
	case theme.SizeNameText: return 13
	case theme.SizeNameSubHeadingText: return 15
	case theme.SizeNameHeadingText: return 20
	case theme.SizeNameLineSpacing: return 4
	case theme.SizeNameScrollBar: return 6
	case theme.SizeNameScrollBarSmall: return 3
	case theme.SizeNameSeparatorThickness: return 1
	case theme.SizeNameInputBorder: return 2
	}
	return theme.DefaultTheme().Size(name)
}

