package views

import (
	"context"
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/fs"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type ExportView struct {
	store      workspace.Store
	ctx        context.Context
	handle     string
	jsonData   string
	scrollable components.Scrollable
	size       measure.Window
	copyResult string
	loadErr    error
}

func NewExportView(s workspace.Store, ctx context.Context, handle string, opts ...func(*ExportView)) ExportView {
	contextData, err := s.ExportContext(ctx, handle)
	if err != nil {
		contextData = &workspace.WorkspaceContext{
			Handle:  handle,
			Purpose: "Error loading workspace",
		}
	}

	data, _ := json.MarshalIndent(contextData, "", "  ")
	jsonData := string(data)

	scrollable := components.NewScrollable(60, 8)

	v := ExportView{
		store:      s,
		ctx:        ctx,
		handle:     handle,
		jsonData:   jsonData,
		scrollable: scrollable,
		loadErr:    err,
	}

	for _, opt := range opts {
		opt(&v)
	}

	return v
}

func (v *ExportView) OnPush() {}

func (v *ExportView) OnResume() {}

func (v *ExportView) IsLoading() bool { return false }

func (v *ExportView) Cancel() {}

func (v *ExportView) SetSize(size measure.Window) {
	v.size = size
	v.scrollable.SetSize(size.ModalWidth()-4, size.ModalHeight()-14)
	v.scrollable.SetContent(v.jsonData)
}

func (v *ExportView) Init() tea.Cmd { return nil }

func (v *ExportView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "c", Help: "[c] Copy", Action: v.copyToClipboard},
		{Key: "s", Help: "[s] Save", Action: v.saveToFile},
		{Key: "esc", Help: "[Esc] Back", Action: v.goBack},
		{Key: "ctrl+c", Help: "[Ctrl+C] Back", Action: v.goBack},
	}
}

func (v *ExportView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ExportView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	v.scrollable.Update(msg)
	return ViewResult{}, nil
}

func (v *ExportView) copyToClipboard() (ViewResult, tea.Cmd) {
	err := v.store.GetClipboard().WriteAll(v.jsonData)
	if err == nil {
		v.copyResult = "Copied to clipboard!"
	} else {
		v.copyResult = "Failed to copy"
	}
	return ViewResult{}, nil
}

func (v *ExportView) saveToFile() (ViewResult, tea.Cmd) {
	jsonData := v.jsonData
	pathInput := NewPathInputView("Save path", ">", func(path string) {
		err := fs.WriteJson(path, []byte(jsonData))
		if err == nil {
			v.copyResult = "Saved to " + path
		} else {
			v.copyResult = "Failed: " + err.Error()
		}
	}, func() {})
	pathInputView := &pathInput
	return ViewResult{NextView: pathInputView}, nil
}

func (v *ExportView) View() string {
	helpItems := []components.HelpItem{
		{Key: "c", Label: "Copy"},
		{Key: "s", Label: "Save"},
		{Key: "↑↓", Label: "Scroll JSON"},
		{Key: "Esc", Label: "Back"},
	}
	helpText := components.RenderHelp(helpItems)

	var statusMsg string
	if v.loadErr != nil {
		statusMsg = lipgloss.NewStyle().
			Foreground(components.ColorError).
			Render("Error: " + v.loadErr.Error())
	} else if v.copyResult != "" {
		if contains(v.copyResult, "Copied") {
			statusMsg = lipgloss.NewStyle().
				Foreground(components.ColorSuccess).
				Render(v.copyResult)
		} else {
			statusMsg = lipgloss.NewStyle().
				Foreground(components.ColorError).
				Render(v.copyResult)
		}
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Padding(0, 1).
		Render("Export: " + v.handle)

	helpHint := lipgloss.NewStyle().
		Foreground(components.ColorMuted).
		MarginTop(1).
		Render(helpText)

	frameStyle := ModalFrame(v.size).Width(v.size.ModalWidth())
	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"\n",
			v.scrollable.View(),
			statusMsg,
			helpHint,
		),
	)
}

type ExportViewSnapshot struct {
	Type          string
	Handle        string
	HasCopyResult bool
	CopyResult    string
}

func (v *ExportView) Snapshot() interface{} {
	return ExportViewSnapshot{
		Type:          "ExportView",
		Handle:        v.handle,
		HasCopyResult: v.copyResult != "",
		CopyResult:    v.copyResult,
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
