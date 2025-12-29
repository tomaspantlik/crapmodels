// Package table slouží pro zobrazení tabulky v okně - každý řádek tabulky lze
// vybírat, pomocí klávesových zkratek je možné procházet tabulku. Je možné nastavit
// i vlastní vzhled okna a textu
package table

import (
	"fmt"
	"strings"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

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
	table    *table.Table

	keys Keys

	selectedLine int
	scrolledTop  int

	borderType          lipgloss.Border
	borderStyle         lipgloss.Style
	titleStyle          lipgloss.Style
	scrollBarStyleBar   lipgloss.Style
	scrollBarStyleSpace lipgloss.Style
	headerStyle         lipgloss.Style
	linesStyle          lipgloss.Style
	selectedLineStyle   lipgloss.Style
}

// NewTableModel() je funkce pro vytvoření nového TableModelu
// Nastavuje některé výchozí vlastnosti jako barvy a vzhled
// Pro nastavení vlastností modelu použít jako parametry funkce WithKeys a další
func NewTableModel(options ...func(*TableModel)) TableModel {
	m := TableModel{
		table:               table.New().Wrap(false),
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
	}

	for _, opt := range options {
		opt(&m)
	}

	return m
}

// TODO: doplnit funkce pro nastavení pevné šířky sloupců
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
			Background(fg).
			Bold(true)
		tm.scrollBarStyleSpace = lipgloss.NewStyle().
			Background(bg).
			Bold(true)
	}
}

// WithScrollBarColors() nastaví barvy scrollbaru
func WithScrollBarColors(bar, space lipgloss.Color) func(*TableModel) {
	return func(tm *TableModel) {
		tm.scrollBarStyleBar = lipgloss.NewStyle().
			Background(bar).
			Bold(true)
		tm.scrollBarStyleSpace = lipgloss.NewStyle().
			Background(space).
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
		if len(m.content) == 0 {
			break
		}

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
			if m.selectedLine > 0 {
				m = m.SetSelectedLine(0)
			}

		case m.keys.Bottom1, m.keys.Bottom2, m.keys.Bottom3:
			if m.selectedLine < len(m.content)-1 {
				m = m.SetSelectedLine(len(m.content) - 1)
			}

		default:
			return m, nil, msg

		}

		return m, nil, nil
	}

	return m, nil, msg
}

// View() je standardní funkce pro bubbletea, rozšířená o parametr background
// Volat v hlavním modelu a výsledek spojit s ostatním výstupem
func (m TableModel) View() string {
	var (
		s      string
		height = min(m.height-4+m.scrolledTop, len(m.content))
	)

	m.table = m.table.Headers(m.headers...).
		ClearRows().
		Rows(m.content[m.scrolledTop:height]...).
		Width(m.width).
		BorderRight(false).
		BorderBottom(false).
		BorderTop(false).
		BorderLeft(false)

	if m.colSizes != nil {
		m.table = m.table.StyleFunc(func(row, col int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				if m.colSizes[col] != 0 {
					return m.headerStyle.Width(m.colSizes[col])
				} else {
					return m.headerStyle
				}

			case m.selectedLine - m.scrolledTop:
				if m.colSizes[col] != 0 {
					return m.selectedLineStyle.Width(m.colSizes[col])
				} else {
					return m.selectedLineStyle
				}

			default:
				if m.colSizes[col] != 0 {
					return m.linesStyle.Width(m.colSizes[col])
				} else {
					return m.linesStyle
				}
			}
		})
	} else {
		m.table = m.table.StyleFunc(func(row, col int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return m.headerStyle

			case m.selectedLine - m.scrolledTop:
				return m.selectedLineStyle

			default:
				return m.linesStyle
			}
		})
	}

	s += m.table.Render()

	s = m.addBorders(s)

	return s
}

