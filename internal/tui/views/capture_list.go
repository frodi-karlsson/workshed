package views

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type CaptureListView struct {
	store    workspace.Store
	ctx      context.Context
	handle   string
	captures []workspace.Capture
	loading  bool
	selected int
	size     measure.Window
}

func NewCaptureListView(s workspace.Store, ctx context.Context, handle string) *CaptureListView {
	return &CaptureListView{
		store:    s,
		ctx:      ctx,
		handle:   handle,
		selected: -1,
	}
}

func (v *CaptureListView) Init() tea.Cmd { return nil }

func (v *CaptureListView) SetSize(size measure.Window) {
	v.size = size
}

func (v *CaptureListView) OnPush() {
	captures, _ := v.store.ListCaptures(v.ctx, v.handle)
	v.captures = captures
	if len(captures) > 0 {
		v.selected = 0
	}
}
func (v *CaptureListView) OnResume() {}
func (v *CaptureListView) IsLoading() bool {
	return v.loading
}

func (v *CaptureListView) Cancel() {
	v.loading = false
}

func (v *CaptureListView) KeyBindings() []KeyBinding {
	return append(
		[]KeyBinding{
			{Key: "up", Help: "[↑] Navigate", Action: v.moveUp, When: v.hasCaptures},
			{Key: "down", Help: "[↓] Navigate", Action: v.moveDown, When: v.hasCaptures},
			{Key: "enter", Help: "[Enter] Details", Action: v.openDetails, When: v.hasSelection},
			{Key: "n", Help: "[n] New", Action: v.createNew},
			{Key: "c", Help: "[c] New", Action: v.createNew},
		},
		GetDismissKeyBindings(v.goBack, "Back")...,
	)
}

func (v *CaptureListView) hasCaptures() bool {
	return len(v.captures) > 0
}

func (v *CaptureListView) hasSelection() bool {
	return v.selected >= 0 && v.selected < len(v.captures)
}

func (v *CaptureListView) moveUp() (ViewResult, tea.Cmd) {
	if v.selected > 0 {
		v.selected--
	}
	return ViewResult{}, nil
}

func (v *CaptureListView) moveDown() (ViewResult, tea.Cmd) {
	if v.selected < len(v.captures)-1 {
		v.selected++
	}
	return ViewResult{}, nil
}

func (v *CaptureListView) openDetails() (ViewResult, tea.Cmd) {
	capture := v.captures[v.selected]
	detailsView := NewCaptureDetailsView(v.store, v.ctx, v.handle, capture.ID)
	return ViewResult{NextView: detailsView}, nil
}

func (v *CaptureListView) createNew() (ViewResult, tea.Cmd) {
	createView := NewCaptureCreateView(v.store, v.ctx, v.handle)
	return ViewResult{NextView: createView}, nil
}

func (v *CaptureListView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *CaptureListView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if v.loading {
		return ViewResult{}, nil
	}
	if len(v.captures) == 0 {
		captures, err := v.store.ListCaptures(v.ctx, v.handle)
		if err == nil {
			v.captures = captures
			if len(captures) > 0 && v.selected == -1 {
				v.selected = 0
			}
		}
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	return ViewResult{}, nil
}

func (v *CaptureListView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)
	dimStyle := lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	selectedStyle := lipgloss.NewStyle().Foreground(components.ColorHighlight).Bold(true)

	if v.loading && len(v.captures) == 0 {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Captures"),
				"",
				subStyle.Render("Loading..."),
			),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Captures"),
		"",
	)

	if len(v.captures) == 0 {
		content += subStyle.Render("No captures yet")
		content += "\n\n" + dimStyle.Render(GenerateHelp(v.KeyBindings()))
		return ModalFrame(v.size).Render(content)
	}

	content += subStyle.Render(formatCaptureCount(len(v.captures)))
	content += "\n\n"

	for i, cap := range v.captures {
		isSelected := i == v.selected

		var lineStyle lipgloss.Style
		if isSelected {
			lineStyle = selectedStyle
		} else {
			lineStyle = subStyle
		}

		timestamp := cap.Timestamp.Format("Jan 02 15:04")
		name := cap.Name
		if name == "" {
			name = "(unnamed)"
		}

		line := lineStyle.Render("○ ")
		if isSelected {
			line += " "
		} else {
			line += "  "
		}
		line += timestamp + "  " + name

		repoCount := len(cap.GitState)
		dirtyCount := 0
		for _, ref := range cap.GitState {
			if ref.Dirty {
				dirtyCount++
			}
		}

		if isSelected {
			details := lipgloss.NewStyle().Foreground(components.ColorMuted).Render(
				"     " + formatInt(repoCount) + " repos",
			)
			if dirtyCount > 0 {
				details += dimStyle.Render(" (" + formatInt(dirtyCount) + " dirty)")
			}
			line += "\n" + details
		}

		content += line + "\n"
	}

	content += "\n" + dimStyle.Render(GenerateHelp(v.KeyBindings()))

	return ModalFrame(v.size).Render(content)
}

type CaptureListViewSnapshot struct {
	Type       string
	Handle     string
	CaptureCnt int
	Selected   int
}

func (v *CaptureListView) Snapshot() interface{} {
	return CaptureListViewSnapshot{
		Type:       "CaptureListView",
		Handle:     v.handle,
		CaptureCnt: len(v.captures),
		Selected:   v.selected,
	}
}

func formatCaptureCount(n int) string {
	if n == 1 {
		return "1 capture"
	}
	return formatInt(n) + " captures"
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
