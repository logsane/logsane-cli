// Package ansirgb converts color.Color's to ANSI colors
package main

import (
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"image/color"
)

func blendAllColors(colors []color.Color) colorful.Color {

	finalColor, _ := colorful.MakeColor(colors[0])

	for _, v := range colors {
		converted, _ := colorful.MakeColor(v)
		finalColor = finalColor.BlendLab(converted, 0.5)
	}
	return finalColor
}

func colorfulToAnsi(color1 color.Color) termenv.Color {
	return termenv.ColorProfile().FromColor(color1)
}
