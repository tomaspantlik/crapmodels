// Package Checkbox slouží pro zobrazení checkboxu
package checkbox

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DefaultSymbols jsou výchozí symboly checkboxu, pokud není nastaveno pomocí WithSymbols()
var DefaultSymbols = Symbols{
	LeftBracket:  "[",
	RightBracket: "]",
	Tick:         "✓",
	Untick:       " ",
}

// Symbols jsou symboly použité pro zobrazení checkboxu
// Použít s WithSymbols() v NewCheckboxModel()
type Symbols struct {
	LeftBracket  string
	RightBracket string
	Tick         string
	Untick       string
}

var (
	// DefaultKeys je výchozí mapování klávesových zkratek
	DefaultKeys = Keys{
		Tick1: tea.KeyEnter.String(),
		Tick2: " ",
	}
)

// Keys je typ pro definování klávesových zkratek
// Vychází z bubbletea.KeyMsg.String()
// Pokud je nastaveno na "", tak se ignoruje
type Keys struct {
	Tick1 string
	Tick2 string
	Tick3 string
	Tick4 string
}

// CheckboxModel je model pro použití v bubbletea aplikaci
// Pro interakci s modelem se používají výhradně receiver funkce, které vracejí
// zpět upravený model
type CheckboxModel struct {
	title  string
	ticked bool

	keys Keys

	symbols Symbols

	titleStyle    lipgloss.Style
	checkboxStyle lipgloss.Style
}

// NewCheckboxModel() je funkce pro vytvoření nového CheckboxModelu
// Nastavuje některé výchozí vlastnosti jako barvy a vzhled
// Pro nastavení vlastností modelu použít jako parametry funkce WithTitle a další
func NewCheckboxModel(options ...func(*CheckboxModel)) CheckboxModel {
	m := CheckboxModel{
		symbols:       DefaultSymbols,
		titleStyle:    lipgloss.NewStyle(),
		checkboxStyle: lipgloss.NewStyle().Bold(true),
	}

	for _, opt := range options {
		opt(&m)
	}

	return m
}

// WithTitleColors() nastaví barvy popisku
func WithTitleColors(fg, bg lipgloss.Color) func(*CheckboxModel) {
	return func(cm *CheckboxModel) {
		cm.titleStyle = lipgloss.NewStyle().
			Foreground(fg).
			Background(bg)
	}
}

// WithTitleColors() nastaví barvy checkboxu
func WithCheckboxColors(fg, bg lipgloss.Color) func(*CheckboxModel) {
	return func(cm *CheckboxModel) {
		cm.checkboxStyle = lipgloss.NewStyle().
			Foreground(fg).
			Background(bg)
	}
}

// WithKeys() definuje vlastní klávesové zkratky modelu
// Jako argument předat typ Keys
// Pokud není použito, model použije výchozí klávesy definované v DefaultKeys
func WithKeys(keys Keys) func(*CheckboxModel) {
	return func(tm *CheckboxModel) {
		tm.keys = keys
	}
}

// WithTitle() definuje popisek checkboxu
func WithTitle(title string) func(*CheckboxModel) {
	return func(wm *CheckboxModel) {
		wm.title = title
	}
}

// WithSymbols() definuje symboly pro zobrazení checkboxu
func WithSymbols(s Symbols) func(*CheckboxModel) {
	return func(wm *CheckboxModel) {
		wm.symbols = s
	}
}

// Update() je standardní definice pro bubbletea
// Návratové proměné jsou rozšířené o bubbletea.Msg
//
// Použití v hlavním modelu - na začátku funkce Update() zavolat:
//
//	m.checkbox, cmd, msg = m.checkbox.Update(msg)
func (m CheckboxModel) Update(msg tea.Msg) (CheckboxModel, tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case m.keys.Tick1, m.keys.Tick2, m.keys.Tick3, m.keys.Tick4:
			m.ticked = !m.ticked

			return m, nil
		}

	}

	return m, msg
}

// View() je standardní funkce pro bubbletea
// Volat v hlavním modelu a výsledek spojit s ostatním výstupem
func (m CheckboxModel) View() string {
	var s string

	if m.ticked {
		s = m.checkboxStyle.Render(m.symbols.LeftBracket + m.symbols.Tick + m.symbols.RightBracket + " ")
	} else {
		s = m.checkboxStyle.Render(m.symbols.LeftBracket + m.symbols.Untick + m.symbols.RightBracket + " ")
	}

	s += m.titleStyle.Render(m.title)

	return s
}

// Tick() nastaví zatržení checkboxu
// Vrací CheckboxModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m CheckboxModel) Tick(t bool) CheckboxModel {
	m.ticked = t

	return m
}

// ToggleTick() přepne zatržení checkboxu
// Vrací CheckboxModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m CheckboxModel) ToggleTick() CheckboxModel {
	m.ticked = !m.ticked

	return m
}

// GetTick() vrátí stav checkboxu
func (m CheckboxModel) GetTick() bool {
	return m.ticked
}

// SetTitle() nastaví popisek checkboxu
// Vrací TableModel, který je potřeba přiřadit/přepsat v hlavním modelu
func (m CheckboxModel) SetTitle(title string) CheckboxModel {
	m.title = title

	return m
}

// SetTitleColors() nastaví barvy popisku checkboxu
func (m CheckboxModel) SetTitleColors(fg, bg lipgloss.Color) CheckboxModel {
	m.titleStyle = m.titleStyle.Foreground(fg).Background(bg)

	return m
}

// SetCheckboxColors() nastaví barvy popisku checkboxu
func (m CheckboxModel) SetCheckboxColors(fg, bg lipgloss.Color) CheckboxModel {
	m.checkboxStyle = m.checkboxStyle.Foreground(fg).Background(bg)

	return m
}
