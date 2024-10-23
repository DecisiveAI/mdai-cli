package forms

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/decisiveai/mdai-cli/internal/types"
)

const maxWidth = 80

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	purple = lipgloss.AdaptiveColor{Light: "#940090", Dark: "#ff7bfb"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Name,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Name = lg.NewStyle().
		Foreground(purple).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type Model struct {
	lg            *lipgloss.Renderer
	styles        *Styles
	form          *huh.Form
	width         int
	focusedTier   string
	exiting       bool
	tieredStorage types.TieredStorageOutputAddFlags
}

var stores = map[string][]string{
	"Hot": {
		"AWS S3 Standard",
		"Google Cloud Standard",
	},
	"Cold": {
		"AWS S3 Standard-IA",
		"Google Coldline",
	},
	"Glacial": {
		"AWS Glacial",
		"Google Archive Storage",
	},
}

var tierNotes = map[string][]string{
	"Hot": {
		`Use Case: Frequently accessed data (real-time access).
			Performance: High-speed access.
			Cost: Higher cost per GB.`,
	},
	"Cold": {
		`Use Case: Infrequently accessed data, long-term storage with occasional retrieval
			Performance: Slower access compared to hot storage.
			Cost: Lower cost than hot storage.`,
	},
	"Glacial": {
		`Use Case: Archival data with rare or almost no access, typically for compliance or historical purposes.
			Performance: Very slow access (hours or days to retrieve).
			Cost: Extremely low cost per GB, ideal for long-term retention.`,
	},
}

func NewModel() Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	var store string

	tiers := func() []string {
		keys := slices.Collect(maps.Keys(stores))
		result := make([]string, 0, len(keys))
		for _, key := range keys {
			result = append(result, key)
		}
		return result
	}()

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("tier").
				Options(huh.NewOptions(tiers...)...).
				Value(&m.tieredStorage.Tier).
				Title("Choose your storage tier").
				Description("This will determine where you data goes when it's filtered"),
			huh.NewNote().DescriptionFunc(func() string {
				s := strings.Join(tierNotes[m.tieredStorage.Tier], "\n")
				return s
			}, m.tieredStorage.Tier),
			huh.NewSelect[string]().
				Key("store").
				OptionsFunc(func() []huh.Option[string] {
					s := stores[m.tieredStorage.Tier]
					// time.Sleep(500 * time.Millisecond)
					return huh.NewOptions(s...)
				}, &m.tieredStorage.Tier).
				Value(&store).
				Title("Choose one of your configured stores"),

			huh.NewInput().
				Key("name").
				Title("Name for storage tier").
				Description("Doesn't need to be fancy").
				Placeholder("log_cold_storage").
				Value(&m.tieredStorage.Key).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("name cannot be empty")
					}
					return nil
				}),
		),

		huh.NewGroup(
			huh.NewInput().
				Key("capacity").
				Title("Capacity of storage tier").
				Placeholder("1000").
				Description("How much storage do you want to allot?").
				Validate(func(str string) error {
					if str == "" {
						return errors.New("capacity cannot be empty")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Key("capacity_type").
				Options(huh.NewOptions("bytes", "mb", "gb", "tb")...).
				Title("Choose the capacity type").
				Description("We want to make sure we setup the right amount"),
			huh.NewInput().
				Key("duration").
				Title("Duration of storage tier data kept").
				Placeholder("30").
				Description("This should be an integer").
				Validate(func(str string) error {
					if str == "" {
						return errors.New("capacity cannot be empty")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Key("duration_type").
				Options(huh.NewOptions("minutes", "hours", "days", "months", "years")...).
				Title("Choose the duration type").
				Description("We want to make sure we setup the right time"),
		),

		huh.NewGroup(
			huh.NewInput().
				Key("format").
				Title("Format for storage tier").
				Value(&m.tieredStorage.Format).
				Placeholder("iceberg").
				Validate(func(str string) error {
					if str == "" {
						return errors.New("format cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Key("location").
				Title("Location of storage tier").
				Value(&m.tieredStorage.Location).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("location cannot be empty")
					}
					return nil
				}),
		),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("pipelines").
				Title("Pipelines").
				Options(
					huh.NewOption("traces", "traces"),
					huh.NewOption("metrics", "metrics"),
					huh.NewOption("Logs", "logs").Selected(true),
				).
				Limit(3).
				Value(&m.tieredStorage.Pipelines),
		),

		huh.NewGroup(
			huh.NewInput().
				Key("description").
				Title("Description (optional)").
				Value(&m.tieredStorage.Description).
				Placeholder("This is tiered storage location that will go to S3").
				Validate(func(str string) error {
					return nil
				}),
		),

		huh.NewGroup(
			huh.NewConfirm().
				Key("done").
				Title("Everything look good?").
				Validate(func(v bool) error {
					if !v {
						return fmt.Errorf("go back and fix issues")
					}
					return nil
				}).
				Affirmative("Yes").
				Negative("No"),
		),
	).WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			m.exiting = true
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	form, teacmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f

		if m.form.Get("tier") == "tier" {
			hoveredOption := m.form.GetString("tier")
			if hoveredOption != m.focusedTier {
				m.focusedTier = hoveredOption
			}
		}

		cmds = append(cmds, teacmd)
	}

	if m.form.State == huh.StateCompleted {
		m.tieredStorage.Tier = strings.ToLower(m.form.GetString("tier"))
		m.tieredStorage.RetentionPeriod = m.form.GetString("duration") + " " + m.form.GetString("duration_type")
		m.tieredStorage.Capacity = m.form.GetString("capacity") + " " + m.form.GetString("capacity_type")
		m.tieredStorage.Key = m.form.GetString("name")
		m.tieredStorage.Description = m.form.GetString("description")
		m.tieredStorage.Format = m.form.GetString("format")
		m.tieredStorage.Pipelines = m.form.Get("pipelines").([]string)
		m.tieredStorage.Location = m.form.GetString("location")

		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	s := m.styles

	switch m.form.State {
	case huh.StateCompleted:
		name := m.form.GetString("name")
		tier := m.form.GetString("tier")
		name = s.Highlight.Render(name)
		tier = s.Highlight.Render(tier)
		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "Fantastic, we'll set up your %s storage tier!\n\n", name)
		return s.Status.Margin(0, 1).Padding(1, 2).Width(48).Render(b.String()) + "\n\n"
	default:
		// Form (left side)
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := m.lg.NewStyle().Margin(1, 0).Render(v)

		// Status (Right side)
		var status string
		{
			var (
				storageTierInfo    = "Configurations"
				storageParamsTitle = ""
				name               string
				storageParams      string
				tier               string
				store              string
				capacity           string
				duration           string
			)

			if m.form.GetString("name") != "" {
				name = s.Name.Render(m.form.GetString("name")) + "\n\n"
			} else {
				name = s.Name.Render("Create storage tier") + "\n\n"
			}

			tier = "Tier: " + m.form.GetString("tier")
			store = "Store: " + m.form.GetString("store")
			capacity = "Capacity: " + m.form.GetString("capacity") + m.form.GetString("capacity_type")
			duration = "Duration: " + m.form.GetString("duration") + " " + m.form.GetString("duration_type")

			storageTierInfo = fmt.Sprintf("%s\n%s", tier, store)

			if m.form.GetString("name") != "" {
				storageParamsTitle = "\n\n" + s.StatusHeader.Render("Capacity & Duration") + "\n"
				storageParams = fmt.Sprintf("%s\n%s", capacity, duration)
			}

			const statusWidth = 28
			statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
			status = s.Status.
				Height(lipgloss.Height(form)).
				Width(statusWidth).
				MarginLeft(statusMarginLeft).
				Render(
					name +
						s.StatusHeader.Render("Configurations") + "\n" +
						storageTierInfo +
						storageParamsTitle +
						storageParams,
				)
		}
		err := m.form.Errors()
		header := m.appBoundaryView("Storage Tier Configuration")
		if len(err) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}
		body := lipgloss.JoinHorizontal(lipgloss.Top, form, status)

		footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		if len(err) > 0 {
			footer = m.appErrorBoundaryView("")
		}

		return s.Base.Render(header + "\n" + body + "\n\n" + footer)
	}
}

func (m Model) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m Model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m Model) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}

func TieredStorageForm() (bool, types.TieredStorageOutputAddFlags) {
	m, err := tea.NewProgram(NewModel()).Run()
	if err != nil {
		fmt.Println("Unable to create storage tier due to", err)
		os.Exit(1)
	}
	return !m.(Model).exiting, m.(Model).tieredStorage
}
