package views

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type ApplyCaptureView struct {
	store    workspace.Store
	ctx      context.Context
	handle   string
	list     list.Model
	captures []workspace.Capture
	size     measure.Window
}

type ApplyCaptureItem struct {
	capture workspace.Capture
}

func (i ApplyCaptureItem) Title() string {
	if i.capture.Name != "" {
		return i.capture.Name
	}
	return fmt.Sprintf("Capture %s", i.capture.ID[:8])
}

func (i ApplyCaptureItem) Description() string {
	return fmt.Sprintf("%s • %d repos", i.capture.Timestamp.Format("Jan 2 15:04"), len(i.capture.GitState))
}

func (i ApplyCaptureItem) FilterValue() string {
	if i.capture.Name != "" {
		return i.capture.Name
	}
	return i.capture.ID
}

func NewApplyCaptureView(s workspace.Store, ctx context.Context, handle string) ApplyCaptureView {
	captures, _ := s.ListCaptures(ctx, handle)

	items := make([]list.Item, len(captures))
	for i, cap := range captures {
		items[i] = ApplyCaptureItem{capture: cap}
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, MaxListHeight)
	l.Title = "Apply Capture"
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Padding(0, 1)

	return ApplyCaptureView{
		store:    s,
		ctx:      ctx,
		handle:   handle,
		list:     l,
		captures: captures,
	}
}

func (v *ApplyCaptureView) OnPush() {}

func (v *ApplyCaptureView) OnResume() {}

func (v *ApplyCaptureView) IsLoading() bool { return false }

func (v *ApplyCaptureView) Cancel() {}

func (v *ApplyCaptureView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *ApplyCaptureView) Init() tea.Cmd { return nil }

func (v *ApplyCaptureView) KeyBindings() []KeyBinding {
	return append(
		[]KeyBinding{{Key: "enter", Help: "[Enter] Apply", Action: v.applyCapture}},
		GetDismissKeyBindings(v.goBack, "Back")...,
	)
}

func (v *ApplyCaptureView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ApplyCaptureView) applyCapture() (ViewResult, tea.Cmd) {
	selected := v.list.SelectedItem()
	if item, ok := selected.(ApplyCaptureItem); ok {
		if err := v.store.ApplyCapture(v.ctx, v.handle, item.capture.ID); err != nil {
			return ViewResult{Action: StackPop{}}, nil
		}
		return ViewResult{Action: StackPopCount{Count: 2}}, nil
	}
	return ViewResult{}, nil
}

func (v *ApplyCaptureView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	return ViewResult{}, cmd
}

func (v *ApplyCaptureView) View() string {
	helpText := GenerateHelp(v.KeyBindings())
	helpHint := lipgloss.NewStyle().
		Foreground(components.ColorMuted).
		MarginTop(1).
		Render(helpText)

	frameStyle := ModalFrame(v.size)
	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			v.list.View(),
			helpHint,
		),
	)
}

type ApplyCaptureViewSnapshot struct {
	Type      string
	Handle    string
	CaptureCt int
}

func (v *ApplyCaptureView) Snapshot() interface{} {
	return ApplyCaptureViewSnapshot{
		Type:      "ApplyCaptureView",
		Handle:    v.handle,
		CaptureCt: len(v.captures),
	}
}
