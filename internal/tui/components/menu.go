package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuItem struct {
	Key     string
	Label   string
	Desc    string
	Section string
}

type MenuSection struct {
	Name  string
	Items []MenuItem
}

type MenuModel struct {
	Sections []MenuSection
	Cols     int
	width    int
	height   int

	selectedSectionIdx int
	selectedItemIdx    int

	pages       [][]MenuSection
	currentPage int
	pagesDirty  bool

	styles MenuStyles
}

const (
	keyWidth         = 3
	labelWidth       = 14
	colGap           = 2
	rowHeight        = 2
	sectionPadding   = 1
	indicatorPadding = 1
	twoColBreakpoint = 60
)

type MenuStyles struct {
	Section       lipgloss.Style
	ItemContainer lipgloss.Style
	Key           lipgloss.Style
	SelectedKey   lipgloss.Style
	Label         lipgloss.Style
	Desc          lipgloss.Style
	PageIndicator lipgloss.Style
}

func NewMenuStyles() MenuStyles {
	return MenuStyles{
		Section: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorText)).
			MarginTop(1).
			MarginBottom(0),
		ItemContainer: lipgloss.NewStyle(),
		Key: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			Width(keyWidth).
			Align(lipgloss.Right).
			MarginRight(1),
		SelectedKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			Bold(true).
			Width(keyWidth).
			Align(lipgloss.Right).
			MarginRight(1),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Width(labelWidth).
			MarginRight(1),
		Desc: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorVeryMuted)),
		PageIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			Align(lipgloss.Center),
	}
}

