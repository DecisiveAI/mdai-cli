package editor

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type editorFinishedMsg struct{ err error }

func openEditor(filename, block, phase string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	var jump string
	if block != "" {
		jump = "+/^" + block + ":"
	} else if phase != "" {
		jump = "+/^ .*" + phase + ":"
	}

	c := exec.Command(editor, filename, jump) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

type model struct {
	filename string
	block    string
	phase    string
	err      error
}

func (m model) Init() tea.Cmd {
	return openEditor(m.filename, m.block, m.phase)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
	case editorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	return ""
}

func NewModel(filename string, block string, phase string) model {
	return model{filename: filename, block: block, phase: phase}
}
