// Package tm slouží pro zobrazení textu/seznamu v okně - každý řádek textu lze vybírat,
// pomocí klávesových zkratek je možné procházet seznam. Je možné nastavit i
// vlastní vzhled okna a textu

package tm

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
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
		Bottom1:         "G",
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

// TextModel je model pro použití v bubbletea aplikaci
// Pro interakci s modelem se používají výhradně receiver funkce, které vracejí
// zpět upravený model
type TextModel struct {
	screenReady         bool
	width, height       int
	defaultStyle        lipgloss.Style
	borderType          lipgloss.Border
	borderStyle         lipgloss.Style
	scrollBarStyleBar   lipgloss.Style
	scrollBarStyleSpace lipgloss.Style
	titleStyle          lipgloss.Style
	linesStyle          lipgloss.Style
	selectedLineStyle   lipgloss.Style

	title         string
	content       []string
	parsedContent []string

	keys Keys

	selectedLine             int
	selectedParsedLines      []int
	selectableParsedLinesMap map[int]int
	parsedSelectableLinesMap map[int][]int
	scrolledTop              int
}

// NewTextModel() je funkce pro vytvoření nového QuitModelu
// Nastavuje některé výchozí vlastnosti jako barvy a vzhled
// Pro nastavení vlastností modelu použít jako parametry funkce WithKeys a další
func NewTextModel(options ...func(*TextModel)) TextModel {
	m := TextModel{
		defaultStyle:        lipgloss.NewStyle(),
		borderType:          lipgloss.RoundedBorder(),
		borderStyle:         lipgloss.NewStyle().Bold(true),
		scrollBarStyleBar:   lipgloss.NewStyle().Bold(true),
		scrollBarStyleSpace: lipgloss.NewStyle().Bold(true),
		titleStyle:          lipgloss.NewStyle().Bold(true),
		linesStyle:          lipgloss.NewStyle(),
		selectedLineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFFFF")).
			Bold(true),
		keys: DefaultKeys,
	}
	for _, opt := range options {
		opt(&m)
	}

	return m
}

// TODO: barvy pro procenta

// WithKeys() definuje vlastní klávesové zkratky modelu
// Jako argument předat typ Keys
// Pokud není použito, model použije výchozí klávesy definované v DefaultKeys
func WithKeys(keys Keys) func(*TextModel) {
	return func(tm *TextModel) {
		tm.keys = keys
	}
}

// WithTitle() definuje titulek okna
// Pokud není použito nebo je titulek == "", tak se nezobrazuje
func WithTitle(title string) func(*TextModel) {
	return func(tm *TextModel) {
		tm.title = title
	}
}

// WithContent() nastaví obsah
func WithContent(content ...string) func(*TextModel) {
	return func(tm *TextModel) {
		tm.content = content
	}
}

// WithBorderType() nastaví styl okraje okna
// Pokud není použito, je nastaven výchozí styl lipgloss.RoundedBorder()
func WithBorderType(borderStyle lipgloss.Border) func(*TextModel) {
	return func(tm *TextModel) {
		tm.borderType = borderStyle
	}
}

// WithDefaultColors() nastaví výchozí barvy pro vše
// !přepíše již nastavené barvy! např. WithTitleColors - předávat jako první argument
func WithDefaultColors(fg, bg lipgloss.Color) func(*TextModel) {
	return func(tm *TextModel) {
		tm.defaultStyle = tm.defaultStyle.
			Foreground(fg).Background(bg)
		tm.titleStyle = tm.titleStyle.
			Foreground(fg).Background(bg).
			Bold(true)
		tm.borderStyle = tm.borderStyle.
			Foreground(fg).Background(bg).
			Bold(true)
		tm.scrollBarStyleBar = tm.scrollBarStyleBar.
			Background(fg).
			Bold(true)
		tm.scrollBarStyleSpace = tm.scrollBarStyleSpace.
			Background(bg).
			Bold(true)
		tm.linesStyle = tm.linesStyle.
			Foreground(fg).Background(bg)
		tm.selectedLineStyle = tm.selectedLineStyle.
			Foreground(bg).Background(fg)
	}
}

