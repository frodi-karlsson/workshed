package views

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type CaptureDetailsView struct {
	store     workspace.Store
	ctx       context.Context
	handle    string
	captureID string
	capture   *workspace.Capture
	preflight workspace.ApplyPreflightResult
	confirm   bool
	done      bool
	loading   bool
	size      measure.Window
}

func NewCaptureDetailsView(s workspace.Store, ctx context.Context, handle, captureID string) *CaptureDetailsView {
	capture, _ := s.GetCapture(ctx, handle, captureID)
	preflight, _ := s.PreflightApply(ctx, handle, captureID)
	return &CaptureDetailsView{
		store:     s,
		ctx:       ctx,
		handle:    handle,
		captureID: captureID,
		capture:   capture,
		preflight: preflight,
	}
}

func (v *CaptureDetailsView) Init() tea.Cmd { return nil }

func (v *CaptureDetailsView) SetSize(size measure.Window) {
	v.size = size
}

func (v *CaptureDetailsView) OnPush()   {}
func (v *CaptureDetailsView) OnResume() {}
func (v *CaptureDetailsView) IsLoading() bool {
	return v.loading
}

func (v *CaptureDetailsView) Cancel() {
	v.loading = false
}

func (v *CaptureDetailsView) KeyBindings() []KeyBinding {
	if v.confirm {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Apply", Action: v.apply},
			{Key: "esc", Help: "[Esc] Cancel", Action: v.cancel},
		}
	}
	if !v.preflight.Valid {
		return GetDismissKeyBindings(v.cancel, "Back")
	}
	return append(
		[]KeyBinding{{Key: "enter", Help: "[Enter] Apply", Action: v.confirmApply}},
		GetDismissKeyBindings(v.cancel, "Back")...,
	)
}

func (v *CaptureDetailsView) confirmApply() (ViewResult, tea.Cmd) {
	if !v.preflight.Valid {
		return ViewResult{}, nil
	}
	v.confirm = true
	return ViewResult{}, nil
}

func (v *CaptureDetailsView) apply() (ViewResult, tea.Cmd) {
	v.loading = true
	go func() {
		err := v.store.ApplyCapture(v.ctx, v.handle, v.captureID)
		v.loading = false
		if err == nil {
			v.done = true
		}
	}()
	return ViewResult{}, nil
}

func (v *CaptureDetailsView) cancel() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *CaptureDetailsView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	return ViewResult{}, nil
}

func (v *CaptureDetailsView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)
	dimStyle := lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	errorStyle := lipgloss.NewStyle().Foreground(components.ColorError)
	warningStyle := lipgloss.NewStyle().Foreground(components.ColorWarning)
	successStyle := lipgloss.NewStyle().Foreground(components.ColorSuccess)

	if v.loading && !v.done {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Applying capture..."),
				"",
				subStyle.Render("Please wait"),
			),
		)
	}

	if v.done {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				successStyle.Render("Capture Applied Successfully"),
				"",
				subStyle.Render("All repositories have been restored to the captured state."),
				"",
				dimStyle.Render("[Enter/Esc] Dismiss"),
			),
		)
	}

	if v.capture == nil {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Capture Details"),
				"",
				subStyle.Render("Loading..."),
			),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Capture Details"),
		"",
	)

	name := v.capture.Name
	if name == "" {
		name = "(unnamed)"
	}
	content += subStyle.Render("Name: ") + name + "\n"
	content += subStyle.Render("Created: ") + v.capture.Timestamp.Format("Jan 02 15:04") + "\n"

	if v.capture.Metadata.Description != "" {
		content += subStyle.Render("Description: ") + v.capture.Metadata.Description + "\n"
	}

	content += "\n" + subStyle.Render("Repositories:")
	content += "\n"

	for _, ref := range v.capture.GitState {
		status := "clean"
		statusStyle := successStyle
		if ref.Dirty {
			status = "dirty"
			statusStyle = warningStyle
		}

		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			"  "+ref.Repository,
			subStyle.Render(" ("+ref.Branch+")"),
			subStyle.Render(" "+ref.Commit[:8]),
			statusStyle.Render(" ["+status+"]"),
		)
		content += line + "\n"
	}

	if v.confirm {
		content += "\n" + warningStyle.Render("Apply this capture?")
		content += "\n" + subStyle.Render("This will checkout each repository to the captured commit.")
		content += "\n\n" + subStyle.Render("[Enter] Confirm  [Esc] Cancel")
	} else if !v.preflight.Valid {
		content += "\n" + errorStyle.Render("Cannot Apply")
		content += "\n\n"

		for _, err := range v.preflight.Errors {
			content += errorStyle.Render("× " + err.Repository + ": " + err.Reason)
			if err.Details != "" {
				content += " (" + err.Details + ")"
			}
			content += "\n"
		}

		content += "\n" + dimStyle.Render("[Esc] Back")
	} else {
		content += "\n" + successStyle.Render("Ready to apply")
		content += "\n"
		helpText := GenerateHelp(v.KeyBindings())
		content += "\n" + dimStyle.Render(helpText)
	}

	return ModalFrame(v.size).Render(content)
}

type CaptureDetailsViewSnapshot struct {
	Type         string
	Handle       string
	CaptureID    string
	CaptureName  string
	RepoCount    int
	PreflightOK  bool
	ErrorCount   int
	Confirm      bool
	ApplySuccess bool
}

func (v *CaptureDetailsView) Snapshot() interface{} {
	name := ""
	if v.capture != nil {
		name = v.capture.Name
	}
	return CaptureDetailsViewSnapshot{
		Type:         "CaptureDetailsView",
		Handle:       v.handle,
		CaptureID:    v.captureID,
		CaptureName:  name,
		RepoCount:    len(v.preflight.Errors),
		PreflightOK:  v.preflight.Valid,
		ErrorCount:   len(v.preflight.Errors),
		Confirm:      v.confirm,
		ApplySuccess: v.done,
	}
}
