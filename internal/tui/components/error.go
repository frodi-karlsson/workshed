package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

type ErrorCategory int

const (
	CategoryValidation ErrorCategory = iota
	CategoryIO
	CategoryNetwork
	CategoryState
	CategoryUser
	CategoryUnknown
)

type UIError struct {
	Message  string
	Severity ErrorSeverity
	Category ErrorCategory
	Code     string
	Details  string
	Recovery []RecoveryOption
}

type RecoveryOption struct {
	Label    string
	Action   string
	Callback func() bool
}

func NewUIError(message string) UIError {
	return UIError{
		Message:  message,
		Severity: SeverityError,
		Category: CategoryUnknown,
	}
}

func (e UIError) WithSeverity(severity ErrorSeverity) UIError {
	e.Severity = severity
	return e
}

func (e UIError) WithCategory(category ErrorCategory) UIError {
	e.Category = category
	return e
}

func (e UIError) WithCode(code string) UIError {
	e.Code = code
	return e
}

func (e UIError) WithDetails(details string) UIError {
	e.Details = details
	return e
}

func (e UIError) WithRecovery(options ...RecoveryOption) UIError {
	e.Recovery = options
	return e
}

func (e UIError) Error() string {
	return e.Message
}

type ErrorDisplay struct {
	Style        lipgloss.Style
	TitleStyle   lipgloss.Style
	MessageStyle lipgloss.Style
	DetailsStyle lipgloss.Style
	HelpStyle    lipgloss.Style
	OptionsStyle lipgloss.Style
	MaxWidth     int
	WrapText     func(text string, width int) string
}

func NewErrorDisplay() ErrorDisplay {
	ed := ErrorDisplay{}
	ed.Style = lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorError).
		Padding(1)

	ed.TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorError)

	ed.MessageStyle = lipgloss.NewStyle().
		Foreground(ColorText)

	ed.DetailsStyle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	ed.HelpStyle = lipgloss.NewStyle().
		Foreground(ColorVeryMuted).
		MarginTop(1)

	ed.OptionsStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		MarginTop(1)

	ed.MaxWidth = 60
	ed.WrapText = defaultWrapText

	return ed
}

func defaultWrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	var currentLine strings.Builder
	words := strings.Fields(text)

	for i, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		} else {
			result.WriteString(currentLine.String())
			result.WriteString("\n")
			currentLine.Reset()
			currentLine.WriteString(word)
		}

		if i == len(words)-1 {
			result.WriteString(currentLine.String())
		}
	}

	return result.String()
}

func (ed ErrorDisplay) Render(err UIError) string {
	var content []string

	title := "Error"
	switch err.Severity {
	case SeverityInfo:
		title = "Info"
	case SeverityWarning:
		title = "Warning"
	case SeverityCritical:
		title = "Critical Error"
	}

	content = append(content, ed.TitleStyle.Render(title))

	wrappedMessage := ed.WrapText(err.Message, ed.MaxWidth)
	content = append(content, ed.MessageStyle.Render(wrappedMessage))

	if err.Details != "" {
		content = append(content, "")
		wrappedDetails := ed.WrapText(err.Details, ed.MaxWidth)
		content = append(content, ed.DetailsStyle.Render(wrappedDetails))
	}

	if len(err.Recovery) > 0 {
		content = append(content, "")
		content = append(content, ed.OptionsStyle.Render("Options:"))
		for i, opt := range err.Recovery {
			content = append(content, ed.OptionsStyle.Render(fmt.Sprintf("  [%d] %s", i+1, opt.Label)))
		}
	}

	helpText := "[Enter] Dismiss  [q] Quit"
	if len(err.Recovery) > 0 {
		helpText += "  [1-" + fmt.Sprintf("%d", len(err.Recovery)) + "] Select option"
	}
	content = append(content, ed.HelpStyle.Render(helpText))

	return ed.Style.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

type ErrorManager struct {
	errors     []UIError
	currentIdx int
	display    ErrorDisplay
	OnDismiss  func()
	OnRecover  func(option RecoveryOption) bool
}

func NewErrorManager() *ErrorManager {
	return &ErrorManager{
		errors:     []UIError{},
		currentIdx: -1,
		display:    NewErrorDisplay(),
	}
}

func (em *ErrorManager) PushError(err UIError) {
	em.errors = append(em.errors, err)
	if em.currentIdx == -1 {
		em.currentIdx = 0
	}
}

func (em *ErrorManager) HasErrors() bool {
	return len(em.errors) > 0 && em.currentIdx >= 0
}

func (em *ErrorManager) CurrentError() *UIError {
	if em.HasErrors() {
		return &em.errors[em.currentIdx]
	}
	return nil
}

func (em *ErrorManager) NextError() {
	if em.HasErrors() && em.currentIdx < len(em.errors)-1 {
		em.currentIdx++
	}
}

func (em *ErrorManager) PreviousError() {
	if em.HasErrors() && em.currentIdx > 0 {
		em.currentIdx--
	}
}

func (em *ErrorManager) Dismiss() {
	if em.HasErrors() {
		em.errors = append(em.errors[:em.currentIdx], em.errors[em.currentIdx+1:]...)
		if em.currentIdx >= len(em.errors) {
			em.currentIdx = len(em.errors) - 1
		}
		if em.currentIdx < 0 {
			em.currentIdx = -1
		}
	}
	if em.OnDismiss != nil {
		em.OnDismiss()
	}
}

func (em *ErrorManager) Recover(optionIdx int) bool {
	if em.HasErrors() && optionIdx >= 0 && optionIdx < len(em.CurrentError().Recovery) {
		option := em.CurrentError().Recovery[optionIdx]
		if em.OnRecover != nil {
			return em.OnRecover(option)
		}
		if option.Callback != nil {
			return option.Callback()
		}
		return true
	}
	return false
}

func (em *ErrorManager) Render() string {
	if err := em.CurrentError(); err != nil {
		return em.display.Render(*err)
	}
	return ""
}

func (em *ErrorManager) Clear() {
	em.errors = []UIError{}
	em.currentIdx = -1
}

func GetSeverityColor(severity ErrorSeverity) lipgloss.Color {
	switch severity {
	case SeverityInfo:
		return ColorSuccess
	case SeverityWarning:
		return ColorWarning
	case SeverityError:
		return ColorError
	case SeverityCritical:
		return ColorError
	default:
		return ColorError
	}
}

func GetCategoryLabel(category ErrorCategory) string {
	switch category {
	case CategoryValidation:
		return "Validation Error"
	case CategoryIO:
		return "I/O Error"
	case CategoryNetwork:
		return "Network Error"
	case CategoryState:
		return "State Error"
	case CategoryUser:
		return "User Action Required"
	default:
		return "Error"
	}
}