// WithTitleColors() nastaví barvu titulku (a procent)
func WithTitleColors(fg, bg lipgloss.Color) func(*TextModel) {
	return func(tm *TextModel) {
		tm.titleStyle = tm.defaultStyle.
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// WithBorderColors() nastaví barvy okraje
// Nastavuje i barvy scrollbaru, pokud je potřeba nastavit vlastní scrollbar barvy,
// tak prva nastavit WithBorderColors() a pak až WithScrollBarCoors()
func WithBorderColors(fg, bg lipgloss.Color) func(*TextModel) {
	return func(tm *TextModel) {
		tm.borderStyle = tm.defaultStyle.
			Foreground(fg).Background(bg).
			Bold(true)
		tm.scrollBarStyleBar = tm.defaultStyle.
			Background(fg).
			Bold(true)
		tm.scrollBarStyleSpace = tm.defaultStyle.
			Background(bg).
			Bold(true)
	}
}

// WithScrollBarColors() nastaví barvy scrollbaru
func WithScrollBarColors(bar, space lipgloss.Color) func(*TextModel) {
	return func(tm *TextModel) {
		tm.scrollBarStyleBar = tm.defaultStyle.
			Background(bar).
			Bold(true)
		tm.scrollBarStyleSpace = tm.defaultStyle.
			Background(space).
			Bold(true)
	}
}

// WithLinesColors() nastaví bavy textu
func WithLinesColors(fg, bg lipgloss.Color) func(*TextModel) {
	return func(tm *TextModel) {
		tm.linesStyle = tm.defaultStyle.
			Foreground(fg).Background(bg)
	}
}

// WithSelectedLineColors() nastaví barvy vybraného řádku
func WithSelectedLineColors(fg, bg lipgloss.Color) func(*TextModel) {
	return func(tm *TextModel) {
		tm.selectedLineStyle = tm.defaultStyle.
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// Init() standardní definice Init() pro bubbletea
func (m TextModel) Init() tea.Cmd {
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
// model si ji přebere a nepošle je dál. Ostatní tea.KeyMsg i tea.Msg posílá zpět
func (m TextModel) Update(msg tea.Msg) (TextModel, tea.Cmd, tea.Msg) {
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
			if m.selectedLine < len(m.content)-1 {
				m = m.SetSelectedLine(m.selectedLine + 1)
			}

		case m.keys.SelectLineUp1, m.keys.SelectLineUp2, m.keys.SelectLineUp3:
			if m.selectedLine > 0 {
				m = m.SetSelectedLine(m.selectedLine - 1)
			}

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

		}

	}

	return m, nil, msg
}

// View() je standardní funkce pro bubbletea, rozšířená o parametr background
// Volat v hlavním modelu a výsledek spojit s ostatním výstupem
func (m TextModel) View() string {
	var s string

	for lineNum, line := range m.parsedContent {
		if lineNum < m.scrolledTop {
			continue
		}

		if lineNum > m.height-3+m.scrolledTop {
			break
		}

		if lineNum != m.scrolledTop {
			s += "\n"
		}

		s += line
	}

	s = m.linesStyle.Width(m.width - 2).Height(m.height - 2).
		Render(s)
	s = m.addBorders(s)

	return s
}

func (m TextModel) addBorders(text string) string {
	if m.height < 3 {
		return text
	}

	borderTop := m.borderType.TopLeft
	if m.title == "" {
		borderTop += strings.Repeat(m.borderType.Top, m.width-2)
		borderTop += m.borderType.TopRight
	} else {
		t := m.title
		if len(m.title) > m.width-4 {
			t = m.title[:m.width-7] + "..."
		}

		o := len(t) % 2
		borderTop += strings.Repeat(
			m.borderType.Top,
			((m.width-1)/2)-(len(t)/2)-1,
		)
		borderTop += "[" + m.titleStyle.Render(t) + m.borderStyle.Render("]")
		borderTop += m.borderStyle.Render(strings.Repeat(
			m.borderType.Top,
			m.width-((m.width-1)/2)-(len(t)/2)-3-o,
		))
		borderTop += m.borderStyle.Render(m.borderType.TopRight)
	}
	borderTop = m.borderStyle.Render(borderTop)

	borderLeft := strings.Repeat(m.borderType.Left+"\n", m.height-3)
	borderLeft += m.borderType.Left
	borderLeft = m.borderStyle.Render(borderLeft)

	var borderRight string
	if len(m.parsedContent) <= m.height-2 {
		borderRight = strings.Repeat(m.borderType.Right+"\n", m.height-3)
		borderRight += m.borderType.Right
	} else {
		s := m.scrolledTop / ((len(m.parsedContent) - 1) / (m.height - 2))

		if m.scrolledTop > len(m.parsedContent)-m.height+1 {
			borderRight += strings.Repeat(m.scrollBarStyleSpace.Render("░")+"\n", m.height-3)
			borderRight += m.scrollBarStyleBar.Render("█")
		} else {
			for l := range m.height - 2 {
				if s == l {
					borderRight += m.scrollBarStyleBar.Render("█")
				} else {
					borderRight += m.scrollBarStyleSpace.Render("░")
				}
				if l < m.height-3 {
					borderRight += "\n"
				}
			}
		}
	}

	var borderBottom string
	if len(m.parsedContent) <= m.height-2 {
		borderBottom = m.borderType.BottomLeft
		borderBottom += strings.Repeat(m.borderType.Bottom, m.width-2)
		borderBottom += m.borderType.BottomRight
	} else {
		var p float64
		if m.scrolledTop >= len(m.parsedContent)-m.height+2 {
			p = 100
		} else {
			p = (float64(m.scrolledTop) / float64(len(m.parsedContent)-1)) * 100
		}

		borderBottom = fmt.Sprintf("[%.0f%%]", p)
		borderBottom = m.borderType.BottomLeft + strings.Repeat(m.borderType.Bottom, m.width-3-len(borderBottom)) + borderBottom + m.borderType.Bottom + m.borderType.BottomRight
	}
	borderBottom = m.borderStyle.Render(borderBottom)

	ret := lipgloss.JoinHorizontal(lipgloss.Left, borderLeft, text, borderRight)
	ret = lipgloss.JoinVertical(lipgloss.Left, borderTop, ret)
	ret = lipgloss.JoinVertical(lipgloss.Left, ret, borderBottom)

	return ret
}

// AppendContent() přidá další řádky do obsahu
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TextModel) AppendContent(text ...string) TextModel {
	var move bool
	if m.selectedLine == len(m.content)-1 {
		move = true
		m.selectedLine = len(m.content) - 1 + len(text)
	}

	m.content = append(m.content, text...)

	m = m.parseContent()

	if move {
		if len(m.parsedContent) > m.height-2 {
			m.scrolledTop = len(m.parsedContent) - m.height + 2
		}
	}

	return m
}

// GetContent() vrátí celý obsah
func (m TextModel) GetContent() []string {
	return m.content
}

// SetContent() nastaví nový obsah, celý starý obsah je zahozený
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TextModel) SetContent(text ...string) TextModel {
	m.selectedLine = 0
	m.scrolledTop = 0

	m.content = text

	m = m.parseContent()

	return m
}

// SetTitle() nastaví titulek okna, pokud je == "", tak se titulek nezobrazuje
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TextModel) SetTitle(title string) TextModel {
	m.title = title

	return m
}

