package views

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type ImportWizardView struct {
	store          workspace.Store
	ctx            context.Context
	invocationCtx  workspace.InvocationContext
	step           int
	filePath       string
	jsonData       string
	parsedContext  *workspace.WorkspaceContext
	preserveHandle bool
	force          bool
	input          textinput.Model
	size           measure.Window
	createdWs      *workspace.Workspace
	err            error
	done           bool
}

func NewImportWizardView(ctx context.Context, s workspace.Store, invocationCtx workspace.InvocationContext) ImportWizardView {
	ti := textinput.New()
	ti.Placeholder = "Path to JSON file"
	ti.Prompt = "> "
	ti.Focus()

	return ImportWizardView{
		store:          s,
		ctx:            ctx,
		invocationCtx:  invocationCtx,
		step:           0,
		input:          ti,
		preserveHandle: false,
		force:          false,
	}
}

func (v *ImportWizardView) Init() tea.Cmd { return textinput.Blink }

func (v *ImportWizardView) SetSize(size measure.Window) {
	v.size = size
}

func (v *ImportWizardView) OnPush()         {}
func (v *ImportWizardView) OnResume()       {}
func (v *ImportWizardView) Cancel()         {}
func (v *ImportWizardView) IsLoading() bool { return false }

func (v *ImportWizardView) KeyBindings() []KeyBinding {
	if v.done {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Dismiss", Action: v.dismiss},
		}
	}
	if v.step == 0 {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Next", Action: v.nextStep},
			{Key: "esc", Help: "[Esc] Cancel", Action: v.cancel},
		}
	}
	if v.step == 1 {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Confirm", Action: v.confirmPreview},
			{Key: "esc", Help: "[Esc] Back", Action: v.backStep},
		}
	}
	if v.step == 2 {
		return []KeyBinding{
			{Key: "p", Help: "[p] Toggle preserve handle", Action: v.togglePreserveHandle},
			{Key: "f", Help: "[f] Toggle force overwrite", Action: v.toggleForce},
			{Key: "enter", Help: "[Enter] Import", Action: v.importWorkspace},
			{Key: "esc", Help: "[Esc] Back", Action: v.backStep},
		}
	}
	return []KeyBinding{}
}

func (v *ImportWizardView) nextStep() (ViewResult, tea.Cmd) {
	path := strings.TrimSpace(v.input.Value())
	if path == "" {
		return ViewResult{}, nil
	}
	v.filePath = path

	data, err := v.readFile(path, v.invocationCtx.GetInvocationCWD())
	if err != nil {
		v.err = err
		v.input.Reset()
		return ViewResult{}, nil
	}
	v.jsonData = data

	var ctx workspace.WorkspaceContext
	if err := json.Unmarshal([]byte(data), &ctx); err != nil {
		v.err = err
		v.input.Reset()
		return ViewResult{}, nil
	}
	v.parsedContext = &ctx

	v.step = 1
	v.input.Blur()
	return ViewResult{}, nil
}

func (v *ImportWizardView) confirmPreview() (ViewResult, tea.Cmd) {
	v.step = 2
	v.input.Blur()
	return ViewResult{}, nil
}

func (v *ImportWizardView) backStep() (ViewResult, tea.Cmd) {
	if v.step > 0 {
		v.step--
		if v.step == 0 {
			v.input.Focus()
		}
	}
	return ViewResult{}, nil
}

func (v *ImportWizardView) togglePreserveHandle() (ViewResult, tea.Cmd) {
	v.preserveHandle = !v.preserveHandle
	return ViewResult{}, nil
}

func (v *ImportWizardView) toggleForce() (ViewResult, tea.Cmd) {
	v.force = !v.force
	return ViewResult{}, nil
}

func (v *ImportWizardView) importWorkspace() (ViewResult, tea.Cmd) {
	if v.parsedContext == nil {
		return ViewResult{}, nil
	}
	ws, err := v.store.ImportContext(v.ctx, workspace.ImportOptions{
		Context:        v.parsedContext,
		InvocationCWD:  v.invocationCtx.GetInvocationCWD(),
		PreserveHandle: v.preserveHandle,
		Force:          v.force,
	})
	if err != nil {
		v.err = err
		return ViewResult{}, nil
	}
	v.createdWs = ws
	v.done = true
	return ViewResult{}, nil
}

