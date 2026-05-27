package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	results   map[string]PingResult
	scanning  bool
	scanDone  bool
	scanTotal int
	scanCount int
	scanCh    chan regionResultMsg
	width     int
	height    int
	bestPing  int
	bestName  string
}

type scanStartedMsg struct {
	ch chan regionResultMsg
}

type regionResultMsg struct {
	region Region
	result PingResult
	index  int
	total  int
}

type scanCompleteMsg struct{}

func startScanCmd() tea.Msg {
	ch := make(chan regionResultMsg, 100)
	go func() {
		index := 0
		total := totalRegions()
		for _, group := range regionGroups {
			for _, server := range group.Servers {
				result := measureRegion(server, pingSamples, pingTimeout)
				ch <- regionResultMsg{region: server, result: result, index: index, total: total}
				index++
			}
		}
		close(ch)
	}()
	return scanStartedMsg{ch: ch}
}

func waitForResult(ch chan regionResultMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return scanCompleteMsg{}
		}
		return msg
	}
}

func initialModel() model {
	return model{
		results: make(map[string]PingResult),
	}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		return startScanCmd()
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scanStartedMsg:
		m.scanning = true
		m.scanDone = false
		m.scanTotal = totalRegions()
		m.scanCount = 0
		m.results = make(map[string]PingResult)
		m.bestPing = 99999
		m.bestName = ""
		m.scanCh = msg.ch
		return m, waitForResult(m.scanCh)

	case regionResultMsg:
		m.results[msg.region.Code] = msg.result
		m.scanCount = msg.index + 1

		if msg.result.Samples > 0 && msg.result.Median < m.bestPing {
			m.bestPing = msg.result.Median
			m.bestName = fmt.Sprintf("%s — %s (%s)", msg.region.Group, msg.region.Name, msg.region.Code)
		}

		return m, waitForResult(m.scanCh)

	case scanCompleteMsg:
		m.scanning = false
		m.scanDone = true
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if !m.scanning {
				return m, func() tea.Msg {
					return startScanCmd()
				}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	sections := []string{}

	sections = append(sections, m.viewHeader())
	sections = append(sections, m.viewStats())

	for _, group := range regionGroups {
		sections = append(sections, m.viewGroup(group))
	}

	sections = append(sections, m.viewLegend())

	if m.scanDone {
		sections = append(sections, m.viewRecommendation())
	}

	sections = append(sections, m.viewControls())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) viewHeader() string {
	title := titleStyle.Render("D A R K T I D E   P I N G   T O O L")
	subtitle := subtitleStyle.Render("SERVER REGION LATENCY TESTER")

	border := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderRed).
		Padding(0, 4).
		Align(lipgloss.Center)

	content := lipgloss.JoinVertical(lipgloss.Center, title, subtitle)
	return border.Render(content)
}

func (m model) viewStats() string {
	if m.scanCount == 0 && !m.scanDone {
		return ""
	}

	var allMedians []int
	worstPing := 0

	for _, group := range regionGroups {
		for _, s := range group.Servers {
			if r, ok := m.results[s.Code]; ok && r.Samples > 0 {
				allMedians = append(allMedians, r.Median)
				if r.Median > worstPing {
					worstPing = r.Median
				}
			}
		}
	}

	avgVal := "--"
	bestVal := "--"
	worstVal := "--"

	if len(allMedians) > 0 {
		sum := 0
		for _, v := range allMedians {
			sum += v
		}
		avgVal = fmt.Sprintf("%d ms", sum/len(allMedians))
		bestVal = fmt.Sprintf("%d ms", m.bestPing)
		worstVal = fmt.Sprintf("%d ms", worstPing)
	}

	col := func(label, value string) string {
		l := statLabelStyle.Render(label)
		v := statValueStyle.Render(value)
		return lipgloss.NewStyle().Width(24).Align(lipgloss.Center).Render(lipgloss.JoinVertical(lipgloss.Center, l, v))
	}

	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		col("AVERAGE LATENCY", avgVal),
		col("BEST REGION", bestVal),
		col("WORST REGION", worstVal),
	)

	return statsPanelStyle.Render(cols)
}