// parseContent() je interní funkce, která zpracuje nastavený/upravený obsah
// volá se při manipulaci s textem, posouvání a při změně velikosti okna/terminálu
func (m TextModel) parseContent() TextModel {
	// TODO: optimalizace, to je ale prasárna
	if m.width == 0 || m.height == 0 {
		return m
	}

	m.parsedContent = []string{}
	m.selectedParsedLines = []int{}
	m.selectableParsedLinesMap = make(map[int]int)
	m.parsedSelectableLinesMap = make(map[int][]int)

	var parsedContentLineNum int
	for lineNum, line := range m.content {
		wrap := wordwrap.String(line, m.width-2)
		ws := strings.Split(wrap, "\n")

		var ws2 []string
		for _, wl := range ws {
			w := []rune(wl)
			if len(w)-1 >= m.width-2 {
				z := 0
				for i := 0; i < len(w)-1-m.width-2+5; i += (m.width - 2) {
					z = i
					ws2 = append(ws2, string(w[i:i+m.width-2]))
				}
				ws2 = append(ws2, string(w[z+m.width-2:]))
			} else {
				ws2 = append(ws2, string(w))
			}
		}

		for i := range ws2 {
			m.selectableParsedLinesMap[i+parsedContentLineNum] = lineNum
			m.parsedSelectableLinesMap[lineNum] = append(m.parsedSelectableLinesMap[lineNum], i+parsedContentLineNum)
		}

		if lineNum == m.selectedLine {
			for i := range ws2 {
				ws2[i] = m.selectedLineStyle.Render(ws2[i])
				m.selectedParsedLines = append(m.selectedParsedLines, parsedContentLineNum+i)
			}
		}

		m.parsedContent = append(m.parsedContent, ws2...)
		parsedContentLineNum += len(ws2)
	}

	return m
}