func (m TableModel) addBorders(table string) string {
	if m.height < 5 {
		return table
	}

	borderTop := m.borderType.TopLeft
	if m.title == "" {
		borderTop += strings.Repeat(m.borderType.Top, m.width-2)
		borderTop += m.borderType.TopRight
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

	var borderRight string
	if len(m.content) <= m.height-2 {
		borderRight = strings.Repeat(m.borderType.Right+"\n", m.height-3)
		borderRight += m.borderType.Right
	} else {
		s := m.scrolledTop / ((len(m.content) - 1) / (m.height - 4))

		borderRight += m.borderStyle.Render(m.borderType.Right) + "\n"
		borderRight += m.borderStyle.Render(m.borderType.Right) + "\n"

		if m.scrolledTop > len(m.content)-m.height-5 {
			borderRight += strings.Repeat(m.scrollBarStyleSpace.Render("░")+"\n", m.height-5)
			borderRight += m.scrollBarStyleBar.Render("█")
		} else {
			for l := range m.height - 4 {
				if s == l {
					borderRight += m.scrollBarStyleBar.Render("█")
				} else {
					borderRight += m.scrollBarStyleSpace.Render("░")
				}
				if l < m.height-5 {
					borderRight += "\n"
				}
			}
		}
	}

	var borderBottom string
	if len(m.content) <= m.height-2 {
		borderBottom = m.borderType.BottomLeft
		borderBottom += strings.Repeat(m.borderType.Bottom, m.width-2)
		borderBottom += m.borderType.BottomRight
	} else {
		var p float64
		if m.scrolledTop >= len(m.content)-m.height+3 {
			p = 100
		} else {
			p = (float64(m.scrolledTop) / float64(len(m.content)-1)) * 100
		}

		borderBottom = fmt.Sprintf("[%.0f%%]", p)
		borderBottom = m.borderType.BottomLeft + strings.Repeat(m.borderType.Bottom, m.width-3-len(borderBottom)) + borderBottom + m.borderType.Bottom + m.borderType.BottomRight
	}
	borderBottom = m.borderStyle.Render(borderBottom)

	var fill string
	zb := m.height - 5 - len(m.content)
	if zb > 0 {
		ll := strings.Split(stripansi.Strip(table), "\n")[0]
		// fill += ll
		ls := strings.Split(ll, m.borderType.Right)
		for range zb + 1 {
			// if i > 0 {
			fill += "\n"
			// }
			for _, lf := range ls[:len(ls)-1] {
				fill += strings.Repeat(" ", len([]rune(lf))) + m.borderType.Right
			}
		}
	}
	table = table + fill

	ret := lipgloss.JoinHorizontal(lipgloss.Left, borderLeft, table, borderRight)
	ret = lipgloss.JoinVertical(lipgloss.Left, borderTop, ret)
	ret = lipgloss.JoinVertical(lipgloss.Left, ret, borderBottom)

	return ret
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

	return m
}

// GetContent() vrátí celý obsah
func (m TableModel) GetContent() [][]string {
	return m.content
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
	if line < len(m.content) && line >= 0 {
		m.selectedLine = line

		// 11 > 0 + (15-5)
		if m.selectedLine >= m.scrolledTop+(m.height-5) {
			// = 11 - (15-5) = 1
			// = 12 - (15-5) = 2
			// = 13 - (15-5) = 3
			m.scrolledTop = m.selectedLine - (m.height - 5)
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
	m = m.SetSelectedLine(len(m.content) - 1)

	return m
}

// ViewScroll() posune pohled o num řádků dolů/nahoru
// Neposunuje aktuálně vybraný řádek
// Pokud je num < 0, posouvá pohled nahoru o num řádků
// Pokud je num > 0, posouvá pohled dolů o num řádků
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) ViewScroll(num int) TableModel {
	if num > 0 {
		if m.scrolledTop+num <= len(m.content)-m.height+4 {
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
	if num > 0 {
		m.scrolledTop += (m.height-5)*num + 1
		if moveSelected {
			m.selectedLine = m.scrolledTop + m.height - 5
		}
	} else if num < 0 {
		m.scrolledTop += (m.height-5)*num - 1
		if moveSelected {
			m.selectedLine = m.scrolledTop
		}
	}

	if m.scrolledTop <= 0 {
		m.scrolledTop = 0
		if moveSelected {
			m.selectedLine = m.scrolledTop
		}
	}
	if m.scrolledTop > len(m.content)-m.height-5 {
		m.scrolledTop = len(m.content) - m.height + 4
		if moveSelected {
			m.selectedLine = m.scrolledTop + m.height - 5
		}
	}

	return m
}

// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TableModel) SetTitle(title string) TableModel {
	m.title = title

	return m
}
