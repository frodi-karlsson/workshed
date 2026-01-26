package views

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type view_ValidateView struct {
	store   workspace.Store
	ctx     context.Context
	handle  string
	result  *workspace.AgentsValidationResult
	errMsg  string
	loading bool
	size    measure.Window
}

func NewValidateView(s workspace.Store, ctx context.Context, handle string) *view_ValidateView {
	return &view_ValidateView{
		store:  s,
		ctx:    ctx,
		handle: handle,
	}
}

type validationResultMsg struct {
	result *workspace.AgentsValidationResult
}

type validationErrorMsg struct {
	err string
}

func (v *view_ValidateView) Init() tea.Cmd {
	v.loading = true
	return func() tea.Msg {
		result, err := v.store.ValidateAgents(v.ctx, v.handle, "AGENTS.md")
		if err == nil {
			return validationResultMsg{result: &result}
		}
		return validationErrorMsg{err: err.Error()}
	}
}

func (v *view_ValidateView) SetSize(size measure.Window) {
	v.size = size
}

func (v *view_ValidateView) OnPush()   {}
func (v *view_ValidateView) OnResume() {}
func (v *view_ValidateView) IsLoading() bool {
	return v.loading
}

func (v *view_ValidateView) Cancel() {
	v.loading = false
}

func (v *view_ValidateView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "[Enter] Dismiss", Action: v.dismiss},
		{Key: "esc", Help: "[Esc] Dismiss", Action: v.dismiss},
	}
}

func (v *view_ValidateView) dismiss() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *view_ValidateView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	switch m := msg.(type) {
	case validationResultMsg:
		v.loading = false
		v.result = m.result
	case validationErrorMsg:
		v.loading = false
		v.errMsg = m.err
	case tea.KeyMsg:
		if result, _, handled := HandleKey(v.KeyBindings(), m); handled {
			return result, nil
		}
	}
	return ViewResult{}, nil
}

func (v *view_ValidateView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)
	errorStyle := lipgloss.NewStyle().Foreground(components.ColorError)
	warningStyle := lipgloss.NewStyle().Foreground(components.ColorWarning)
	successStyle := lipgloss.NewStyle().Foreground(components.ColorSuccess)

	if v.loading || (v.result == nil && v.errMsg == "") {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Validating AGENTS.md..."),
				subStyle.Render("Please wait"),
			),
		)
	}

	if v.result != nil && !v.result.Valid {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("AGENTS.md Validation Failed"),
			"",
			errorStyle.Render(v.result.Explanation),
			"",
		)

		content += subStyle.Render("Required sections:")
		content += "\n"
		for _, section := range []string{
			"Running",
			"Philosophy",
			"Code Guidelines",
			"Testing Philosophy",
			"Design Smells to Watch For",
			"Final Note",
		} {
			content += "  - " + section + "\n"
		}

		content += "\n" + subStyle.Render("[Enter/Esc] Dismiss")
		return ModalFrame(v.size).Render(content)
	}

	if v.errMsg != "" {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("AGENTS.md Validation Error"),
				"",
				errorStyle.Render("Error: "+v.errMsg),
				"",
			),
		)
	}

	if v.result == nil {
		return ModalFrame(v.size).Render("Error: could not validate AGENTS.md")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("AGENTS.md Validation"),
		"",
	)

	if v.result.Valid {
		content += successStyle.Render("All sections valid")
	} else {
		content += errorStyle.Render("Validation failed")
	}
	content += "\n\n"

	if len(v.result.Sections) > 0 {
		content += subStyle.Render("Sections:")
		content += "\n"
		for _, sec := range v.result.Sections {
			status := successStyle.Render("OK")
			if !sec.Valid || sec.Errors > 0 || sec.Warnings > 0 {
				status = warningStyle.Render("Issues")
			}
			content += "  " + status + "  " + sec.Name
			if sec.Errors > 0 || sec.Warnings > 0 {
				content += subStyle.Render(" (" + string(rune('0'+sec.Errors)) + " err, " + string(rune('0'+sec.Warnings)) + " warn)")
			}
			content += "\n"
		}
		content += "\n"
	}

	if len(v.result.Errors) > 0 {
		content += errorStyle.Render("Errors:")
		content += "\n"
		for _, err := range v.result.Errors {
			content += "  - " + err.Message
			if err.Field != "" {
				content += " (" + err.Field + ")"
			}
			content += "\n"
		}
		content += "\n"
	}

	if len(v.result.Warnings) > 0 {
		content += warningStyle.Render("Warnings:")
		content += "\n"
		for _, warn := range v.result.Warnings {
			content += "  - " + warn.Message
			if warn.Field != "" {
				content += " (" + warn.Field + ")"
			}
			content += "\n"
		}
		content += "\n"
	}

	if v.result.Valid && len(v.result.Errors) == 0 && len(v.result.Warnings) == 0 {
		content += successStyle.Render("AGENTS.md is properly configured!")
	} else if !v.result.Valid {
		content += subStyle.Render("Please fix the errors above.")
	}

	if v.result.Explanation != "" {
		content += "\n\n" + subStyle.Render(v.result.Explanation)
	}

	content += "\n\n" + subStyle.Render("[Enter/Esc] Dismiss")

	return ModalFrame(v.size).Render(content)
}

type ValidateViewSnapshot struct {
	Type       string
	Handle     string
	Valid      bool
	SectionCnt int
	ErrorCnt   int
	WarnCnt    int
	ErrMsg     string
}

func (v *view_ValidateView) Snapshot() interface{} {
	sectionCnt := 0
	errorCnt := 0
	warnCnt := 0
	valid := false
	if v.result != nil {
		sectionCnt = len(v.result.Sections)
		errorCnt = len(v.result.Errors)
		warnCnt = len(v.result.Warnings)
		valid = v.result.Valid
	}
	return ValidateViewSnapshot{
		Type:       "ValidateView",
		Handle:     v.handle,
		Valid:      valid,
		SectionCnt: sectionCnt,
		ErrorCnt:   errorCnt,
		WarnCnt:    warnCnt,
		ErrMsg:     v.errMsg,
	}
}