// ViewScroll() posune pohled o num řádků dolů/nahoru
// Neposunuje aktuálně vybraný řádek
// Pokud je num < 0, posouvá pohled nahoru o num řádků
// Pokud je num > 0, posouvá pohled dolů o num řádků
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TextModel) ViewScroll(num int) TextModel {
	if num > 0 {
		if m.scrolledTop+num < len(m.parsedContent)-m.height+3 {
			m.scrolledTop += num
		}
	}
	if num < 0 {
		if m.scrolledTop+num > 0 {
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
func (m TextModel) PageScroll(num int, moveSelected bool) TextModel {
	if num > 0 {

		if m.scrolledTop < len(m.parsedContent)-m.height+2 {
			m.scrolledTop += (m.height - 2) * num
			if m.scrolledTop > len(m.parsedContent)-2 {
				m.scrolledTop = len(m.parsedContent) - m.height + 2
				if moveSelected {
					m.selectedLine = len(m.content) - 1
				}
			} else if m.scrolledTop > len(m.parsedContent)-1-m.height-2 {
				m.scrolledTop = len(m.parsedContent) - m.height + 2
				if moveSelected {
					m.selectedLine = len(m.content) - 1
				}
			} else {
				if moveSelected {
					m.selectedLine = m.selectableParsedLinesMap[m.scrolledTop+m.height-3]
					m.scrolledTop = m.parsedSelectableLinesMap[m.selectedLine][len(m.parsedSelectableLinesMap[m.selectedLine])-1] - m.height + 3
				}
			}
		}

	} else if num < 0 {

		if m.scrolledTop > 0 {
			m.scrolledTop -= (m.height - 2) * (-num)
			if m.scrolledTop < 0 {
				m.scrolledTop = 0
				if moveSelected {
					m.selectedLine = 0
				}
			} else {
				if moveSelected {
					m.selectedLine = m.selectableParsedLinesMap[m.scrolledTop]
					m.scrolledTop = m.parsedSelectableLinesMap[m.selectedLine][0]
				}
			}
		}
	}

	m = m.parseContent()

	return m
}

// GetSelectedLine() vrátí index aktuálně vybraného řádku
func (m TextModel) GetSelectedLine() int {
	return m.selectedLine
}

// SetSelectedLine() nastaví vybraný řádek
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TextModel) SetSelectedLine(numLine int) TextModel {
	m.selectedLine = numLine
	m = m.parseContent()

	if m.selectedParsedLines[len(m.selectedParsedLines)-1] > m.scrolledTop+m.height-len(m.selectedParsedLines)-2 {
		m.scrolledTop = m.selectedParsedLines[len(m.selectedParsedLines)-1] - m.height + 3
	}

	if m.selectedParsedLines[0] < m.scrolledTop {
		m.scrolledTop = m.selectedParsedLines[0]
	}

	if m.scrolledTop < 0 {
		m.scrolledTop = 0
	}

	return m
}

// SelectLastLine() nastaví vybraný řádek na poslední
func (m TextModel) SelectLastLine() TextModel {
	if m.selectedLine < len(m.content)-1 {
		m = m.SetSelectedLine(len(m.content) - 1)
	}

	return m
}

// SetSize() nastaví velikost okna
// Vrací TextModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m TextModel) SetSize(width, height int) TextModel {
	m.width, m.height = width, height
	m = m.parseContent()

	return m
}
