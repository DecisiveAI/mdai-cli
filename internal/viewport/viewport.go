package viewport

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	model struct {
		viewport  viewport.Model
		messages  chan string
		debug     chan string
		errs      chan error
		done      chan bool
		task      chan string
		debugMode bool
		quietMode bool
		spinner   spinner.Model
		title     string
	}
	responseMsg   string
	responseDebug string
	responseError error
	responseDone  bool
	responseTask  string
)

var (
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d3"))
	debugStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
)

func InitialModel(messages chan string, debug chan string, errs chan error, done chan bool, task chan string, debugMode bool, quietMode bool) model {
	vp := viewport.New(150, 25)
	vp.YOffset = 0
	return model{
		viewport:  vp,
		messages:  messages,
		debug:     debug,
		errs:      errs,
		done:      done,
		task:      task,
		debugMode: debugMode,
		quietMode: quietMode,
		spinner:   spinner.New(spinner.WithSpinner(spinner.Meter), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#800080")))),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		waitForMessages(m.messages),
		waitForDebug(m.debug),
		waitForErrors(m.errs),
		waitForTask(m.task),
		waitForDone(m.done),
	)
}

func waitForMessages(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

func waitForDebug(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return responseDebug(<-sub)
	}
}

func waitForErrors(sub chan error) tea.Cmd {
	return func() tea.Msg {
		return responseError(<-sub)
	}
}

func waitForDone(sub chan bool) tea.Cmd {
	return func() tea.Msg {
		return responseDone(<-sub)
	}
}

func waitForTask(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return responseTask(<-sub)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case responseMsg:
		m.viewport.SetContent(m.viewport.View() + "\n" + normalStyle.Render(string(msg)))
		m.viewport.GotoBottom()
		return m, tea.Batch(
			waitForMessages(m.messages),
			waitForDebug(m.debug),
			waitForErrors(m.errs),
			waitForTask(m.task),
			waitForDone(m.done),
		)
	case responseDebug:
		if m.debugMode {
			m.viewport.SetContent(m.viewport.View() + "\n" + debugStyle.Render(string(msg)))
			m.viewport.GotoBottom()
		}
		return m, tea.Batch(
			waitForMessages(m.messages),
			waitForDebug(m.debug),
			waitForErrors(m.errs),
			waitForTask(m.task),
			waitForDone(m.done),
		)
	case responseError:
		m.viewport.SetContent(m.viewport.View() + "\n" + errorStyle.Render(msg.Error()))
		m.viewport.GotoBottom()
		return m, tea.Quit
	case responseDone:
		m.title = "installation complete"
		return m, tea.Quit
	case responseTask:
		m.title = string(msg)
		return m, tea.Batch(
			waitForMessages(m.messages),
			waitForDebug(m.debug),
			waitForErrors(m.errs),
			waitForTask(m.task),
			waitForDone(m.done),
		)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		MarginBottom(0)

	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0).
		MarginTop(0).
		MarginLeft(0)

	title := titleStyle.Render(m.title)
	viewport := viewportStyle.Render(m.viewport.View())

	if m.quietMode {
		return fmt.Sprintf("%s %s", m.spinner.View(), title)
	}
	return fmt.Sprintf(
		"%s %s\n%s",
		m.spinner.View(),
		title,
		viewport,
	)
}
