// Package table slouží pro zobrazení tabulky v okně - každý řádek tabulky lze
// vybírat, pomocí klávesových zkratek je možné procházet tabulku. Je možné nastavit
// i vlastní vzhled okna a textu. Obsah je možné filtrovat a řadit
package table

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SortOrder int

const (
	SortAscendig   SortOrder = iota // řadit vzestupně
	SortDescending                  // řadit sestupně
	SortUnsorted                    // neřadit
)

var sortOrderName = map[SortOrder]string{
	SortAscendig:   "SortAscending",
	SortDescending: "SortDescending",
	SortUnsorted:   "SortUnsorted",
}

func (so SortOrder) String() string {
	return sortOrderName[so]
}

var (
	// DefaultKeys je výchozí mapování klávesových zkratek
	DefaultKeys = Keys{
		SelectLineDown1: "j",
		SelectLineDown2: tea.KeyDown.String(),
		SelectLineUp1:   "k",
		SelectLineUp2:   tea.KeyUp.String(),
		MoveViewDown1:   "J",
		MoveViewDown2:   tea.KeyShiftDown.String(),
		MoveViewUp1:     "K",
		MoveViewUp2:     tea.KeyShiftUp.String(),
		PageDown1:       tea.KeyCtrlD.String(),
		PageDown2:       tea.KeyCtrlF.String(),
		PageDown3:       tea.KeyPgDown.String(),
		PageUp1:         tea.KeyCtrlU.String(),
		PageUp2:         tea.KeyCtrlB.String(),
		PageUp3:         tea.KeyPgUp.String(),
		Top1:            "g",
		Top2:            tea.KeyHome.String(),
		Bottom1:         "G",
		Bottom2:         tea.KeyEnd.String(),
	}
)

// Keys je typ pro definování klávesových zkratek
// Vychází z bubbletea.KeyMsg.String()
// Každá akce může mít více klávesových zkratek (SelectLineDown1, SelectLineDown2, ...)
// Pokud je nastaveno na "", tak se ignoruje
type Keys struct {
	SelectLineDown1 string
	SelectLineDown2 string
	SelectLineDown3 string
	SelectLineUp1   string
	SelectLineUp2   string
	SelectLineUp3   string
	MoveViewDown1   string
	MoveViewDown2   string
	MoveViewDown3   string
	MoveViewUp1     string
	MoveViewUp2     string
	MoveViewUp3     string
	PageDown1       string
	PageDown2       string
	PageDown3       string
	PageUp1         string
	PageUp2         string
	PageUp3         string
	Top1            string
	Top2            string
	Top3            string
	Bottom1         string
	Bottom2         string
	Bottom3         string
}

// TableModel je model pro použití v bubbletea aplikaci
// Pro interakci s modelem se používají výhradně receiver funkce, které vracejí
// zpět upravený model
type TableModel struct {
	width, height int

	title    string
	headers  []string
	content  [][]string
	colSizes []int

	keys Keys

	selectedLine int
	scrolledTop  int

	filter          string
	filterColums    []int
	filteredContent [][]string
	sortByCol       int
	sortOrder       SortOrder
	sortedContent   [][]string

	borderType          lipgloss.Border
	borderStyle         lipgloss.Style
	titleStyle          lipgloss.Style
	scrollBarStyleBar   lipgloss.Style
	scrollBarStyleSpace lipgloss.Style
	headerStyle         lipgloss.Style
	linesStyle          lipgloss.Style
	selectedLineStyle   lipgloss.Style
	filterStyle         lipgloss.Style
}

// NewTableModel() je funkce pro vytvoření nového TableModelu
// Nastavuje některé výchozí vlastnosti jako barvy a vzhled
// Pro nastavení vlastností modelu použít jako parametry funkce WithKeys a další
func NewTableModel(options ...func(*TableModel)) TableModel {
	m := TableModel{
		keys:                DefaultKeys,
		borderType:          lipgloss.RoundedBorder(),
		borderStyle:         lipgloss.NewStyle().Bold(true),
		titleStyle:          lipgloss.NewStyle().Bold(true),
		scrollBarStyleBar:   lipgloss.NewStyle().Bold(true),
		scrollBarStyleSpace: lipgloss.NewStyle().Bold(true),
		headerStyle:         lipgloss.NewStyle().Bold(true),
		linesStyle:          lipgloss.NewStyle(),
		selectedLineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFFFF")).
			Bold(true),
		filterStyle: lipgloss.NewStyle().Italic(true),
		sortOrder:   SortUnsorted,
	}

	for _, opt := range options {
		opt(&m)
	}

	if len(m.headers) == 0 {
		// TODO: tabulka bez headerů?
		panic("nejsou nastaveny headry!")
	}

	if len(m.filterColums) == 0 {
		for i := range len(m.headers) {
			m.filterColums = append(m.filterColums, i)
		}
	}

	return m
}

