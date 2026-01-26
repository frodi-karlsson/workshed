package views

import (
	"context"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type TemplateConfigView struct {
	store         workspace.Store
	ctx           context.Context
	pathInput     components.PathCompleter
	varsInput     textinput.Model
	template      string
	templateVars  map[string]string
	varsInputMode bool
	errorMsg      string
	size          measure.Window
}

func NewTemplateConfigView(ctx context.Context, s workspace.Store, template string, templateVars map[string]string) TemplateConfigView {
	pathInput := components.NewPathCompleter()
	pathInput.SetPlaceholder("Template directory path (leave empty for no template)")
	pathInput.SetPrompt("> ")
	if template != "" {
		pathInput.SetValue(template)
	}
	pathInput.Focus()

	varsTi := textinput.New()
	varsTi.Placeholder = "key=value (press Enter to add, leave empty to finish)"
	varsTi.Prompt = "> "
	varsTi.CharLimit = 100
	varsTi.Focus()

	return TemplateConfigView{
		store:        s,
		ctx:          ctx,
		pathInput:    pathInput,
		varsInput:    varsTi,
		template:     template,
		templateVars: templateVars,
	}
}

func (v *TemplateConfigView) Init() tea.Cmd {
	return textinput.Blink
}

func (v *TemplateConfigView) SetSize(size measure.Window) {
	v.size = size
	v.pathInput.SetWidth(size.ContentWidth())
}

func (v *TemplateConfigView) OnPush()   {}
func (v *TemplateConfigView) OnResume() {}
func (v *TemplateConfigView) IsLoading() bool {
	return false
}
func (v *TemplateConfigView) Cancel() {}

func (v *TemplateConfigView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return ViewResult{Action: StackPop{}}, nil
		case tea.KeyEnter:
			if !v.varsInputMode {
				path := strings.TrimSpace(v.pathInput.Value())
				if path != "" {
					info, err := os.Stat(path)
					if err != nil {
						v.errorMsg = "Path does not exist: " + path
						v.pathInput.SetValue("")
						return ViewResult{}, nil
					}
					if !info.IsDir() {
						v.errorMsg = "Not a directory: " + path
						v.pathInput.SetValue("")
						return ViewResult{}, nil
					}
				}
				v.template = path
				v.varsInputMode = true
				return ViewResult{}, textinput.Blink
			} else {
				varInput := strings.TrimSpace(v.varsInput.Value())
				if varInput != "" {
					parts := strings.SplitN(varInput, "=", 2)
					if len(parts) == 2 {
						v.templateVars[parts[0]] = parts[1]
					}
				}
				v.varsInput = textinput.New()
				v.varsInput.Placeholder = "key=value (press Enter to add, leave empty to finish)"
				v.varsInput.Prompt = "> "
				v.varsInput.CharLimit = 100
				v.varsInput.Focus()
				return ViewResult{}, textinput.Blink
			}
		case tea.KeyRight:
			if !v.varsInputMode {
				v.varsInputMode = true
				return ViewResult{}, textinput.Blink
			}
		case tea.KeyLeft:
			if v.varsInputMode {
				v.varsInputMode = false
				return ViewResult{}, textinput.Blink
			}
		}

		if !v.varsInputMode {
			_, cmd := v.pathInput.Update(msg)
			v.errorMsg = ""
			return ViewResult{}, cmd
		} else {
			updated, cmd := v.varsInput.Update(msg)
			v.varsInput = updated
			return ViewResult{}, cmd
		}
	}

	return ViewResult{}, nil
}

func (v *TemplateConfigView) View() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText)

	borderStyle := ModalFrame(v.size)

	if !v.varsInputMode {
		var errorDisplay string
		if v.errorMsg != "" {
			errorDisplay = "\n" + lipgloss.NewStyle().Foreground(components.ColorError).Render(v.errorMsg) + "\n"
		}

		return borderStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Template Directory"), "\n", "\n",
				v.pathInput.View(), "\n", "\n",
				lipgloss.NewStyle().Foreground(components.ColorMuted).Render("Contents will be copied into the workspace."),
				errorDisplay,
				"\n",
				lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Next or skip  [→] Skip  [Esc] Back"),
			),
		)
	}

	var varLines []string
	for k, val := range v.templateVars {
		varLines = append(varLines, "  "+k+"="+val)
	}

	var varsContent string
	if len(varLines) > 0 {
		varsContent = lipgloss.JoinVertical(lipgloss.Left, varLines...)
	} else {
		varsContent = lipgloss.NewStyle().Foreground(components.ColorMuted).Render("  No variables added yet")
	}

	return borderStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Template Variables"), "\n", "\n",
			"Template: "+v.template+"\n",
			"Variables (for {{key}} substitution):", "\n",
			varsContent, "\n",
			v.varsInput.View(), "\n",
			"\n",
			lipgloss.NewStyle().Foreground(components.ColorMuted).Render("Example: env=dev → {{env}}/config.json → dev/config.json"),
			"\n",
			lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Add variable  [←] Back  [Esc] Finish"),
		),
	)
}

type TemplateConfigViewSnapshot struct {
	Type          string
	Path          string
	TemplateVars  map[string]string
	VarsInputMode bool
	HasError      bool
	ErrorMsg      string
}

func (v *TemplateConfigView) Snapshot() interface{} {
	return TemplateConfigViewSnapshot{
		Type:          "TemplateConfigView",
		Path:          v.template,
		TemplateVars:  v.templateVars,
		VarsInputMode: v.varsInputMode,
		HasError:      v.errorMsg != "",
		ErrorMsg:      v.errorMsg,
	}
}
