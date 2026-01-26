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

type modal_ApplyView struct {
	store        workspace.Store
	ctx          context.Context
	handle       string
	capture      *workspace.Capture
	preflight    workspace.ApplyPreflightResult
	confirm      bool
	applySuccess bool
	loading      bool
	size         measure.Window
}

func NewApplyView(s workspace.Store, ctx context.Context, handle, captureID string) *modal_ApplyView {
	capture, _ := s.GetCapture(ctx, handle, captureID)
	preflight, _ := s.PreflightApply(ctx, handle, captureID)
	return &modal_ApplyView{
		store:     s,
		ctx:       ctx,
		handle:    handle,
		capture:   capture,
		preflight: preflight,
	}
}

func (v *modal_ApplyView) Init() tea.Cmd { return nil }

func (v *modal_ApplyView) SetSize(size measure.Window) {
	v.size = size
}

func (v *modal_ApplyView) OnPush()   {}
func (v *modal_ApplyView) OnResume() {}
func (v *modal_ApplyView) IsLoading() bool {
	return v.loading
}

func (v *modal_ApplyView) Cancel() {
	v.loading = false
}

func (v *modal_ApplyView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if key.IsCancel(msg) {
		return ViewResult{Action: StackPop{}}, nil
	}

	if v.loading {
		return ViewResult{}, nil
	}

	if v.applySuccess {
		if key.IsEnter(msg) || key.IsCancel(msg) {
			return ViewResult{Action: StackPop{}}, nil
		}
		return ViewResult{}, nil
	}

	if key.IsEnter(msg) {
		if v.confirm {
			v.loading = true
			go func() {
				err := v.store.ApplyCapture(v.ctx, v.handle, v.capture.ID)
				v.loading = false
				if err == nil {
					v.applySuccess = true
				}
			}()
		} else {
			v.confirm = true
		}
		return ViewResult{}, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyBackspace {
		v.confirm = false
	}

	return ViewResult{}, nil
}

func (v *modal_ApplyView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)
	errorStyle := lipgloss.NewStyle().Foreground(components.ColorError)
	warningStyle := lipgloss.NewStyle().Foreground(components.ColorWarning)
	successStyle := lipgloss.NewStyle().Foreground(components.ColorSuccess)

	if v.loading {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Applying capture..."),
				subStyle.Render("Please wait"),
			),
		)
	}

	if v.applySuccess {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Apply Successful"),
				"",
				subStyle.Render("Capture "+v.capture.Name+" has been applied."),
				"",
				subStyle.Render("[Enter/Esc] Dismiss"),
			),
		)
	}

	if v.capture == nil {
		return ModalFrame(v.size).Render("Loading...")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Apply Capture: "+v.capture.Name),
		"",
	)

	if !v.preflight.Valid {
		content += lipgloss.JoinVertical(
			lipgloss.Left,
			errorStyle.Render("Cannot apply - issues detected"),
			"",
		)
		for _, err := range v.preflight.Errors {
			content += errorStyle.Render("  - " + err.Repository + ": " + err.Reason)
			if err.Details != "" {
				content += "\n    " + subStyle.Render(err.Details)
			}
			content += "\n"
		}
		content += "\n" + subStyle.Render("Fix these issues before applying.")
	} else {
		content += subStyle.Render("This will apply the captured git state to your workspace.")
		content += "\n\n"

		var repoLines []string
		for _, ref := range v.capture.GitState {
			status := "clean"
			if ref.Dirty {
				status = "dirty"
			}
			repoLines = append(repoLines, lipgloss.JoinHorizontal(
				lipgloss.Left,
				"  "+ref.Repository,
				subStyle.Render(" ("+ref.Branch+" @ "+ref.Commit[0:7]+")"),
				lipgloss.NewStyle().Render(" ["+status+"]"),
			))
		}
		content += subStyle.Render("Repositories:")
		content += "\n" + lipgloss.JoinVertical(lipgloss.Left, repoLines...)

		if !v.confirm {
			content += "\n\n" + subStyle.Render("No destructive actions will be performed.")
			content += "\n" + warningStyle.Render("[Enter] Proceed  [Esc] Cancel")
		} else {
			content += "\n\n" + successStyle.Render("Ready to apply. Confirm?")
			content += "\n" + subStyle.Render("[Enter] Confirm  [Backspace] Go back  [Esc] Cancel")
		}
	}

	return ModalFrame(v.size).Render(content)
}

type ApplyViewSnapshot struct {
	Type         string
	Handle       string
	CaptureID    string
	CaptureName  string
	Preflight    ApplyPreflightSnapshot
	Confirm      bool
	ApplySuccess bool
	RepoCount    int
}

type ApplyPreflightSnapshot struct {
	Valid  bool
	Errors []PreflightErrorSnapshot
}

type PreflightErrorSnapshot struct {
	Repository string
	Reason     string
	Details    string
}

func (v *modal_ApplyView) Snapshot() interface{} {
	repoCount := 0
	captureName := ""
	if v.capture != nil {
		repoCount = len(v.capture.GitState)
		captureName = v.capture.Name
	}

	preflight := ApplyPreflightSnapshot{
		Valid:  v.preflight.Valid,
		Errors: []PreflightErrorSnapshot{},
	}
	for _, err := range v.preflight.Errors {
		preflight.Errors = append(preflight.Errors, PreflightErrorSnapshot{
			Repository: err.Repository,
			Reason:     err.Reason,
			Details:    err.Details,
		})
	}

	return ApplyViewSnapshot{
		Type:         "ApplyView",
		Handle:       v.handle,
		CaptureID:    "",
		CaptureName:  captureName,
		Preflight:    preflight,
		Confirm:      v.confirm,
		ApplySuccess: v.applySuccess,
		RepoCount:    repoCount,
	}
}
