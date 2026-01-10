package window

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WindowModel struct {
	width, height int

	title   string
	content string

	borderType   lipgloss.Border
	borderStyle  lipgloss.Style
	titleStyle   lipgloss.Style
	contentStyle lipgloss.Style

	contentVPos, contentHPos lipgloss.Position
	contentPadding           int
}

func NewModel(options ...func(*WindowModel)) WindowModel {
	m := WindowModel{
		borderType:   lipgloss.RoundedBorder(),
		borderStyle:  lipgloss.NewStyle().Bold(true),
		titleStyle:   lipgloss.NewStyle().Bold(true),
		contentStyle: lipgloss.NewStyle(),
		contentVPos:  lipgloss.Center,
		contentHPos:  lipgloss.Center,
	}

	for _, opt := range options {
		opt(&m)
	}

	return m
}

// WithTitle() definuje titulek okna
// Pokud není použito nebo je titulek == "", tak se nezobrazuje
func WithTitle(title string) func(*WindowModel) {
	return func(wm *WindowModel) {
		wm.title = title
	}
}

// WithContent() nastaví obsah okna
func WithContent(content string) func(*WindowModel) {
	return func(wm *WindowModel) {
		wm.content = content
	}
}

// WithBorderType() nastaví typ okraje okna
// Pokud není použito, je nastaven výchozí styl lipgloss.RoundedBorder()
func WithBorderType(borderStyle lipgloss.Border) func(*WindowModel) {
	return func(wm *WindowModel) {
		wm.borderType = borderStyle
	}
}

// WithTitleColors() nastaví barvu titulku okna
func WithTitleColors(fg, bg lipgloss.Color) func(*WindowModel) {
	return func(wm *WindowModel) {
		wm.titleStyle = lipgloss.NewStyle().
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// WithBorderColors() nastaví barvy okraje
func WithBorderColors(fg, bg lipgloss.Color) func(*WindowModel) {
	return func(wm *WindowModel) {
		wm.borderStyle = lipgloss.NewStyle().
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// WithContentColors() nastaví barvy obsahu okna
func WithContentColors(fg, bg lipgloss.Color) func(*WindowModel) {
	return func(wm *WindowModel) {
		wm.contentStyle = lipgloss.NewStyle().
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// WithContentPosition() nastaví vertikální a horizontální zarovnání obsahu
func WithContentPosition(vertical, horizontal lipgloss.Position) func(*WindowModel) {
	return func(wm *WindowModel) {
		wm.contentVPos = vertical
		wm.contentHPos = horizontal
	}
}

// WithContentPadding() nastaví okraje okna
func WithContentPadding(p int) func(*WindowModel) {
	return func(wm *WindowModel) {
		wm.contentPadding = p
	}
}

// Update() je standardní definice pro bubbletea
// Návratové proměné jsou rozšířené o bubbletea.Msg
//
// Použití v hlavním modelu - na začátku funkce Update() zavolat:
//
//	m.win, cmd, msg = m.win.Update(msg)
func (m WindowModel) Update(msg tea.Msg) (WindowModel, tea.Msg) {
	return m, msg
}

// View() je standardní funkce pro bubbletea
// Volat v hlavním modelu a výsledek spojit s ostatním výstupem
func (m WindowModel) View() string {
	var s string

	s = m.contentStyle.
		Padding(m.contentPadding).
		Width(m.width - 2).Height(m.height - 2).
		MaxWidth(m.width - 2).MaxHeight(m.height - 2).
		AlignVertical(m.contentVPos).
		AlignHorizontal(m.contentHPos).
		Render(m.content)

	s = m.addBorders(s)

	return s
}

func (m WindowModel) addBorders(content string) string {
	var s string

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

	s = lipgloss.NewStyle().
		BorderStyle(m.borderType).
		BorderTop(false).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBackground(m.borderStyle.GetBackground()).
		BorderForeground(m.borderStyle.GetForeground()).
		Render(content)

	s = lipgloss.JoinVertical(lipgloss.Top, borderTop, s)

	return s
}

// SetContent() nastaví nový obsah, starý obsah zahodí
// Vrací WindowModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m WindowModel) SetContent(content string) WindowModel {
	m.content = content

	return m
}

// GetContent() vrátí obsah okna
func (m WindowModel) GetContent() string {
	return m.content
}

// SetSize() nastaví velikost okna
// Vrací WindowModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m WindowModel) SetSize(width, height int) WindowModel {
	m.width, m.height = width, height

	return m
}

// SetTitle() nastaví titulek okna, pokud je nastaveno na "" tak se nezobrazuje vůbec
// Vrací WindowModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m WindowModel) SetTitle(title string) WindowModel {
	m.title = title

	return m
}
