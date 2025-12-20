// Package tabs slouží pro zobrazení přepínatelných záložek
package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// DefaultKeys je výchozí mapování klávesových zkratek
	DefaultKeys = Keys{
		Next1: tea.KeyTab.String(),
		Next2: tea.KeyCtrlN.String(),
		Prev1: tea.KeyShiftTab.String(),
		Prev2: tea.KeyCtrlP.String(),
	}
)

// Keys je typ pro definování klávesových zkratek
// Vychází z bubbletea.KeyMsg.String()
// Každá akce může mít více klávesových zkratek (Next1, Next2, ...)
// Pokud je nastaveno na "", tak se ignoruje
type Keys struct {
	Next1 string
	Next2 string
	Next3 string
	Prev1 string
	Prev2 string
	Prev3 string
}

// TextModel je model pro použití v bubbletea aplikaci
// Pro interakci s modelem se používají výhradně receiver funkce, které vracejí
// zpět upravený model
type TabsModel struct {
	width, height int

	keys             Keys
	borderType       lipgloss.Border
	tabStyle         lipgloss.Style
	selectedTabStyle lipgloss.Style
	borderStyle      lipgloss.Style

	tabs        []string
	selectedTab int
}

// NewTabsModel() je funkce pro vytvoření nového TabsModelu
// Nastavuje některé výchozí vlastnosti jako barvy a vzhled
// Pro nastavení vlastností modelu použít jako parametry funkce WithKeys a další
func NewTabsModel(options ...func(*TabsModel)) TabsModel {
	t := TabsModel{
		keys:       DefaultKeys,
		borderType: lipgloss.RoundedBorder(),
		tabStyle: lipgloss.NewStyle().
			Align(lipgloss.Center),
		selectedTabStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Bold(true).
			Background(lipgloss.Color("#FFFFFF")).
			Foreground(lipgloss.Color("#000000")),
		borderStyle: lipgloss.NewStyle(),
	}

	for _, opt := range options {
		opt(&t)
	}

	return t
}

// WithKeys() definuje vlastní klávesové zkratky modelu
// Jako argument předat typ Keys
// Pokud není použito, model použije výchozí klávesy definované v DefaultKeys
func WithKeys(keys Keys) func(*TabsModel) {
	return func(tm *TabsModel) {
		tm.keys = keys
	}
}

// WithTabs() nastaví záložky
// Pokud je délka textu tabů větší než šířka, zkracuje se jejich text
func WithTabs(tabs ...string) func(*TabsModel) {
	return func(tm *TabsModel) {
		tm.tabs = tabs
	}
}

// WithBorderType() nastaví styl okraje záložek
// Pokud není použito, je nastaven výchozí styl lipgloss.RoundedBorder()
func WithBorderType(borderStyle lipgloss.Border) func(*TabsModel) {
	return func(tm *TabsModel) {
		tm.borderType = borderStyle
	}
}

// WithTabColors() nastaví barvu pozadí a popředí pro všechny nevybrané taby
func WithTabColors(bg, fg lipgloss.Color) func(*TabsModel) {
	return func(tm *TabsModel) {
		tm.tabStyle = tm.tabStyle.Background(bg).Foreground(fg)
	}
}

// WithSelectedTabColors() nastaví barvu pozadí a popředí pro vybraný tab
func WithSelectedTabColors(bg, fg lipgloss.Color) func(*TabsModel) {
	return func(tm *TabsModel) {
		tm.selectedTabStyle = tm.selectedTabStyle.Background(bg).Foreground(fg)
	}
}

// WithBorderColors() nastaví barvy okraje tabů
func WithBorderColors(fg, bg lipgloss.Color) func(*TabsModel) {
	return func(tm *TabsModel) {
		tm.borderStyle = tm.borderStyle.
			Foreground(fg).Background(bg).
			Bold(true)
	}
}

