package tui

import (
	"context"
	"fmt"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/tui/views"
	"github.com/frodi/workshed/internal/workspace"
)

type StackModel struct {
	stack         []views.View
	store         workspace.Store
	ctx           context.Context
	windowSize    measure.Window
	invocationCtx workspace.InvocationContext
}

type StackSnapshot struct {
	Views []ViewSnapshot
}

type ViewSnapshot struct {
	Type string
	Data interface{}
}

func NewStackModel(ctx context.Context, s workspace.Store, invocationCtx workspace.InvocationContext) StackModel {
	dashboard := views.NewDashboardView(ctx, s, invocationCtx)
	return StackModel{
		stack:         []views.View{&dashboard},
		store:         s,
		ctx:           ctx,
		invocationCtx: invocationCtx,
	}
}

func (m StackModel) Init() tea.Cmd {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[0].Init()
}

func (m *StackModel) SetSize(size measure.Window) {
	m.windowSize = size
	for _, v := range m.stack {
		v.SetSize(size)
	}
}

func (m StackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.stack) == 0 {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		size := measure.Window{Width: msg.Width, Height: msg.Height}
		m.SetSize(size)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
			top := m.stack[len(m.stack)-1]
			if top.IsLoading() {
				top.Cancel()
				return m, nil
			}
		}
	}

	top := m.stack[len(m.stack)-1]
	result, cmd := top.Update(msg)

	if result.Action != nil {
		m.handleStackAction(result.Action)
		if len(m.stack) == 0 {
			return m, tea.Quit
		}
		m.stack[len(m.stack)-1].OnResume()
		return m, nil
	}

	if result.NextView != nil {
		m.stack = append(m.stack, result.NextView)
		newView := m.stack[len(m.stack)-1]
		newView.OnPush()
		if m.windowSize.Width > 0 {
			newView.SetSize(m.windowSize)
		}
		initCmd := newView.Init()
		return m, tea.Batch(cmd, initCmd)
	}

	return m, cmd
}

func (m StackModel) View() string {
	if len(m.stack) == 0 {
		return ""
	}

	view := m.stack[len(m.stack)-1].View()

	if m.stack[len(m.stack)-1].IsLoading() {
		spinner := lipgloss.NewStyle().
			Foreground(components.ColorGold).
			Render("◐ Loading...")
		view = view + "\n" + spinner
	}

	return view
}

func (m StackModel) Snapshot() StackSnapshot {
	snapshots := make([]ViewSnapshot, len(m.stack))
	for i, view := range m.stack {
		snapshots[i] = ViewSnapshot{
			Type: reflect.TypeOf(view).String(),
			Data: view.Snapshot(),
		}
	}
	return StackSnapshot{Views: snapshots}
}

func (m StackModel) IsIdle() bool {
	for _, view := range m.stack {
		if view.IsLoading() {
			return false
		}
	}
	return true
}

func (m *StackModel) handleStackAction(action views.StackAction) {
	switch a := action.(type) {
	case views.StackPop:
		if len(m.stack) > 0 {
			m.stack = m.stack[:len(m.stack)-1]
		}
	case views.StackPopUntilType:
		for len(m.stack) > 0 {
			t := reflect.TypeOf(m.stack[len(m.stack)-1])
			if t == a.Type {
				break
			}
			m.stack = m.stack[:len(m.stack)-1]
		}
	case views.StackPopCount:
		for i := 0; i < a.Count && len(m.stack) > 0; i++ {
			m.stack = m.stack[:len(m.stack)-1]
		}
	case views.StackDismissAll:
		m.stack = nil
	}
}

func RunStackModel(ctx context.Context, s workspace.Store, invocationCtx workspace.InvocationContext) error {
	m := NewStackModel(ctx, s, invocationCtx)

	p := tea.NewProgram(m, tea.WithAltScreen())

	ctxCancel, cancel := context.WithCancel(ctx)
	go func() {
		<-ctxCancel.Done()
		p.Quit()
	}()

	finalModel, err := p.Run()
	cancel()
	if err != nil {
		return fmt.Errorf("running stack model: %w", err)
	}

	if fm, ok := finalModel.(StackModel); ok {
		if len(fm.stack) == 0 {
			return nil
		}
	}

	return nil
}