// TODO: barvy pro procenta

// WithKeys() definuje vlastní klávesové zkratky modelu
// Jako argument předat typ Keys
// Pokud není použito, model použije výchozí klávesy definované v DefaultKeys
func WithKeys(keys Keys) func(*TableModel) {
	return func(tm *TableModel) {
		tm.keys = keys
	}
}

// WithTitle() definuje titulek okna
// Pokud není použito nebo je titulek == "", tak se nezobrazuje
func WithTitle(title string) func(*TableModel) {
	return func(tm *TableModel) {
		tm.title = title
	}
}

// WithHeaders() nastaví headers tabulky
func WithHeaders(headers ...string) func(*TableModel) {
	return func(tm *TableModel) {
		tm.headers = headers
	}
}

// WithContent() nastaví obsah tabulky
func WithContent(content ...[]string) func(*TableModel) {
	return func(tm *TableModel) {
		tm.content = content
		tm.filteredContent = content
	}
}

// WithColSizes() nastaví šířku sloupečků
// Počet hodnot musí být stejný jako počet sloupečků
// Pokud je velikost == 0, tak je použita automatická velikost
func WithColSizes(s ...int) func(*TableModel) {
	return func(tm *TableModel) {
		tm.colSizes = s
	}
}

// WithBorderType() nastaví styl okraje okna
// Pokud není použito, je nastaven výchozí styl lipgloss.RoundedBorder()
func WithBorderType(borderStyle lipgloss.Border) func(*TableModel) {
	return func(tm *TableModel) {
		tm.borderType = borderStyle
	}
}

