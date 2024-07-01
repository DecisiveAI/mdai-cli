package processmanager

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	items                []string
	index                int
	width                int
	height               int
	spinner              spinner.Model
	progress             progress.Model
	helmchartinstalldone bool
	runfunc              func(string) error
	manifestapplyfunc    func() error
	addreposfunc         func() error
}

var (
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

func NewModel(items []string, runfunc func(string) error, manifestapplyfunc func() error, addreposfunc func() error) tea.Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return model{
		items:             items,
		spinner:           s,
		progress:          p,
		runfunc:           runfunc,
		manifestapplyfunc: manifestapplyfunc,
		addreposfunc:      addreposfunc,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.addrepos(),
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case helmReposAddedMsg:
		return m, install(m.items[m.index], m.runfunc)
	case installedPkgMsg:
		if m.index >= len(m.items)-1 {
			m.helmchartinstalldone = true
			return m, tea.Batch(
				m.progress.SetPercent(100),
				tea.Printf("%s %s", checkMark, m.items[m.index]),
				m.applymanifest(),
			)
		}

		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.items)-1))

		m.index++
		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s %s", checkMark, m.items[m.index]),
			install(m.items[m.index], m.runfunc),
		)
	case manifestAppliedMsg:
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	n := len(m.items)
	w := lipgloss.Width(strconv.Itoa(n))

	if m.helmchartinstalldone {
		return doneStyle.Render(fmt.Sprintf("Done! Installed %d helm charts.\n", n))
	}

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.index+1, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+pkgCount))

	pkgName := currentPkgNameStyle.Render(m.items[m.index])
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Installing " + pkgName)

	cellsRemaining := max(0, 100-lipgloss.Width(spin+info+prog+pkgCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + pkgCount
}

type (
	helmReposAddedMsg  string
	installedPkgMsg    string
	manifestAppliedMsg string
)

func install(pkg string, runfunc func(string) error) tea.Cmd {
	return func() tea.Msg {
		if err := runfunc(pkg); err != nil {
			tea.Printf("error: %s\n", err.Error())
			return tea.Quit
		}
		return installedPkgMsg(pkg)
	}
}

func (m model) addrepos() tea.Cmd {
	return func() tea.Msg {
		if err := m.addreposfunc(); err != nil {
			tea.Printf("error: %s\n", err.Error())
			return tea.Quit
		}
		return helmReposAddedMsg("helm repos added")
	}
}

func (m model) applymanifest() tea.Cmd {
	return func() tea.Msg {
		if err := m.manifestapplyfunc(); err != nil {
			tea.Printf("error: %s\n", err.Error())
			return tea.Quit
		}
		return manifestAppliedMsg("mdai-operator")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
