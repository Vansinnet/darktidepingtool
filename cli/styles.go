package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleRed    = lipgloss.Color("#CC0000")
	subGreen    = lipgloss.Color("#4D9D4D")
	goodGreen   = lipgloss.Color("#4DFF4D")
	okYellow    = lipgloss.Color("#FFDD00")
	poorOrange  = lipgloss.Color("#FF9933")
	badRed      = lipgloss.Color("#FF3333")
	borderRed   = lipgloss.Color("#660000")
	dimText     = lipgloss.Color("#555555")
	mutedText   = lipgloss.Color("#888888")
	brightText  = lipgloss.Color("#D4D4D4")
	bgDark      = lipgloss.Color("#1A0A0A")
	boxBg       = lipgloss.Color("#1A0810")
	scanGreen   = lipgloss.Color("#4D9D4D")
	darkBorder  = lipgloss.Color("#330000")
)

func pingColor(ms int) lipgloss.Color {
	switch {
	case ms <= 80:
		return goodGreen
	case ms <= 120:
		return okYellow
	case ms <= 180:
		return poorOrange
	default:
		return badRed
	}
}

func pingBgColor(ms int) lipgloss.Color {
	switch {
	case ms <= 80:
		return lipgloss.Color("#0D330D")
	case ms <= 120:
		return lipgloss.Color("#332B00")
	case ms <= 180:
		return lipgloss.Color("#331A00")
	default:
		return lipgloss.Color("#330D0D")
	}
}

func pingClass(ms int) string {
	switch {
	case ms <= 80:
		return "GOOD"
	case ms <= 120:
		return "ACCEPTABLE"
	case ms <= 180:
		return "HIGH"
	default:
		return "VERY HIGH"
	}
}

func jitterColor(ms int) lipgloss.Color {
	switch {
	case ms <= 15:
		return goodGreen
	case ms <= 30:
		return okYellow
	case ms <= 60:
		return poorOrange
	default:
		return badRed
	}
}

func stabilityDesc(score int) string {
	switch {
	case score >= 90:
		return "EXCELLENT"
	case score >= 75:
		return "STABLE"
	case score >= 60:
		return "MODERATE"
	case score >= 40:
		return "UNSTABLE"
	default:
		return "POOR"
	}
}

func stabilityScore(jitter int, packetLoss float64) int {
	jitterScore := 100 - (jitter / 2)
	if jitterScore < 0 {
		jitterScore = 0
	}
	lossScore := 100 - int(packetLoss*5)
	if lossScore < 0 {
		lossScore = 0
	}
	score := (jitterScore + lossScore) / 2
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

var (
	emptyStyle = lipgloss.NewStyle()

	titleStyle = lipgloss.NewStyle().
			Foreground(titleRed).
			Bold(true).
			Align(lipgloss.Center)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(subGreen).
			Align(lipgloss.Center)

	groupHeaderStyle = lipgloss.NewStyle().
				Foreground(titleRed).
				Bold(true).
				Padding(0, 1)

	scanningStyle = lipgloss.NewStyle().
			Foreground(scanGreen).
			Blink(true).
			Align(lipgloss.Center)

	boxBorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(borderRed).
			Padding(1, 2)

	statsPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(borderRed).
			Padding(1, 2)

	statLabelStyle = lipgloss.NewStyle().
			Foreground(titleRed).
			Bold(true).
			Align(lipgloss.Center)

	statValueStyle = lipgloss.NewStyle().
			Foreground(subGreen).
			Bold(true).
			Align(lipgloss.Center)

	legendStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(darkBorder).
			Padding(1, 2)

	legendTitleStyle = lipgloss.NewStyle().
				Foreground(titleRed).
				Bold(true)

	legendItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	recommendStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(goodGreen).
			Padding(1, 2)

	recTitleStyle = lipgloss.NewStyle().
			Foreground(mutedText).
			Bold(true).
			Align(lipgloss.Center)

	recServerStyle = lipgloss.NewStyle().
			Foreground(goodGreen).
			Bold(true).
			Align(lipgloss.Center)

	controlsStyle = lipgloss.NewStyle().
			Foreground(mutedText).
			Align(lipgloss.Center)

	regionNameStyle = lipgloss.NewStyle().
			Foreground(brightText).
			Bold(true)

	regionCodeStyle = lipgloss.NewStyle().
			Foreground(dimText)

	detailStyle = lipgloss.NewStyle().
			Foreground(dimText)

	testingStyle = lipgloss.NewStyle().
			Foreground(dimText)
)

func pingValueBox(ms int) string {
	if ms >= 9999 {
		return lipgloss.NewStyle().
			Foreground(badRed).
			Background(boxBg).
			Padding(0, 2).
			Render("FAIL")
	}
	text := fmt.Sprintf(" %d ms ", ms)
	return lipgloss.NewStyle().
		Foreground(pingColor(ms)).
		Background(pingBgColor(ms)).
		Padding(0, 1).
		Render(text)
}

func stabilityBar(score int) string {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	filled := score / 10
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < 10; i++ {
		bar += "░"
	}
	barColor := goodGreen
	switch {
	case score < 40:
		barColor = badRed
	case score < 60:
		barColor = poorOrange
	case score < 75:
		barColor = okYellow
	}
	return lipgloss.NewStyle().Foreground(barColor).Render(bar)
}

func jitterDot(ms int) string {
	dot := "●"
	return lipgloss.NewStyle().Foreground(jitterColor(ms)).Render(dot)
}

func lossLabel(packetLoss float64) string {
	if packetLoss > 0 {
		c := poorOrange
		if packetLoss >= 30 {
			c = badRed
		}
		return lipgloss.NewStyle().Foreground(c).Render(fmt.Sprintf("⚠ %.0f%%", packetLoss))
	}
	return lipgloss.NewStyle().Foreground(subGreen).Render("✓ 0%")
}
