package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Dicklesworthstone/system_resource_protection_script/internal/config"
	"github.com/Dicklesworthstone/system_resource_protection_script/internal/model"
	"github.com/Dicklesworthstone/system_resource_protection_script/internal/sampler"
)

// Model renders live samples from the sampler.
type Model struct {
	cfg       config.Config
	latest    model.Sample
	stream    <-chan model.Sample
	ctxCancel context.CancelFunc
	width     int
	height    int
}

func New(cfg config.Config) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	s := sampler.New(cfg.Interval)
	return &Model{
		cfg:       cfg,
		stream:    s.Stream(ctx),
		ctxCancel: cancel,
		width:     120,
		height:    40,
	}
}

// Messages
type (
	tickMsg   struct{}
	sampleMsg model.Sample
)

func tickCmd() tea.Cmd { return tea.Tick(time.Second/5, func(time.Time) tea.Msg { return tickMsg{} }) }

func (m *Model) Init() tea.Cmd { return tickCmd() }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.ctxCancel()
			return m, tea.Quit
		}
	case tickMsg:
		select {
		case samp, ok := <-m.stream:
			if ok {
				m.latest = samp
			}
		default:
		}
		return m, tickCmd()
	}
	return m, nil
}

// Styles
var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	gaugeFill   = "█"
	gaugeEmpty  = "░"
	cardStyle   = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("60")).
			Padding(0, 1).
			MarginRight(1)
)

func (m *Model) View() string {
	s := m.latest
	header := titleStyle.Render("System Resource Monitor (Go)") + "  " +
		subtleStyle.Render(s.Timestamp.Format("Mon Jan 2 15:04:05 MST 2006"))

	cpuCard := card("CPU",
		fmt.Sprintf("%s  load %.2f %.2f %.2f",
			gaugeBar(s.CPU.Total, 28),
			s.CPU.Load1, s.CPU.Load5, s.CPU.Load15))

	memPct := pct(s.Memory.UsedBytes, s.Memory.TotalBytes)
	memCard := card("Memory",
		fmt.Sprintf("%s  %.1f/%.1f GiB | Swap %3.0f%%",
			gaugeBar(memPct, 28),
			bytesToGiB(s.Memory.UsedBytes),
			bytesToGiB(s.Memory.TotalBytes),
			pct(s.Memory.SwapUsed, s.Memory.SwapTotal)))

	io := s.IO
	ioCard := card("IO / NET",
		fmt.Sprintf("Disk R/W: %.1f / %.1f MB/s   Net RX/TX: %.1f / %.1f Mb/s",
			io.DiskReadMBs, io.DiskWriteMBs, io.NetRxMbps, io.NetTxMbps))

	gpuCard := ""
	if len(s.GPUs) > 0 {
		lines := make([]string, 0, len(s.GPUs))
		for _, g := range s.GPUs {
			lines = append(lines,
				fmt.Sprintf("%s %4.0f%% mem:%4.0f/%-4.0fMiB %2.0f°C",
					truncate(g.Name, 10), g.Util, g.MemUsedMB, g.MemTotalMB, g.TempC))
		}
		gpuCard = card("GPU", strings.Join(lines, "\n"))
	}

	battCard := ""
	if s.Battery.Percent > 0 {
		battCard = card("Battery",
			fmt.Sprintf("%.0f%% (%s)", s.Battery.Percent, s.Battery.State))
	}

	topTable := card("Top CPU",
		renderTable([]string{"cmd", "pid", "ni", "cpu", "mem"},
			s.Top, 10))

	throttledTable := card("Throttled (nice>0)",
		renderTable([]string{"cmd", "pid", "ni", "cpu", "mem"},
			s.Throttled, 6))

	cgTable := ""
	if len(s.Cgroups) > 0 {
		rows := make([]string, 0, min(6, len(s.Cgroups)))
		for i := 0; i < min(6, len(s.Cgroups)); i++ {
			cg := s.Cgroups[i]
			rows = append(rows, fmt.Sprintf("%-18s %5.1f%%", truncate(cg.Name, 18), cg.CPU))
		}
		cgTable = card("Top cgroups", strings.Join(rows, "\n"))
	}

	columns := []string{cpuCard, memCard, ioCard}
	if gpuCard != "" {
		columns = append(columns, gpuCard)
	}
	if battCard != "" {
		columns = append(columns, battCard)
	}

	line1 := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	line2 := lipgloss.JoinHorizontal(lipgloss.Top, topTable, throttledTable, cgTable)

	return lipgloss.JoinVertical(lipgloss.Left, header, line1, line2)
}

// Helpers
func gaugeBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int((pct / 100) * float64(width))
	if filled > width {
		filled = width
	}
	return fmt.Sprintf("[%s%s] %5.1f%%",
		strings.Repeat(gaugeFill, filled),
		strings.Repeat(gaugeEmpty, width-filled),
		pct)
}

func card(title, body string) string {
	titleStr := labelStyle.Render(title)
	content := titleStr + "\n" + body
	return cardStyle.Render(content)
}

func renderTable(headers []string, rows []model.Process, limit int) string {
	max := min(limit, len(rows))
	var b strings.Builder
	fmt.Fprintf(&b, "%-18s %-6s %-3s %-6s %-6s\n", headers[0], headers[1], headers[2], headers[3], headers[4])
	for i := 0; i < max; i++ {
		r := rows[i]
		fmt.Fprintf(&b, "%-18s %-6d %3d %6.1f %6.1f\n",
			truncate(r.Command, 18), r.PID, r.Nice, r.CPU, r.Memory)
	}
	return strings.TrimRight(b.String(), "\n")
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func pct(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) * 100 / float64(total)
}

func bytesToGiB(b uint64) float64 { return float64(b) / (1024 * 1024 * 1024) }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RunTUI starts the Bubble Tea program.
func RunTUI(cfg config.Config) error {
	prog := tea.NewProgram(New(cfg), tea.WithAltScreen())
	_, err := prog.Run()
	return err
}
