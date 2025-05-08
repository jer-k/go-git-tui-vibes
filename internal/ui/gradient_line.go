package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type GradientLine struct {
	startColor string
	endColor   string
	width      int
}

func NewGradientLine(startColor, endColor string, width int) GradientLine {
	return GradientLine{
		startColor: startColor,
		endColor:   endColor,
		width:      width,
	}
}

func hexToRGB(hex string) (int, int, int) {
	hex = hex[1:] // Remove #
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func rgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func interpolateColor(startR, startG, startB, endR, endG, endB int, t float64) (int, int, int) {
	r := int(math.Round(float64(startR) + t*float64(endR-startR)))
	g := int(math.Round(float64(startG) + t*float64(endG-startG)))
	b := int(math.Round(float64(startB) + t*float64(endB-startB)))
	return r, g, b
}

func (g GradientLine) Render() string {
	if g.width <= 0 {
		return ""
	}

	startR, startG, startB := hexToRGB(g.startColor)
	endR, endG, endB := hexToRGB(g.endColor)

	segments := make([]string, g.width)
	for i := 0; i < g.width; i++ {
		t := float64(i) / float64(g.width-1)
		r, gr, b := interpolateColor(startR, startG, startB, endR, endG, endB, t)
		color := rgbToHex(r, gr, b)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		segments[i] = style.Render("─")
	}

	return strings.Join(segments, "")
}

func (g GradientLine) RenderWithPadding(leftPad, rightPad int) string {
	leftPadding := strings.Repeat(" ", leftPad)
	rightPadding := strings.Repeat(" ", rightPad)
	return leftPadding + g.Render() + rightPadding
}