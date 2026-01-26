package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ColorBorder     = lipgloss.Color("#874B07")
	ColorSuccess    = lipgloss.Color("#4CD964")
	ColorError      = lipgloss.Color("#FF6B6B")
	ColorText       = lipgloss.Color("#D4D4D4")
	ColorMuted      = lipgloss.Color("#888888")
	ColorVeryMuted  = lipgloss.Color("#666666")
	ColorBackground = lipgloss.Color("#3C3C3C")
	ColorHighlight  = lipgloss.Color("#5C5CFF")
	ColorWarning    = lipgloss.Color("#FFB347")
	ColorGold       = lipgloss.Color("#FFD700")
)

type StyleSet struct {
	Frame        lipgloss.Style
	Title        lipgloss.Style
	Header       lipgloss.Style
	Text         lipgloss.Style
	Muted        lipgloss.Style
	VeryMuted    lipgloss.Style
	Success      lipgloss.Style
	Error        lipgloss.Style
	Warning      lipgloss.Style
	Highlight    lipgloss.Style
	Input        lipgloss.Style
	Button       lipgloss.Style
	ButtonActive lipgloss.Style
	ListItem     lipgloss.Style
	Selected     lipgloss.Style
	Focused      lipgloss.Style
	Help         lipgloss.Style
}

func NewStyleSet() StyleSet {
	s := StyleSet{}

	s.Frame = lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1)

	s.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		Background(ColorBackground).
		Padding(0, 1)

	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText)

	s.Text = lipgloss.NewStyle().
		Foreground(ColorText)

	s.Muted = lipgloss.NewStyle().
		Foreground(ColorMuted)

	s.VeryMuted = lipgloss.NewStyle().
		Foreground(ColorVeryMuted)

	s.Success = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	s.Error = lipgloss.NewStyle().
		Foreground(ColorError)

	s.Warning = lipgloss.NewStyle().
		Foreground(ColorWarning)

	s.Highlight = lipgloss.NewStyle().
		Foreground(ColorHighlight)

	s.Input = lipgloss.NewStyle().
		Foreground(ColorText)

	s.Button = lipgloss.NewStyle().
		Foreground(ColorText).
		Padding(0, 1)

	s.ButtonActive = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true).
		Padding(0, 1)

	s.ListItem = lipgloss.NewStyle().
		Foreground(ColorText)

	s.Selected = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	s.Focused = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Underline(true)

	s.Help = lipgloss.NewStyle().
		Foreground(ColorMuted).
		MarginTop(1)

	return s
}

type Modal struct {
	Style       lipgloss.Style
	BorderColor lipgloss.Color
	Width       int
	Height      int
}

func NewModal() Modal {
	m := Modal{}
	m.Style = lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1)
	m.BorderColor = ColorBorder
	return m
}

func (m *Modal) SetBorderColor(color lipgloss.Color) *Modal {
	m.BorderColor = color
	m.Style = m.Style.BorderForeground(color)
	return m
}

func (m *Modal) SetSize(width, height int) *Modal {
	m.Width = width
	m.Height = height
	if width > 0 {
		m.Style = m.Style.Width(width)
	}
	if height > 0 {
		m.Style = m.Style.Height(height)
	}
	return m
}

func (m Modal) Render(content string) string {
	return m.Style.Render(content)
}

type InputField struct {
	Style            lipgloss.Style
	FocusStyle       lipgloss.Style
	PlaceholderStyle lipgloss.Style
	Focused          bool
}

func NewInputField() InputField {
	return InputField{
		Style: lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1),
		FocusStyle: lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorHighlight),
		PlaceholderStyle: lipgloss.NewStyle().
			Foreground(ColorVeryMuted).
			Padding(0, 1),
	}
}

func (i InputField) Render(value, placeholder string) string {
	if i.Focused {
		if value == "" {
			return i.FocusStyle.Render(placeholder)
		}
		return i.FocusStyle.Render(value)
	}
	if value == "" {
		return i.Style.Render(placeholder)
	}
	return i.Style.Render(value)
}

func (i *InputField) SetFocused(focused bool) *InputField {
	i.Focused = focused
	return i
}

type ListItem struct {
	Style         lipgloss.Style
	SelectedStyle lipgloss.Style
	Selected      bool
}

func NewListItem() ListItem {
	return ListItem{
		Style: lipgloss.NewStyle().
			Foreground(ColorText),
		SelectedStyle: lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true),
	}
}

func (i ListItem) Render(title string) string {
	if i.Selected {
		return i.SelectedStyle.Render(title)
	}
	return i.Style.Render(title)
}

func (i *ListItem) SetSelected(selected bool) *ListItem {
	i.Selected = selected
	return i
}

type Button struct {
	Style       lipgloss.Style
	ActiveStyle lipgloss.Style
	Label       string
	Active      bool
}

func NewButton(label string) Button {
	return Button{
		Style: lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 2).
			Margin(0, 1),
		ActiveStyle: lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true).
			Padding(0, 2).
			Margin(0, 1).
			Underline(true),
		Label: label,
	}
}

func (b Button) Render() string {
	if b.Active {
		return b.ActiveStyle.Render("[" + b.Label + "]")
	}
	return b.Style.Render("[" + b.Label + "]")
}

func (b *Button) SetActive(active bool) *Button {
	b.Active = active
	return b
}

type ProgressBar struct {
	Style     lipgloss.Style
	FillStyle lipgloss.Style
	Current   int
	Total     int
	Width     int
}

func NewProgressBar(current, total, width int) ProgressBar {
	return ProgressBar{
		Style: lipgloss.NewStyle().
			Foreground(ColorVeryMuted),
		FillStyle: lipgloss.NewStyle().
			Foreground(ColorSuccess),
		Current: current,
		Total:   total,
		Width:   width,
	}
}

func (p ProgressBar) Render() string {
	if p.Total == 0 {
		return p.Style.Render("[" + lipgloss.NewStyle().Width(p.Width).Render("") + "]")
	}

	fillWidth := (p.Current * p.Width) / p.Total
	emptyWidth := p.Width - fillWidth

	fill := p.FillStyle.Render(lipgloss.NewStyle().Width(fillWidth).Render("#"))
	empty := p.Style.Render(lipgloss.NewStyle().Width(emptyWidth).Render("-"))

	return p.Style.Render("[" + fill + empty + "]")
}

type HelpItem struct {
	Key   string
	Label string
}

func RenderHelp(items []HelpItem) string {
	if len(items) == 0 {
		return ""
	}
	var parts []string
	for _, item := range items {
		if item.Label != "" {
			parts = append(parts, "["+item.Key+"] "+item.Label)
		} else {
			parts = append(parts, "["+item.Key+"]")
		}
	}
	return lipgloss.NewStyle().Foreground(ColorVeryMuted).Render(strings.Join(parts, "  "))
}

var (
	HelpDismiss   = RenderHelp([]HelpItem{{"Esc/q/Enter", "Dismiss"}})
	HelpNavigate  = RenderHelp([]HelpItem{{"↑↓/j/k", "Navigate"}, {"Enter", "Select"}, {"Esc", "Cancel"}})
	HelpInput     = RenderHelp([]HelpItem{{"Enter", "Confirm"}, {"Tab", "Browse"}, {"Esc", "Cancel"}})
	HelpDashboard = RenderHelp([]HelpItem{{"c", "Create"}, {"Enter", "Menu"}, {"l", "Filter"}, {"?", "Help"}, {"q", "Quit"}})
)
