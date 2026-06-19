package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// customTheme overrides the default theme for a more modern feel.
type customTheme struct {
	fyne.Theme
}

// NewCustomTheme returns a new instance of the custom theme.
func NewCustomTheme() fyne.Theme {
	return &customTheme{Theme: theme.DefaultTheme()}
}

func (c *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		if variant == theme.VariantLight {
			return color.NRGBA{R: 245, G: 245, B: 247, A: 255}
		}
		return color.NRGBA{R: 28, G: 28, B: 30, A: 255}
	}
	if name == theme.ColorNamePrimary {
		return color.NRGBA{R: 10, G: 132, B: 255, A: 255}
	}
	return c.Theme.Color(name, variant)
}

func (c *customTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 8.0 // slightly more padding for a spacious feel
	}
	return c.Theme.Size(name)
}
