package views

import (
	"context"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/key"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type CaptureCreateView struct {
	store     workspace.Store
	ctx       context.Context
	handle    string
	workspace *workspace.Workspace
	nameInput textinput.Model
	descInput textinput.Model
	capture   *workspace.Capture
	done      bool
	loading   bool
	size      measure.Window
	step      int
}

func NewCaptureCreateView(s workspace.Store, ctx context.Context, handle string) *CaptureCreateView {
	ws, _ := s.Get(ctx, handle)
	nameTi := textinput.New()
	nameTi.Placeholder = "Capture name (e.g., 'Before migration')"
	nameTi.Prompt = "> "
	nameTi.Focus()

	descTi := textinput.New()
	descTi.Placeholder = "Description (optional)"
	descTi.Prompt = "> "
	descTi.Focus()

	return &CaptureCreateView{
		store:     s,
		ctx:       ctx,
		handle:    handle,
		workspace: ws,
		nameInput: nameTi,
		descInput: descTi,
	}
}

func (v *CaptureCreateView) Init() tea.Cmd { return textinput.Blink }

func (v *CaptureCreateView) SetSize(size measure.Window) {
	v.size = size
}

func (v *CaptureCreateView) OnPush()   {}
func (v *CaptureCreateView) OnResume() {}
func (v *CaptureCreateView) IsLoading() bool {
	return v.loading
}

func (v *CaptureCreateView) Cancel() {
	v.loading = false
}

func (v *CaptureCreateView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if key.IsCancel(msg) {
		return ViewResult{Action: StackPop{}}, nil
	}

	if v.loading {
		return ViewResult{}, nil
	}

	if key.IsEnter(msg) {
		if v.done {
			return ViewResult{Action: StackPop{}}, nil
		}

		if v.step == 0 {
			v.step = 1
			return ViewResult{}, nil
		}

		v.loading = true
		go func() {
			capture, err := v.store.CaptureState(v.ctx, v.handle, workspace.CaptureOptions{
				Name:        v.nameInput.Value(),
				Kind:        workspace.CaptureKindManual,
				Description: v.descInput.Value(),
			})
			v.loading = false
			if err == nil {
				v.capture = capture
				v.done = true
			}
		}()
		return ViewResult{}, nil
	}

	if key.IsTab(msg) {
		v.step = (v.step + 1) % 2
		return ViewResult{}, nil
	}

	if v.step == 0 {
		updated, cmd := v.nameInput.Update(msg)
		v.nameInput = updated
		return ViewResult{}, cmd
	}

	updated, cmd := v.descInput.Update(msg)
	v.descInput = updated
	return ViewResult{}, cmd
}

func (v *CaptureCreateView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)

	if v.loading {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Creating capture..."),
				subStyle.Render("Please wait"),
			),
		)
	}

	if v.done && v.capture != nil {
		var repoLines []string
		for _, ref := range v.capture.GitState {
			status := "clean"
			statusStyle := subStyle
			if ref.Dirty {
				status = "dirty"
				statusStyle = lipgloss.NewStyle().Foreground(components.ColorWarning)
			}
			repoLines = append(repoLines, lipgloss.JoinHorizontal(
				lipgloss.Left,
				"  "+ref.Repository,
				subStyle.Render(" ("+ref.Branch+")"),
				statusStyle.Render(" ["+status+"]"),
			))
		}

		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Capture Created"),
				"",
				subStyle.Render("ID: "),
				lipgloss.NewStyle().Foreground(components.ColorText).Render(v.capture.ID),
				"",
				subStyle.Render("Name: "),
				lipgloss.NewStyle().Foreground(components.ColorText).Render(v.capture.Name),
				"",
				subStyle.Render("Repositories:"),
				lipgloss.JoinVertical(lipgloss.Left, repoLines...),
				"",
				subStyle.Render("[Enter/Esc] Dismiss"),
			),
		)
	}

	if v.workspace == nil {
		return ModalFrame(v.size).Render("Loading...")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Create Capture"),
		"",
		subStyle.Render("Step 1/2: Name"),
		v.nameInput.View(),
		"",
	)

	if v.step == 0 {
		content += subStyle.Render("[Enter] Next  [Tab] Skip description  [Esc] Cancel")
	} else {
		content += subStyle.Render("Step 2/2: Description (optional)")
		content += "\n" + v.descInput.View()
		content += "\n\n" + subStyle.Render("[Enter] Create  [Tab] Back to name  [Esc] Cancel")
	}

	return ModalFrame(v.size).Render(content)
}

type CaptureCreateViewSnapshot struct {
	Type        string
	Handle      string
	Name        string
	Description string
	Step        int
	Done        bool
	CaptureID   string
	RepoCount   int
	GitState    []GitRefSnapshot
}

type GitRefSnapshot struct {
	Repository string
	Branch     string
	Commit     string
	Dirty      bool
}

func (v *CaptureCreateView) Snapshot() interface{} {
	repoCount := 0
	var gitState []GitRefSnapshot
	if v.capture != nil {
		repoCount = len(v.capture.GitState)
		for _, ref := range v.capture.GitState {
			gitState = append(gitState, GitRefSnapshot{
				Repository: ref.Repository,
				Branch:     ref.Branch,
				Commit:     ref.Commit,
				Dirty:      ref.Dirty,
			})
		}
	} else if v.workspace != nil {
		repoCount = len(v.workspace.Repositories)
		for _, r := range v.workspace.Repositories {
			gitState = append(gitState, GitRefSnapshot{
				Repository: r.Name,
				Branch:     r.Ref,
				Commit:     "",
				Dirty:      false,
			})
		}
	}
	return CaptureCreateViewSnapshot{
		Type:        "CaptureCreateView",
		Handle:      v.handle,
		Name:        v.nameInput.Value(),
		Description: v.descInput.Value(),
		Step:        v.step,
		Done:        v.done,
		CaptureID:   "",
		RepoCount:   repoCount,
		GitState:    gitState,
	}
}