// WithTitleColors() nastaví barvu titulku (a procent)
func WithTitleColors(fg, bg lipgloss.Color) func(*TableModel) {
	return func(tm *TableModel) {
		tm.titleStyle = lipgloss.NewStyle().
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// WithHeadersColors() nastaví barvu headerů
func WithHeadersColors(fg, bg lipgloss.Color) func(*TableModel) {
	return func(tm *TableModel) {
		tm.headerStyle = lipgloss.NewStyle().
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// WithBorderColors() nastaví barvy okraje
// Nastavuje i barvy scrollbaru, pokud je potřeba nastavit vlastní scrollbar barvy,
// tak prva nastavit WithBorderColors() a pak až WithScrollBarColors()
func WithBorderColors(fg, bg lipgloss.Color) func(*TableModel) {
	return func(tm *TableModel) {
		tm.borderStyle = lipgloss.NewStyle().
			Foreground(fg).Background(bg).
			Bold(true)
		tm.scrollBarStyleBar = lipgloss.NewStyle().
			Foreground(fg).
			Bold(true)
		tm.scrollBarStyleSpace = lipgloss.NewStyle().
			Foreground(bg).
			Bold(true)
	}
}

// WithScrollBarColors() nastaví barvy scrollbaru
func WithScrollBarColors(bar, space lipgloss.Color) func(*TableModel) {
	return func(tm *TableModel) {
		tm.scrollBarStyleBar = lipgloss.NewStyle().
			Foreground(bar).
			Bold(true)
		tm.scrollBarStyleSpace = lipgloss.NewStyle().
			Foreground(space).
			Bold(true)
	}
}

// WithLinesColors() nastaví bavy řádků
func WithLinesColors(fg, bg lipgloss.Color) func(*TableModel) {
	return func(tm *TableModel) {
		tm.linesStyle = lipgloss.NewStyle().
			Foreground(fg).Background(bg)
	}
}

// WithSelectedLineColors() nastaví barvy vybraného řádku
func WithSelectedLineColors(fg, bg lipgloss.Color) func(*TableModel) {
	return func(tm *TableModel) {
		tm.selectedLineStyle = lipgloss.NewStyle().
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// WithFilterColums() nastaví, podle kterých sloupečků se má filtrovat obsah pomocí SetFilter()
// Pokud není nastaveno, filtruje podle všech sloupečků
func WithFilterColums(cols ...int) func(*TableModel) {
	return func(tm *TableModel) {
		tm.filterColums = cols
	}
}

// Init() standardní definice Init() pro bubbletea
func (m TableModel) Init() tea.Cmd {
	return nil
}

// Update() je standardní definice pro bubbletea
// Návratové proměné jsou rozšířené o bubbletea.Msg
//
// Použití v hlavním modelu - na začátku funkce Update() zavolat:
//
//	m.text, cmd, msg = m.text.Update(msg)
//
// Pokud je předána klávesová zkratka, která je v modelu zaregistrovaná pro ovládání,
// model si ji přebere a nepošle ji dál. Ostatní tea.KeyMsg i tea.Msg posílá zpět
func (m TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd, tea.Msg) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		if m.width > msg.Width {
			m.width = msg.Width
		}
		if m.height > msg.Height {
			m.height = msg.Height
		}

	case tea.KeyMsg:

		switch msg.String() {

		case m.keys.SelectLineDown1, m.keys.SelectLineDown2, m.keys.SelectLineDown3:
			m = m.SetSelectedLine(m.selectedLine + 1)

		case m.keys.SelectLineUp1, m.keys.SelectLineUp2, m.keys.SelectLineUp3:
			m = m.SetSelectedLine(m.selectedLine - 1)

		case m.keys.MoveViewDown1, m.keys.MoveViewDown2, m.keys.MoveViewDown3:
			m = m.ViewScroll(1)

		case m.keys.MoveViewUp1, m.keys.MoveViewUp2, m.keys.MoveViewUp3:
			m = m.ViewScroll(-1)

		case m.keys.PageDown1, m.keys.PageDown2, m.keys.PageDown3:
			m = m.PageScroll(1, true)

		case m.keys.PageUp1, m.keys.PageUp2, m.keys.PageUp3:
			m = m.PageScroll(-1, true)

		case m.keys.Top1, m.keys.Top2, m.keys.Top3:
			m = m.SetSelectedLine(0)

		case m.keys.Bottom1, m.keys.Bottom2, m.keys.Bottom3:
			m = m.SelectLastLine()

		default:
			return m, nil, msg

		}

		return m, nil, nil
	}

	return m, nil, msg
}

// View() je standardní funkce pro bubbletea
// Volat v hlavním modelu a výsledek spojit s ostatním výstupem
func (m TableModel) View() string {
	height := m.height
	if m.filter != "" {
		height--
	}

	var (
		s           string
		linesHeight = min(height-3+m.scrolledTop, len(m.sortedContent))
		colSizes    = m.computeColSizes()
	)

	var (
		headers string
		table   string
	)

	for i, h := range m.headers {
		if i > 0 && i < len(m.headers) {
			headers = lipgloss.JoinHorizontal(
				lipgloss.Left,
				headers,
				m.headerStyle.Render(m.borderType.Right),
			)
		}

		if m.sortByCol == i {
			switch m.sortOrder {
			case SortAscendig:
				h = " " + h
			case SortDescending:
				h = " " + h
			}
		}

		stripCol := stripansi.Strip(h)
		hLen := len([]rune(stripCol))
		if hLen > colSizes[i] {
			if hLen > 3 {
				h = string([]rune(stripCol)[:colSizes[i]-1]) + "…"
			}
		}

		headers = lipgloss.JoinHorizontal(
			lipgloss.Left,
			headers,
			m.headerStyle.Width(colSizes[i]).Inline(true).MaxWidth(colSizes[i]).Render(h),
		)
	}

	if m.filter != "" {
		headers = lipgloss.JoinVertical(
			lipgloss.Top, m.filterStyle.Render(" Filtr: "+m.filter), headers,
		)
	}

	selectedLine := m.selectedLine - m.scrolledTop
	for lineNum, line := range m.sortedContent[m.scrolledTop:linesHeight] {
		if lineNum > linesHeight {
			break
		}
		var tl string

		style := m.linesStyle
		if lineNum == selectedLine {
			style = m.selectedLineStyle
		}

		for i, col := range line {
			if i > 0 && i < len(m.headers) {
				tl = lipgloss.JoinHorizontal(
					lipgloss.Left,
					tl,
					style.Render(m.borderType.Right),
				)
			}
			stripCol := stripansi.Strip(col)
			if len([]rune(stripCol)) > colSizes[i] {
				col = string([]rune(stripCol)[:colSizes[i]-1]) + "…"
			}
			tl = lipgloss.JoinHorizontal(
				lipgloss.Left,
				tl,
				style.Width(colSizes[i]).Inline(true).MaxWidth(colSizes[i]).Render(col),
			)
		}
		if lineNum == 0 {
			table = tl
		} else {
			table = lipgloss.JoinVertical(lipgloss.Top, table, tl)
		}
	}

	lines := len(m.sortedContent[m.scrolledTop:linesHeight])
	if lines < height {
		var fill string
		for i := range colSizes {
			if i > 0 && i < len(m.headers) {
				fill = lipgloss.JoinHorizontal(
					lipgloss.Left,
					fill,
					m.linesStyle.Render(m.borderType.Right),
				)
			}
			fill = lipgloss.JoinHorizontal(
				lipgloss.Left,
				fill,
				m.linesStyle.Width(colSizes[i]).Inline(true).MaxWidth(colSizes[i]).Render(" "),
			)
		}

		for i := range height - lines - 3 {
			if len(m.sortedContent) > 0 {
				table = lipgloss.JoinVertical(lipgloss.Top, table, fill)
			} else {
				table += fill
				if i < (height - lines - 4) {
					table += "\n"
				}
			}
		}
	}

	s = lipgloss.JoinVertical(
		lipgloss.Top, headers, table,
	)

	s = m.addBorders(s)

	return s

}

func (m TableModel) addBorders(table string) string {
	contentLength := len(m.sortedContent)

	borderTop := m.borderType.TopLeft
	if m.title == "" {
		borderTop += strings.Repeat(m.borderType.Top, m.width-2)
		borderTop += m.borderStyle.Render(m.borderType.TopRight)
	} else {
		t := m.title
		if len([]rune(m.title)) > m.width-4 {
			t = m.title[:m.width-7] + "..."
		}

		o := len([]rune(t)) % 2
		borderTop += strings.Repeat(
			m.borderType.Top,
			((m.width-1)/2)-(len([]rune(t))/2)-1,
		)
		borderTop += "[" + m.titleStyle.Render(t) + m.borderStyle.Render("]")
		borderTop += m.borderStyle.Render(strings.Repeat(
			m.borderType.Top,
			m.width-((m.width-1)/2)-(len([]rune(t))/2)-3-o,
		))
		borderTop += m.borderStyle.Render(m.borderType.TopRight)
	}
	borderTop = m.borderStyle.Render(borderTop)

	borderLeft := strings.Repeat(m.borderType.Left+"\n", m.height-3)
	borderLeft += m.borderType.Left
	borderLeft = m.borderStyle.Render(borderLeft)

	height := m.height - 3

	var borderRight string
	if contentLength <= height {
		borderRight = strings.Repeat(
			m.borderStyle.Render(m.borderType.Right)+"\n",
			height,
		)
		borderRight += m.borderStyle.Render(m.borderType.Right)
	} else {
		s := m.scrolledTop / ((contentLength - 1) / (height - 1))

		borderRight += m.borderStyle.Render(m.borderType.Right) + "\n"

		if m.scrolledTop > contentLength-height-1 {
			borderRight += strings.Repeat(m.scrollBarStyleSpace.Render("░")+"\n", height-1)
			borderRight += m.scrollBarStyleBar.Render("█")
		} else {
			for l := range height {
				if s == l {
					borderRight += m.scrollBarStyleBar.Render("█")
				} else {
					borderRight += m.scrollBarStyleSpace.Render("░")
				}
				if l < height-1 {
					borderRight += "\n"
				}
			}
		}
	}

	var borderBottom string
	if contentLength == 0 {
		borderBottom = m.borderType.BottomLeft
		borderBottom += strings.Repeat(m.borderType.Bottom, m.width-2)
		borderBottom += m.borderType.BottomRight
	} else if contentLength <= height {
		borderBottom += fmt.Sprintf("[%d/%d]", m.selectedLine+1, len(m.sortedContent)) + m.borderType.Bottom
		borderBottom = m.borderType.BottomLeft +
			strings.Repeat(m.borderType.Bottom, m.width-len(borderBottom)) +
			borderBottom +
			m.borderType.BottomRight
	} else {
		var p float64
		if m.scrolledTop >= contentLength-height {
			p = 100
		} else {
			p = (float64(m.scrolledTop) / float64(contentLength-1)) * 100
		}

		borderBottom = fmt.Sprintf("[%.0f%%]", p) + m.borderType.Bottom
		borderBottom += fmt.Sprintf("[%d/%d]", m.selectedLine+1, len(m.sortedContent))
		borderBottom = m.borderType.BottomLeft +
			strings.Repeat(m.borderType.Bottom, m.width-1-len(borderBottom)) +
			borderBottom +
			m.borderType.Bottom +
			m.borderType.BottomRight
	}
	borderBottom = m.borderStyle.Render(borderBottom)

	ret := lipgloss.JoinHorizontal(lipgloss.Left, borderLeft, table, borderRight)
	ret = lipgloss.JoinVertical(lipgloss.Left, borderTop, ret)
	ret = lipgloss.JoinVertical(lipgloss.Left, ret, borderBottom)

	return ret
}

// Sort() seřadí tabulku podle sloupečku col a ve směru dir
// Pro zrušení řazení předat do dir NoSort
func (m TableModel) Sort(col int, dir SortOrder) TableModel {
	if col > len(m.headers)-1 {
		return m
	}
	m.sortByCol = col
	m.sortOrder = dir

	if dir == SortUnsorted {
		m.sortedContent = m.filteredContent
	} else {
		m.sortedContent = m.sortFilteredContent()
	}

	return m
}

// GetSorting() vrátí sloupeček, podle kterého se řadí a směř řazení
func (m TableModel) GetSorting() (sortCol int, sortOder SortOrder) {
	return m.sortByCol, m.sortOrder
}

// SetFilter() nastaví filtr na tabulce
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SetFilter(filter string) TableModel {
	m.filter = filter
	m.filteredContent = m.filterContent()
	m.sortedContent = m.sortFilteredContent()
	m.scrolledTop = 0
	m.selectedLine = 0

	return m
}

// GetFilter() vrátí nastavený filtr na tabulce
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) GetFilter() string {
	return m.filter
}

// SetHeaders() nastaví nové headery
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SetHeaders(headers ...string) TableModel {
	m.headers = headers

	return m
}

// SetContent() nastaví nové řádky, starý obsah zahodí
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SetContent(rows ...[]string) TableModel {
	m.content = rows
	m.filteredContent = m.filterContent()
	m.sortedContent = m.sortFilteredContent()

	return m
}

// SetColSizes() nastaví šířku sloupečků
// Počet hodnot musí být stejný jako počet sloupečků
// Pokud je velikost == 0, tak je použita automatická velikost
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SetColSizes(s ...int) TableModel {
	m.colSizes = s

	return m
}

// SetContent() přidá další řádky
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) AppendContent(rows ...[]string) TableModel {
	m.content = append(m.content, rows...)
	m.filteredContent = m.filterContent()
	m.sortedContent = m.sortFilteredContent()

	return m
}

// GetContent() vrátí celý obsah, i když je nastavený filter
func (m TableModel) GetContent() [][]string {
	return m.content
}

// GetFilteredContent() vrátí obsah, pokud je nastavený filtr, tak filtrovaný
func (m TableModel) GetFilteredContent() [][]string {
	return m.filteredContent
}

// SetSize() nastaví velikost okna
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SetSize(width, height int) TableModel {
	m.width, m.height = width, height

	return m
}

// SetSelectedLine() nastaví vybraný řádek
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SetSelectedLine(line int) TableModel {
	if line < len(m.filteredContent) && line >= 0 {
		m.selectedLine = line

		if m.selectedLine >= m.scrolledTop+(m.height-4) {
			if m.filter == "" {
				m.scrolledTop = m.selectedLine - (m.height - 4)
			} else {
				m.scrolledTop = m.selectedLine - (m.height - 5)
			}
		} else if m.selectedLine < m.scrolledTop {
			m.scrolledTop = m.selectedLine
		}
	}

	return m
}

// GetSelectedLine() vrátí index aktuálně vybraného řádku
func (m TableModel) GetSelectedLine() int {
	return m.selectedLine
}

// SelectLastLine() nastaví vybraný řádek na poslední
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SelectLastLine() TableModel {
	m = m.SetSelectedLine(len(m.filteredContent) - 1)

	return m
}

// ViewScroll() posune pohled o num řádků dolů/nahoru
// Neposunuje aktuálně vybraný řádek
// Pokud je num < 0, posouvá pohled nahoru o num řádků
// Pokud je num > 0, posouvá pohled dolů o num řádků
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) ViewScroll(num int) TableModel {
	if num > 0 {
		if m.scrolledTop+num <= len(m.filteredContent)-m.height+3 {
			m.scrolledTop += num
		}
	} else if num < 0 {
		if m.scrolledTop+num >= 0 {
			m.scrolledTop += num
		}
	}

	return m
}

// PageScroll() posune pohled o stránku dolů/nahoru
// Pokud je moveSelected == true, posune i aktuálně vybraný řádek
// Pokud je num < 0, posune pohled o num stránek nahoru
// Pokud je num > 0, posune pohled o num stránek dolů
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) PageScroll(num int, moveSelected bool) TableModel {
	height := m.height
	if m.filter != "" {
		height--
	}

	m.scrolledTop += (height - 3) * num
	if num > 0 {
		if moveSelected {
			m.selectedLine = m.scrolledTop + height - 4
		}
	} else if num < 0 {
		if moveSelected {
			m.selectedLine = m.scrolledTop
		}
	}

	if m.scrolledTop < 0 {
		m.scrolledTop = 0
		if moveSelected {
			m.selectedLine = m.scrolledTop
		}
	}

	lenContent := len(m.filteredContent)
	if m.scrolledTop > lenContent-height-3 {
		m.scrolledTop = lenContent - height + 3
		m.scrolledTop = max(m.scrolledTop, 0)
		if moveSelected {
			m.selectedLine = m.scrolledTop + height - 4
			if m.selectedLine > lenContent {
				m.selectedLine = lenContent - 1
			}
		}
	}

	return m
}