func NewMenuModel() MenuModel {
	return MenuModel{
		Cols:               2,
		selectedSectionIdx: 0,
		selectedItemIdx:    0,
		width:              70,
		height:             22,
		pagesDirty:         true,
		currentPage:        0,
		styles:             NewMenuStyles(),
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) SetSections(sections []MenuSection) {
	m.Sections = sections
	m.selectedSectionIdx = 0
	m.selectedItemIdx = 0
	m.pagesDirty = true
	m.normalizeCursor()
}

func (m *MenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.recalculateCols()
	m.pagesDirty = true
	m.syncPageToSelection()
}

func (m *MenuModel) TotalItems() int {
	total := 0
	for _, section := range m.Sections {
		total += len(section.Items)
	}
	return total
}

func (m *MenuModel) SelectedItem() *MenuItem {
	if len(m.Sections) == 0 {
		return nil
	}
	m.normalizeCursor()

	section := m.Sections[m.selectedSectionIdx]
	if len(section.Items) == 0 {
		return nil
	}

	item := section.Items[m.selectedItemIdx]
	return &item
}

func (m *MenuModel) normalizeCursor() {
	if len(m.Sections) == 0 {
		m.selectedSectionIdx = 0
		m.selectedItemIdx = 0
		return
	}

	if m.selectedSectionIdx < 0 {
		m.selectedSectionIdx = 0
	} else if m.selectedSectionIdx >= len(m.Sections) {
		m.selectedSectionIdx = len(m.Sections) - 1
	}

	section := m.Sections[m.selectedSectionIdx]
	if len(section.Items) == 0 {
		m.selectedItemIdx = 0
	} else {
		if m.selectedItemIdx < 0 {
			m.selectedItemIdx = 0
		} else if m.selectedItemIdx >= len(section.Items) {
			m.selectedItemIdx = len(section.Items) - 1
		}
	}
}

func (m *MenuModel) syncPageToSelection() {
	m.calculatePages()
	if len(m.pages) == 0 {
		return
	}

	targetSectionName := m.Sections[m.selectedSectionIdx].Name

	for pageIdx, pageSections := range m.pages {
		for _, s := range pageSections {
			if s.Name == targetSectionName {
				m.currentPage = pageIdx
				return
			}
		}
	}
}

func (m *MenuModel) Update(msg tea.KeyMsg) {
	m.recalculateCols()
	m.calculatePages()

	if len(m.Sections) == 0 {
		return
	}

	switch msg.Type {
	case tea.KeyUp:
		m.moveUp()
	case tea.KeyDown:
		m.moveDown()
	case tea.KeyLeft:
		m.moveLeft()
	case tea.KeyRight:
		m.moveRight()
	case tea.KeyTab:
		if msg.String() == "shift+tab" {
			m.prevPage()
		} else {
			m.nextPage()
		}
	}

	m.normalizeCursor()
	m.syncPageToSelection()
}

func (m *MenuModel) moveRight() {
	m.selectedItemIdx++
	section := m.Sections[m.selectedSectionIdx]

	if m.selectedItemIdx >= len(section.Items) {
		m.selectedItemIdx = len(section.Items) - 1
	}
}

func (m *MenuModel) moveLeft() {
	m.selectedItemIdx--
	if m.selectedItemIdx < 0 {
		m.selectedItemIdx = 0
	}
}

func (m *MenuModel) moveDown() {
	// Moving down adds 'Cols' to the index
	newIdx := m.selectedItemIdx + m.Cols
	section := m.Sections[m.selectedSectionIdx]

	if newIdx < len(section.Items) {
		m.selectedItemIdx = newIdx
	} else {
		if m.selectedSectionIdx < len(m.Sections)-1 {
			m.selectedSectionIdx++
			m.selectedItemIdx = 0 // Start at top of next section
		} else {
			m.selectedItemIdx = len(section.Items) - 1
		}
	}
}

func (m *MenuModel) moveUp() {
	newIdx := m.selectedItemIdx - m.Cols

	if newIdx >= 0 {
		m.selectedItemIdx = newIdx
	} else {
		if m.selectedSectionIdx > 0 {
			m.selectedSectionIdx--
			prevSection := m.Sections[m.selectedSectionIdx]

			m.selectedItemIdx = len(prevSection.Items) - 1
		} else {
			m.selectedItemIdx = 0
		}
	}
}

func (m *MenuModel) nextPage() {
	if len(m.pages) <= 1 {
		return
	}
	m.currentPage = (m.currentPage + 1) % len(m.pages)
	m.selectFirstItemOfPage()
}

func (m *MenuModel) prevPage() {
	if len(m.pages) <= 1 {
		return
	}
	m.currentPage = (m.currentPage - 1 + len(m.pages)) % len(m.pages)
	m.selectFirstItemOfPage()
}

func (m *MenuModel) selectFirstItemOfPage() {
	if len(m.pages) == 0 {
		return
	}

	pageSections := m.pages[m.currentPage]
	if len(pageSections) > 0 {
		targetName := pageSections[0].Name
		for i, s := range m.Sections {
			if s.Name == targetName {
				m.selectedSectionIdx = i
				m.selectedItemIdx = 0
				return
			}
		}
	}
}

func (m *MenuModel) recalculateCols() {
	newCols := 1
	if m.width >= twoColBreakpoint {
		newCols = 2
	}
	if m.Cols != newCols {
		m.Cols = newCols
		m.pagesDirty = true
	}
}

func (m *MenuModel) linesForSection(section MenuSection) int {
	if len(section.Items) == 0 {
		return 1 + sectionPadding
	}
	rows := (len(section.Items) + m.Cols - 1) / m.Cols
	return 1 + (rows * rowHeight) + sectionPadding
}

func (m *MenuModel) calculatePages() {
	if !m.pagesDirty {
		return
	}

	var pages [][]MenuSection
	var currentPage []MenuSection
	currentLines := 0
	maxLines := m.height - indicatorPadding

	if maxLines < 5 {
		maxLines = 5
	}

	for _, section := range m.Sections {
		sectionLines := m.linesForSection(section)

		if currentLines+sectionLines > maxLines && len(currentPage) > 0 {
			pages = append(pages, currentPage)
			currentPage = []MenuSection{}
			currentLines = 0
		}

		currentPage = append(currentPage, section)
		currentLines += sectionLines
	}

	if len(currentPage) > 0 {
		pages = append(pages, currentPage)
	}

	if len(pages) == 0 {
		pages = append(pages, []MenuSection{})
	}

	m.pages = pages
	m.pagesDirty = false
}

func (m *MenuModel) View() string {
	m.recalculateCols()
	m.calculatePages()

	if len(m.Sections) == 0 {
		return "No items available"
	}

	if m.currentPage >= len(m.pages) {
		m.currentPage = 0
	}

	pageSections := m.pages[m.currentPage]
	var lines []string

	for _, section := range pageSections {
		lines = append(lines, m.renderSection(section))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	indicator := m.renderPageIndicator()

	return lipgloss.JoinVertical(lipgloss.Left, body, indicator)
}

func (m *MenuModel) renderSection(section MenuSection) string {
	header := m.styles.Section.Render(section.Name)
	if len(section.Items) == 0 {
		return header
	}

	var itemRows []string
	for i := 0; i < len(section.Items); i += m.Cols {
		end := i + m.Cols
		if end > len(section.Items) {
			end = len(section.Items)
		}

		rowItems := section.Items[i:end]
		rowStr := m.renderRowItems(rowItems, section.Name, i)
		itemRows = append(itemRows, rowStr)
	}

	sectionBody := lipgloss.JoinVertical(lipgloss.Left, itemRows...)
	return lipgloss.JoinVertical(lipgloss.Left, header, sectionBody)
}

func (m *MenuModel) renderRowItems(items []MenuItem, sectionName string, startIndex int) string {
	var itemBlocks []string
	for i, item := range items {
		actualIdx := startIndex + i

		isSectionSelected := m.Sections[m.selectedSectionIdx].Name == sectionName
		isItemSelected := m.selectedItemIdx == actualIdx
		isSelected := isSectionSelected && isItemSelected

		itemBlocks = append(itemBlocks, m.renderItem(item, isSelected))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, itemBlocks...)
}

func (m *MenuModel) renderItem(item MenuItem, isSelected bool) string {
	var keyStyled, labelStyled string

	keyStr := "[" + item.Key + "]"

	if isSelected {
		keyStyled = m.styles.SelectedKey.Render(keyStr)
		labelStyled = m.styles.Label.Bold(true).Render(item.Label)
	} else {
		keyStyled = m.styles.Key.Render(keyStr)
		labelStyled = m.styles.Label.Render(item.Label)
	}

	topLine := lipgloss.JoinHorizontal(lipgloss.Left, keyStyled, labelStyled)

	// Robust Width Calculation
	descPadding := keyWidth + 1
	cellWidth := m.calculateCellWidth()
	descWidth := cellWidth - descPadding
	if descWidth < 5 {
		descWidth = 5
	} // Safety floor

	var descStyled string
	if item.Desc != "" {
		descStyled = m.truncateString(item.Desc, descWidth)
		descStyled = m.styles.Desc.Render(descStyled)
	}

	descLine := lipgloss.JoinHorizontal(lipgloss.Left,
		strings.Repeat(" ", descPadding),
		descStyled,
	)

	block := lipgloss.JoinVertical(lipgloss.Left, topLine, descLine)

	return m.styles.ItemContainer.
		Width(cellWidth).
		Render(block) + strings.Repeat(" ", colGap)
}

func (m *MenuModel) truncateString(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 3 {
		return "..."
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	return string(r[:width-3]) + "..."
}

func (m *MenuModel) calculateCellWidth() int {
	availableWidth := m.width
	if availableWidth <= 0 {
		availableWidth = 80
	}

	effectiveWidth := availableWidth - ((m.Cols - 1) * colGap) - 2
	widthPerCol := effectiveWidth / m.Cols

	if widthPerCol < 20 {
		return 20
	}
	return widthPerCol
}

func (m *MenuModel) renderPageIndicator() string {
	pageCount := len(m.pages)
	if pageCount <= 1 {
		return ""
	}

	var dots []string
	for i := 0; i < pageCount; i++ {
		if i == m.currentPage {
			dots = append(dots, "●")
		} else {
			dots = append(dots, "○")
		}
	}
	return m.styles.PageIndicator.Render(strings.Join(dots, " "))
}

func (m *MenuModel) CurrentPage() int {
	return m.currentPage + 1
}

func (m *MenuModel) PageCount() int {
	return len(m.pages)
}

func (m *MenuModel) SelectByKey(key string) *MenuItem {
	for sectionIdx, section := range m.Sections {
		for itemIdx, item := range section.Items {
			if item.Key == key {
				m.selectedSectionIdx = sectionIdx
				m.selectedItemIdx = itemIdx
				m.syncPageToSelection()
				m.pagesDirty = true
				return &item
			}
		}
	}
	return nil
}
