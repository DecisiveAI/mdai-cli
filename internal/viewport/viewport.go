package viewport

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type channels struct {
	messages chan string
	debug    chan string
	errs     chan error
	done     chan bool
	task     chan string
}

type modes struct {
	debug bool
	quiet bool
}

type (
	model struct {
		viewport viewport.Model
		channels channels
		modes    modes
		styles   styles
		spinner  spinner.Model
		title    string
		content  *strings.Builder
		start    time.Time
		ready    bool
		hasError bool
	}
	responseMsg   string
	responseDebug string
	responseError error
	responseDone  bool
	responseTask  string
)

type styles struct {
	err,
	normal,
	debug,
	title,
	viewport lipgloss.Style
}

func InitialModel(messages chan string, debug chan string, errs chan error, done chan bool, task chan string, debugMode bool, quietMode bool) model {
	return model{
		channels: channels{messages, debug, errs, done, task},
		modes:    modes{debugMode, quietMode},
		styles: styles{
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d3")),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")),
			lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(0),
			lipgloss.NewStyle().Padding(0).MarginTop(0).MarginLeft(0),
		},
		spinner: spinner.New(spinner.WithSpinner(spinner.Meter), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#800080")))),
		content: &strings.Builder{},
		start:   time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	m.writeContent("process started...", m.styles.normal, false)
	return tea.Batch(
		m.spinner.Tick,
		m.waitFor("message"),
		m.waitFor("debug"),
		m.waitFor("error"),
		m.waitFor("task"),
		m.waitFor("done"),
	)
}

func (m model) waitFor(what string) tea.Cmd {
	switch what {
	case "message":
		return func() tea.Msg { return responseMsg(<-m.channels.messages) }
	case "error":
		return func() tea.Msg { return responseError(<-m.channels.errs) }
	case "debug":
		return func() tea.Msg { return responseDebug(<-m.channels.debug) }
	case "task":
		return func() tea.Msg { return responseTask(<-m.channels.task) }
	case "done":
		return func() tea.Msg { return responseDone(<-m.channels.done) }
	}
	return nil
}

func (m model) writeContent(content string, style lipgloss.Style, timing bool) {
	format := "[%s] %s"
	args := []any{time.Now().Format(time.DateTime), content}
	if timing {
		format += " (%s)"
		args = append(args, time.Since(m.start).String())
	}
	_, _ = fmt.Fprintln(m.content,
		style.Width(m.viewport.Width-10).Render(
			fmt.Sprintf(format, args...),
		),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case responseMsg:
		m.writeContent(string(msg), m.styles.normal, true)
		m.viewport.SetContent(m.content.String())
		m.viewport.GotoBottom()
		m.start = time.Now()
	case responseDebug:
		if m.modes.debug {
			m.writeContent(string(msg), m.styles.debug, false)
			m.viewport.SetContent(m.content.String())
			m.viewport.GotoBottom()
		}
	case responseError:
		m.hasError = true
		m.channels.done <- true
		m.writeContent(msg.Error(), m.styles.err, false)
		m.viewport.SetContent(m.content.String())
		m.viewport.GotoBottom()
		return m, m.waitFor("done")
	case responseDone:
		m.title = "process complete"
		if m.hasError {
			m.title += " with error"
		}
		m.viewport.GotoBottom()
		return m, tea.Quit
	case responseTask:
		m.title = string(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-2, msg.Height-3)
			m.viewport.Style = m.styles.viewport
			m.viewport.YPosition = 0
			m.viewport.HighPerformanceRendering = false
			m.viewport.SetContent(m.content.String())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	return m, tea.Batch(
		m.waitFor("message"),
		m.waitFor("debug"),
		m.waitFor("error"),
		m.waitFor("task"),
		m.waitFor("done"),
	)
}

func (m model) View() string {
	title := m.styles.title.Render(m.title)
	if m.modes.quiet {
		return fmt.Sprintf("%s %s", m.spinner.View(), title)
	}
	return fmt.Sprintf(
		"%s %s (%s)\n%s",
		m.spinner.View(),
		title,
		time.Since(m.start).Truncate(time.Second).String(),
		m.viewport.View(),
	)
}
