package forms

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/decisiveai/mdai-cli/internal/types"
)

const maxWidth = 80

var (
	red          = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	blue         = lipgloss.AdaptiveColor{Light: "#96B3DE", Dark: "#96B3DE"}
	yellow       = lipgloss.AdaptiveColor{Light: "#F7B718", Dark: "#F7B718"}
	purple       = lipgloss.AdaptiveColor{Light: "#E288F6", Dark: "#E288F6"}
	grey         = lipgloss.AdaptiveColor{Light: "#6F6F6F", Dark: "#6F6F6F"}
	light_purple = lipgloss.AdaptiveColor{Light: "#D865F2", Dark: "#D865F2"}
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
		Foreground(blue).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(blue).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(yellow).
		Bold(true)
	s.Name = lg.NewStyle().
		Foreground(purple).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(light_purple)
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(grey)
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

var boldStyle = lipgloss.NewStyle().Bold(true)

type TierInfo struct {
	Stores  []string
	Notes   []string
	Formats []string
}

var (
	tiers = map[string]TierInfo{
		"hot": {
			Stores: []string{
				"mdai-hot-s3",
				"my-hot-gcs",
				"quick-one-zone-s3",
				"intelli-s3",
			},
			Notes: []string{boldStyle.Render("Use Case") + ": Frequently accessed data (real-time access).",
				boldStyle.Render("Performance") + ": High-speed access.",
				boldStyle.Render("Cost") + ": Higher cost per GB."},
			Formats: []string{"CSV",
				"JSON",
				"clickhouse",
				"druid",
				"pinot"},
		},
		"cold": {
			Stores: []string{
				"my-cold-gcs-nearline",
				"cloud-cold-gcs",
				"mdai-ia-s3",
				"onezone-cold-s3",
			},
			Notes: []string{
				boldStyle.Render("Use Case") + ": Infrequently accessed data, long-term storage with occasional retrieval.",
				boldStyle.Render("Performance") + ": Slower access compared to hot storage.",
				boldStyle.Render("Cost") + ": Lower cost than hot storage.",
			},
			Formats: []string{"CSV",
				"clickhouse",
				"druid",
				"pinot",
				"iceberg"},
		},
		"glacial": {
			Stores: []string{"archive-gcs",
				"instant-glacier-s3",
				"flex-glacier-s3",
				"deep-glacier-s3"},
			Notes: []string{
				boldStyle.Render("Use Case") + ": Archival data with rare or almost no access, typically for compliance or historical purposes.",
				boldStyle.Render("Performance") + ": Very slow access (hours or days to retrieve).",
				boldStyle.Render("Cost") + ": Extremely low cost per GB, ideal for long-term retention.",
			},
			Formats: []string{"7zip",
				"arc",
				"br",
				"rar",
				"tar",
				"GZIP",
				"zpaq",
				"zst"},
		},
	}
)

func NewModel() Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("tier").
				Options(huh.NewOption("Hot", "hot"),
					huh.NewOption("Cold", "cold"),
					huh.NewOption("Glacial", "glacial")).
				Value(&m.tieredStorage.Tier).
				Title("Choose your storage tier").
				Description("This will determine where you data goes when it's filtered"),
			huh.NewNote().DescriptionFunc(func() string {
				s := strings.Join(tiers[m.tieredStorage.Tier].Notes, "\n")
				return s
			}, &m.tieredStorage.Tier),

			huh.NewSelect[string]().
				Key("store").
				OptionsFunc(func() []huh.Option[string] {
					time.Sleep(500 * time.Millisecond)
					return huh.NewOptions(tiers[m.tieredStorage.Tier].Stores...)
				}, &m.tieredStorage.Tier).
				Value(&m.tieredStorage.Store).
				Title("Choose one of your configured stores"),
		),

		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Title("Name for storage tier").
				Description("Doesn't need to be fancy").
				Placeholder("log_cold_storage").
				Value(&m.tieredStorage.Name).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("name cannot be empty")
					}
					return nil
				}),
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
			huh.NewSelect[string]().
				Key("format").
				OptionsFunc(func() []huh.Option[string] {
					return huh.NewOptions(tiers[m.tieredStorage.Tier].Formats...)
				}, &m.tieredStorage.Tier).
				Value(&m.tieredStorage.Format).
				Title("Choose a file format"),
			huh.NewMultiSelect[string]().
				Key("pipelines").
				Title("Pipelines").
				Description("Ctrl+A to select all").
				Options(
					huh.NewOption("Traces", "traces"),
					huh.NewOption("Metrics", "metrics"),
					huh.NewOption("Logs", "logs").Selected(true),
				).
				Limit(3).
				Value(&m.tieredStorage.Pipelines).
				Validate(func([]string) error {
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
	).WithWidth(45).
		WithHeight(20).
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
		m.tieredStorage.Name = m.form.GetString("name")
		m.tieredStorage.Description = m.form.GetString("description")
		m.tieredStorage.Tier = strings.ToLower(m.form.GetString("tier"))
		m.tieredStorage.Store = m.form.GetString("store")

		m.tieredStorage.Format = m.form.GetString("format")
		m.tieredStorage.Duration = m.form.GetString("duration") + " " + m.form.GetString("duration_type")
		m.tieredStorage.Capacity = m.form.GetString("capacity") + " " + m.form.GetString("capacity_type")

		m.tieredStorage.Pipelines = m.form.Get("pipelines").([]string)

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
		_, _ = fmt.Fprintf(&b, "Fantastic, we'll set up your %s storage tier!\n", name)
		return s.Status.Margin(0, 1).Padding(1, 2).Width(48).Render(b.String()) + "\n"
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
				description        string
				storageParams      string
				tier               string
				store              string
				format             string
				pipelines          string
				capacity           string
				duration           string
			)

			if m.form.GetString("name") != "" {
				name = s.Name.Render(m.form.GetString("name"))
				if m.form.GetString("description") != "" {
					description = "Description: " + m.form.GetString("description") + "\n"
				}
			} else {
				name = s.Name.Render("[STORE NAME TBD]") + "\n"
			}

			tier = "Tier: " + m.form.GetString("tier")
			store = "Store: " + m.form.GetString("store")
			format = "Format: " + m.form.GetString("format")
			if val := m.form.Get("pipelines"); val != nil {
				if pipelinesSlice, ok := val.([]string); ok {
					pipelines = "Pipelines: " + strings.Join(pipelinesSlice, ", ")
				} else {
					pipelines = "Pipelines: "
				}
			} else {
				pipelines = "Pipelines: [None Selected]"
			}
			capacity = "Capacity: " + m.form.GetString("capacity") + m.form.GetString("capacity_type")
			duration = "Duration: " + m.form.GetString("duration") + " " + m.form.GetString("duration_type")

			storageTierInfo = fmt.Sprintf("%s\n%s", tier, store)

			if m.form.GetString("name") != "" {
				storageParamsTitle = "\n\n" + s.StatusHeader.Render("Settings") + "\n"
				storageParams = fmt.Sprintf("%s\n%s\n%s\n%s", format, pipelines, capacity, duration)
			}

			const statusWidth = 28
			statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
			status = s.Status.
				Height(lipgloss.Height(form)).
				Width(statusWidth).
				MarginLeft(statusMarginLeft).
				Render(
					name + "\n" +
						description + "\n" +
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
		lipgloss.WithWhitespaceForeground(blue),
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