// SetTitle() nastaví titulek tabulky, pokud je nastaveno na "" tak se nezobrazuje vůbec
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SetTitle(title string) TableModel {
	m.title = title

	return m
}

func (m TableModel) sortFilteredContent() [][]string {
	s := make([][]string, len(m.filteredContent))
	copy(s, m.filteredContent)

	if m.sortOrder != SortUnsorted {
		c := collate.New(language.Czech)
		slices.SortFunc(s, func(a, b []string) int {
			switch m.sortOrder {
			case SortAscendig:
				return c.CompareString(a[m.sortByCol], b[m.sortByCol])
			case SortDescending:
				return c.CompareString(b[m.sortByCol], a[m.sortByCol])
			default:
				return 0
			}
		})
	}

	return s
}

func (m TableModel) filterContent() [][]string {
	if m.filter == "" {
		return m.content
	}

	var ret [][]string

	filter := strings.ToLower(m.filter)

	for _, line := range m.content {
	line:
		for _, colN := range m.filterColums {
			if strings.Contains(strings.ToLower(line[colN]), filter) {
				ret = append(ret, line)
				break line
			}
		}
	}
	return ret
}

func (m TableModel) computeColSizes() []int {
	colSizes := make([]int, len(m.headers))

	var fixedSize int

	for colNum := range m.headers {
		if m.colSizes[colNum] != 0 {
			colSizes[colNum] = m.colSizes[colNum]
			fixedSize += m.colSizes[colNum]
		}
	}

	for colNum, hCol := range m.headers {
		if m.colSizes[colNum] == 0 {
			l := len([]rune(stripansi.Strip(hCol)))
			if colSizes[colNum] < l {
				colSizes[colNum] = l
			}
		}
	}

	for _, line := range m.filteredContent {
		for colNum, col := range line {
			if m.colSizes[colNum] == 0 {
				l := len([]rune(stripansi.Strip(col)))
				if colSizes[colNum] < l {
					colSizes[colNum] = l
				}
			}
		}
	}

	var varSize, varNum int
	for i, col := range colSizes {
		if m.colSizes[i] == 0 {
			varSize += col
			varNum++
		}
	}

	base := (m.width - (len(m.headers) + 1) - fixedSize) / varNum
	rem := (m.width - (len(m.headers) + 1) - fixedSize) % varNum

	for i := range colSizes {
		if m.colSizes[i] == 0 {
			if rem > 0 {
				colSizes[i] = base + 1
				rem--
			} else {
				colSizes[i] = base
			}
		}
	}

	return colSizes
}