func (v *ImportWizardView) cancel() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ImportWizardView) dismiss() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ImportWizardView) readFile(userPath, invocationCWD string) (string, error) {
	absPath := userPath
	if !filepath.IsAbs(userPath) {
		absPath = filepath.Join(invocationCWD, userPath)
	}

	cleanedPath := filepath.Clean(absPath)
	if strings.Contains(cleanedPath, "..") {
		return "", fmt.Errorf("invalid path: path traversal not allowed")
	}

	data, err := os.ReadFile(cleanedPath)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}
	return string(data), nil
}

func (v *ImportWizardView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	var cmd tea.Cmd

	if v.step == 0 {
		updatedInput, inputCmd := v.input.Update(msg)
		v.input = updatedInput
		cmd = inputCmd
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		if result, keyCmd, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, keyCmd
		}
	}

	return ViewResult{}, cmd
}

func (v *ImportWizardView) View() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText)

	if v.done {
		if v.createdWs == nil {
			return ModalFrame(v.size).Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					headerStyle.Render("Import Failed"), "\n",
					lipgloss.NewStyle().Foreground(components.ColorError).Render(v.err.Error()), "\n",
					lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Dismiss"),
				),
			)
		}

		var content []string
		content = append(content, headerStyle.Render("Import Successful!"), "\n")
		content = append(content, lipgloss.NewStyle().Foreground(components.ColorMuted).Render("Handle:"), " ", v.createdWs.Handle, "\n")
		content = append(content, lipgloss.NewStyle().Foreground(components.ColorMuted).Render("Purpose:"), " ", v.createdWs.Purpose, "\n")
		content = append(content, lipgloss.NewStyle().Foreground(components.ColorMuted).Render("Repos:"), " ", string(rune('0'+len(v.createdWs.Repositories))), "\n")
		content = append(content, "\n")
		content = append(content, lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Dismiss"))

		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(lipgloss.Left, content...),
		)
	}

	if v.err != nil {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Import Error"), "\n",
				lipgloss.NewStyle().Foreground(components.ColorError).Render(v.err.Error()), "\n",
				lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Retry  [Esc] Cancel"),
			),
		)
	}

	switch v.step {
	case 0:
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Import Workspace"), "\n", "\n",
				"Paste or type the path to an exported JSON file:", "\n",
				v.input.View(), "\n", "\n",
				lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Next  [Esc] Cancel"),
			),
		)
	case 1:
		var repoLines []string
		for _, r := range v.parsedContext.Repositories {
			repoLines = append(repoLines, "  • "+r.URL)
		}
		reposContent := lipgloss.JoinVertical(lipgloss.Left, repoLines...)

		var content []string
		content = append(content, headerStyle.Render("Preview"), "\n", "\n")
		content = append(content, "Purpose: "+v.parsedContext.Purpose, "\n")
		content = append(content, "Repositories:", "\n")
		if len(repoLines) > 0 {
			content = append(content, reposContent, "\n")
		} else {
			content = append(content, lipgloss.NewStyle().Foreground(components.ColorMuted).Render("  No repositories"), "\n")
		}
		content = append(content, "\n")
		content = append(content, lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Continue  [Esc] Back"))

		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(lipgloss.Left, content...),
		)
	case 2:
		var phMark, fMark string
		if v.preserveHandle {
			phMark = "✓"
		}
		if v.force {
			fMark = "✓"
		}

		var content []string
		content = append(content, headerStyle.Render("Options"), "\n", "\n")
		content = append(content, "[p] Preserve handle "+phMark+"  (use original handle instead of generating new)", "\n")
		content = append(content, "[f] Force overwrite "+fMark+"    (overwrite if workspace exists)", "\n")
		content = append(content, "\n")
		content = append(content, lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Import  [Esc] Back"))

		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(lipgloss.Left, content...),
		)
	}

	return ""
}

type ImportWizardViewSnapshot struct {
	Type           string
	Step           int
	FilePath       string
	HasContext     bool
	Purpose        string
	RepoCount      int
	PreserveHandle bool
	Force          bool
	Done           bool
	HasError       bool
}

func (v *ImportWizardView) Snapshot() interface{} {
	return ImportWizardViewSnapshot{
		Type:       "ImportWizardView",
		Step:       v.step,
		FilePath:   v.filePath,
		HasContext: v.parsedContext != nil,
		Purpose: func() string {
			if v.parsedContext != nil {
				return v.parsedContext.Purpose
			}
			return ""
		}(),
		RepoCount: func() int {
			if v.parsedContext != nil {
				return len(v.parsedContext.Repositories)
			}
			return 0
		}(),
		PreserveHandle: v.preserveHandle,
		Force:          v.force,
		Done:           v.done,
		HasError:       v.err != nil,
	}
}