func (m model) viewGroup(group RegionGroup) string {
	header := groupHeaderStyle.Render(fmt.Sprintf("■ %s", group.Label))

	rows := make([]string, 0, len(group.Servers))
	for _, server := range group.Servers {
		rows = append(rows, m.viewServerRow(server))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, header, body)

	wrapper := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(darkBorder).
		Padding(0, 1).
		Width(76)

	return wrapper.Render(content)
}

func (m model) viewServerRow(server Region) string {
	r, hasResult := m.results[server.Code]

	name := fmt.Sprintf("%-14s", server.Name)
	code := fmt.Sprintf("%-14s", server.Code)

	if !hasResult {
		return fmt.Sprintf("  %s%s  %s",
			name, code,
			testingStyle.Render("Testing..."),
		)
	}

	if r.Error != "" && r.Samples == 0 {
		return fmt.Sprintf("  %s%s  %s",
			name, code,
			pingValueBox(9999),
		)
	}

	sc := stabilityScore(r.Jitter, r.PacketLoss)
	bar := stabilityBar(sc)
	dot := jitterDot(r.Jitter)
	loss := lossLabel(r.PacketLoss)
	class := pingClass(r.Median)

	classStyle := lipgloss.NewStyle().Foreground(pingColor(r.Median)).Width(10)

	mainLine := fmt.Sprintf("  %s%s  %s  %s  %s",
		name,
		code,
		pingValueBox(r.Median),
		bar,
		classStyle.Render(class),
	)

	detailLine := fmt.Sprintf("    p10:%-4d p90:%-4d %s jitter:%-3dms  %s",
		r.P10, r.P90, dot, r.Jitter, loss,
	)
	return lipgloss.NewStyle().Padding(0, 1).Render(
		lipgloss.JoinVertical(lipgloss.Left, mainLine, detailLine),
	)
}

func (m model) viewLegend() string {
	title := legendTitleStyle.Render("█ LATENCY GUIDE")

	items := []string{
		legendItemStyle.Copy().Foreground(goodGreen).Render("■ GOOD (0-80ms)"),
		legendItemStyle.Copy().Foreground(okYellow).Render("■ ACCEPTABLE (81-120ms)"),
		legendItemStyle.Copy().Foreground(poorOrange).Render("■ HIGH (121-180ms)"),
		legendItemStyle.Copy().Foreground(badRed).Render("■ VERY HIGH (180ms+)"),
	}

	itemLine := lipgloss.JoinHorizontal(lipgloss.Top, items...)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		itemLine,
	)

	return legendStyle.Render(content)
}

func (m model) viewRecommendation() string {
	title := recTitleStyle.Render("LOWEST LATENCY REGION")
	server := recServerStyle.Render(m.bestName)
	ping := lipgloss.NewStyle().Foreground(mutedText).Align(lipgloss.Center).Render(
		fmt.Sprintf("LATENCY: %d ms", m.bestPing),
	)

	content := lipgloss.JoinVertical(lipgloss.Center, title, server, ping)
	return recommendStyle.Render(content)
}

func (m model) viewControls() string {
	scanLabel := ""
	if m.scanning {
		scanLabel = lipgloss.NewStyle().Foreground(scanGreen).Render("⏳ Scanning...")
	} else if m.scanDone {
		scanLabel = lipgloss.NewStyle().Foreground(mutedText).Render("[R] Scan again")
	} else {
		scanLabel = lipgloss.NewStyle().Foreground(mutedText).Render("[R] Scan again")
	}

	quitLabel := lipgloss.NewStyle().Foreground(mutedText).Render("[Q] Quit")

	line := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(30).Align(lipgloss.Center).Render(scanLabel),
		lipgloss.NewStyle().Width(20).Align(lipgloss.Center).Render(quitLabel),
	)

	return controlsStyle.Render(line)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}


