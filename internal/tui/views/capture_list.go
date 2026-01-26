package views

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/key"
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

func (v *CaptureListView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if key.IsCancel(msg) {
		return ViewResult{Action: StackPop{}}, nil
	}

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

	if key.IsDown(msg) {
		if v.selected < len(v.captures)-1 {
			v.selected++
		}
	} else if km, ok := msg.(tea.KeyMsg); ok && len(km.Runes) == 1 {
		if string(km.Runes[0]) == "j" && v.selected < len(v.captures)-1 {
			v.selected++
		}
	}

	if key.IsUp(msg) {
		if v.selected > 0 {
			v.selected--
		}
	} else if km, ok := msg.(tea.KeyMsg); ok && len(km.Runes) == 1 {
		if string(km.Runes[0]) == "k" && v.selected > 0 {
			v.selected--
		}
	}

	if key.IsEnter(msg) && v.selected >= 0 && v.selected < len(v.captures) {
		capture := v.captures[v.selected]
		detailsView := NewCaptureDetailsView(v.store, v.ctx, v.handle, capture.ID)
		return ViewResult{NextView: detailsView}, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok && len(km.Runes) == 1 {
		key := string(km.Runes[0])
		if key == "n" || key == "c" {
			createView := NewCaptureCreateView(v.store, v.ctx, v.handle)
			return ViewResult{NextView: createView}, nil
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
		content += "\n\n" + dimStyle.Render("[n/c] Create capture  [Esc] Back")
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

	content += "\n" + dimStyle.Render("[↑↓/j/k] Navigate  [Enter] Details  [n/c] New  [Esc] Back")

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
