package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Scrollable struct {
	vp          viewport.Model
	label       string
	width       int
	height      int
	border      lipgloss.Border
	borderColor lipgloss.Color
}

func NewScrollable(width, height int) Scrollable {
	vp := viewport.New(width, height)
	vp.KeyMap = viewport.KeyMap{}
	return Scrollable{
		vp:          vp,
		width:       width,
		height:      height,
		border:      lipgloss.RoundedBorder(),
		borderColor: ColorBorder,
	}
}

func (s *Scrollable) SetLabel(label string) {
	s.label = label
}

func (s *Scrollable) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.vp.Width = width
	s.vp.Height = height
}

func (s *Scrollable) SetContent(content string) {
	s.vp.SetContent(content)
}

func (s *Scrollable) Update(msg tea.Msg) (changed bool, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			s.vp.LineUp(1)
			return true, nil
		case "down", "j":
			s.vp.LineDown(1)
			return true, nil
		case "pgup":
			s.vp.ViewUp()
			return true, nil
		case "pgdown":
			s.vp.ViewDown()
			return true, nil
		case "home":
			s.vp.GotoTop()
			return true, nil
		case "end":
			s.vp.GotoBottom()
			return true, nil
		}
	}
	return false, nil
}

func (s *Scrollable) View() string {
	borderStyle := lipgloss.NewStyle().
		Border(s.border).
		BorderForeground(s.borderColor).
		Padding(1).
		Width(s.width).
		Height(s.height)

	var content string
	if s.label != "" {
		labelStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Padding(0, 1)
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			labelStyle.Render(s.label),
			"\n",
			s.vp.View(),
		)
	} else {
		content = s.vp.View()
	}

	return borderStyle.Render(content)
}

func (s *Scrollable) AtTop() bool {
	return s.vp.AtTop()
}

func (s *Scrollable) AtBottom() bool {
	return s.vp.AtBottom()
}

func (s *Scrollable) ScrollToTop() {
	s.vp.GotoTop()
}

func (s *Scrollable) ScrollToBottom() {
	s.vp.GotoBottom()
}

func (s *Scrollable) LineUp() {
	s.vp.LineUp(1)
}

func (s *Scrollable) LineDown() {
	s.vp.LineDown(1)
}

func (s Scrollable) Width() int {
	return s.width
}

func (s Scrollable) Height() int {
	return s.height
}