// Init() standardní definice Init() pro bubbletea
func (m TabsModel) Init() tea.Cmd {
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
//
// Pak použít něco jako toto v hlavním Update() pro přepínání obsahu pomocí tabů:
//
// var cmd tea.Cmd
// switch m.tabs.GetSelectedTab() {
// case 0:
//
//	m.text1, cmd, msg = m.text1.Update(msg)
//	cmds = append(cmds, cmd)
//
// case 1:
//
//		m.text2, cmd, msg = m.text2.Update(msg)
//		cmds = append(cmds, cmd)
//	}
func (m TabsModel) Update(msg tea.Msg) (TabsModel, tea.Cmd, tea.Msg) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		if msg.Height < m.height {
			m.height = msg.Height
		}
		if msg.Width < m.width {
			m.width = msg.Width
		}

	case tea.KeyMsg:
		switch msg.String() {
		case m.keys.Next1, m.keys.Next2, m.keys.Next3:
			if m.selectedTab < len(m.tabs)-1 {
				m.selectedTab++
			} else {
				m.selectedTab = 0
			}

		case m.keys.Prev1, m.keys.Prev2, m.keys.Prev3:
			if m.selectedTab > 0 {
				m.selectedTab--
			} else {
				m.selectedTab = len(m.tabs) - 1
			}
		}

	}

	return m, nil, msg
}

// View() je standardní funkce pro bubbletea
// V hlavním View() použít npař.:
//
// var w string
//
// switch m.tabs.GetSelectedTab() {
// case 0:
//
//	w = m.text1.View()
//
// case 1:
//
//		w = m.text2.View()
//	}
//
// s := lipgloss.JoinHorizontal(lipgloss.Left, m.tabs.View(), w)
//
// return s
func (m TabsModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var s string

	b := m.borderType.TopLeft
	b += strings.Repeat(m.borderType.Top, m.width-2)
	b += m.borderType.TopRight
	s = m.borderStyle.Render(b) + "\n" + s

	for i, tab := range m.tabs {
		t := m.borderStyle.Render(m.borderType.Left)

		var w string
		if len([]rune(tab)) > m.width-3 {
			r := []rune(tab)
			r = r[:m.width-5]
			w = string(r) + ".."
		} else {
			w = tab
		}

		if i == m.selectedTab {
			w = m.selectedTabStyle.Render(w)
			t += w + m.selectedTabStyle.Width(1).Render(">")
			t += m.borderStyle.Render(m.borderType.Left)
		} else {
			w = m.tabStyle.Render(w)
			t += w + m.borderStyle.Render(m.borderType.Left)
		}

		if i < len(m.tabs)-1 {
			t += "\n" + m.borderStyle.Render(m.borderType.MiddleLeft)
			t += m.borderStyle.Render(strings.Repeat(m.borderType.Bottom, m.width-2))
			t += m.borderStyle.Render(m.borderType.MiddleRight) + "\n"
		} else {
			t += "\n" + m.borderStyle.Render(m.borderType.BottomLeft)
			t += m.borderStyle.Render(strings.Repeat(m.borderType.Bottom, m.width-2))
			t += m.borderStyle.Render(m.borderType.BottomRight) + "\n"
		}

		s += t

	}

	return s
}

// GetTabs() vrátí všechny nastavené záložky
func (m TabsModel) GetTabs() []string {
	return m.tabs
}

// GetSelectedTab() vrátí vybranou záložku
func (m TabsModel) GetSelectedTab() int {
	return m.selectedTab
}

// SetTabs() nastaví nové záložky
// Pokud je délka textu tabů větší než šířka, zkracuje se jejich text
func (m TabsModel) SetTabs(tabs ...string) TabsModel {
	m.tabs = tabs

	return m
}

// SetSize() nastaví velikost okna
// Vrací TabsModel, který je potřeba přiřadit/přepsat v hlavním modelu
// Pokud je délka textu tabů větší než šířka, zkracuje se jejich text
func (m TabsModel) SetSize(width, height int) TabsModel {
	m.width, m.height = width, height

	m.tabStyle = m.tabStyle.Width(m.width - 2)
	m.selectedTabStyle = m.selectedTabStyle.Width(m.width - 3)

	return m
}
