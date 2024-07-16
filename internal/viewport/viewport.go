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
		ready     bool
		content   string
		start     time.Time
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
	titleStyle  = lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(0)
)

func InitialModel(messages chan string, debug chan string, errs chan error, done chan bool, task chan string, debugMode bool, quietMode bool) model {
	return model{
		messages:  messages,
		debug:     debug,
		errs:      errs,
		done:      done,
		task:      task,
		debugMode: debugMode,
		quietMode: quietMode,
		spinner:   spinner.New(spinner.WithSpinner(spinner.Meter), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#800080")))),
		content:   "[" + time.Now().Format(time.DateTime) + "] process started...",
		start:     time.Now(),
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
		ts := "[" + time.Now().Format(time.DateTime) + "] "
		m.content = strings.Join([]string{m.content, normalStyle.Width(m.viewport.Width - 10).Render(ts + string(msg) + " (" + time.Since(m.start).String() + ")")}, "\n")
		m.viewport.SetContent(m.content)
		if len(m.content) > m.viewport.VisibleLineCount() {
			m.viewport.GotoBottom()
		}
		m.start = time.Now()
		return m, tea.Batch(
			waitForMessages(m.messages),
			waitForDebug(m.debug),
			waitForErrors(m.errs),
			waitForTask(m.task),
			waitForDone(m.done),
		)
	case responseDebug:
		if m.debugMode {
			ts := "[" + time.Now().Format(time.DateTime) + "] "
			m.content = strings.Join([]string{m.content, debugStyle.Width(m.viewport.Width - 10).Render(ts + string(msg))}, "\n")
			m.viewport.SetContent(m.content)
			if len(m.content) > m.viewport.VisibleLineCount() {
				m.viewport.GotoBottom()
			}
		}
		return m, tea.Batch(
			waitForMessages(m.messages),
			waitForDebug(m.debug),
			waitForErrors(m.errs),
			waitForTask(m.task),
			waitForDone(m.done),
		)
	case responseError:
		ts := "[" + time.Now().Format(time.DateTime) + "] "
		m.content = strings.Join([]string{m.content, errorStyle.Width(m.viewport.Width - 10).Render(ts + msg.Error())}, "\n")
		m.viewport.SetContent(m.content)
		if len(m.content) > m.viewport.VisibleLineCount() {
			m.viewport.GotoBottom()
		}
		return m, tea.Quit
	case responseDone:
		m.title = "process complete"
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
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-2, msg.Height-3)
			m.viewport.Style = lipgloss.NewStyle().
				//	Border(lipgloss.RoundedBorder()).
				Padding(0).
				MarginTop(0).
				MarginLeft(0)
			m.viewport.YPosition = 0
			m.viewport.HighPerformanceRendering = false
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	return m, nil
}

func (m model) View() string {
	title := titleStyle.Render(m.title)
	if m.quietMode {
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
