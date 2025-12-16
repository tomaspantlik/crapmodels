// Package qm slouží pro zobrazení potvrzovacího okna pro ukončení aplikace
// Lze upravit vzhled (barvy a okraj okna) a klávesové zkratky

package qm

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// DefaultKeys je výchozí mapování klávesových zkratek
	DefaultKeys = Keys{
		Show1:         "q",
		Show2:         tea.KeyEsc.String(),
		Yes1:          "a",
		No1:           "n",
		No2:           tea.KeyEsc.String(),
		Next1:         tea.KeyLeft.String(),
		Next2:         tea.KeyRight.String(),
		Next3:         "h",
		Next4:         "l",
		Next5:         tea.KeyTab.String(),
		SelectButton1: tea.KeyEnter.String(),
		SelectButton2: " ",
	}

	// DefaultQuestion je výchozí text zobrazený nad tlačítky
	DefaultQuestion = "Ukončit aplikaci?"

	// DefaultYes je výchozí text tlačítka pro potvrzení ukončení
	DefaultYes = "[a]no"

	// DefaultYes je výchozí text tlačítka pro zrušení ukončení
	DefaultNo = "[n]e"
)

// Keys je typ pro definování klávesových zkratek
// Vychází z bubbletea.KeyMsg.String()
// Každá akce může mít více klávesových zkratek (Show1, Show2, ...)
// Pokud je nastaveno na "", tak se ignoruje
type Keys struct {
	Show1         string
	Show2         string
	Show3         string
	Yes1          string
	Yes2          string
	Yes3          string
	No1           string
	No2           string
	No3           string
	Next1         string
	Next2         string
	Next3         string
	Next4         string
	Next5         string
	SelectButton1 string
	SelectButton2 string
	SelectButton3 string
}

// QuitModel je model pro použití v bubbletea aplikaci
// Pro interakci s modelem se používají výhradně receiver funkce, které vracejí
// zpět upravený model
type QuitModel struct {
	displayed bool

	selectedButton uint

	screenWidth, screenHeight int

	keys          Keys
	questionStr   string
	yesStr, noStr string

	defaultStyle          lipgloss.Style
	windowStyle           lipgloss.Style
	borderType            lipgloss.Border
	borderStyle           lipgloss.Style
	unselectedButtonStyle lipgloss.Style
	selectedButtonStyle   lipgloss.Style
	whiteSpaceBg          lipgloss.Color
}

// NewQuitModel() je funkce pro vytvoření nového QuitModelu
// Nastavuje některé výchozí vlastnosti jako texty a barvy
// Pro nastavení vlastností modelu použít jako parametry funkce WithKeys a další
func NewQuitModel(options ...func(*QuitModel)) QuitModel {
	qm := QuitModel{
		selectedButton: 0,
		keys:           DefaultKeys,
		questionStr:    DefaultQuestion,
		yesStr:         DefaultYes,
		noStr:          DefaultNo,
		windowStyle:    lipgloss.NewStyle().Bold(true),
		borderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Bold(true),
		selectedButtonStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#FFFFFF")).
			Foreground(lipgloss.Color("#000000")).
			Width(10).Align(lipgloss.Center).
			Underline(true).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()),
		unselectedButtonStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Width(10).Align(lipgloss.Center).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()),
		whiteSpaceBg: lipgloss.Color("#000000"),
	}

	for _, opt := range options {
		opt(&qm)
	}

	return qm
}

// WithKeys() definuje vlastní klávesové zkratky modelu
// Jako argument předat typ Keys
// Pokud není použito, model použije výchozí klávesy definované v DefaultKeys
func WithKeys(keys Keys) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.keys = keys
	}
}

// WithQuestion() definuje vlastní text zobrazený nad tlačítky
// Pokud není použito, použije se DefaultQuestion
func WithQuestion(question string) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.questionStr = question
	}
}

// WithYesNoStr() definuje vlastní texty pro tlačítka
// Pokud není použito, použije se DefaultYes a DefaultNo
func WithYesNoStr(yes, no string) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.yesStr, qm.noStr = yes, no
	}
}

// WithBorderType() definuje typ okraje (lipgloss.Border) okna
// Pokud není použito, použije se lipgloss.RoundedBorder()
func WithBorderType(borderStyle lipgloss.Border) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.borderStyle = qm.borderStyle.BorderStyle(borderStyle).Bold(true)
	}
}

// WithBorderColors() definuje barvu popředí a pozadí okraje okna
func WithBorderColors(fg, bg lipgloss.Color) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.borderStyle = qm.borderStyle.
			BorderBackground(bg).BorderForeground(fg).
			Bold(true)
	}
}

// WithWindowColors() definuje barvu popředí a pozadí okna
func WithWindowColors(fg, bg lipgloss.Color) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.windowStyle = qm.windowStyle.
			Foreground(fg).Background(bg).Bold(true)
	}
}

// WithUnselectedButtonColors() definuje barvu popředí a pozadí tlačítka, které
// není vybráno
func WithUnselectedButtonColors(fg, bg lipgloss.Color) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.unselectedButtonStyle = qm.windowStyle.
			Foreground(fg).Background(bg).
			BorderBackground(qm.windowStyle.GetBackground()).
			Width(10).Align(lipgloss.Center).Bold(true).
			BorderStyle(lipgloss.RoundedBorder())
	}
}

// WithUnselectedButtonColors() definuje barvu popředí a pozadí tlačítka, které
// je vybráno
func WithSelectedButtonColors(fg, bg lipgloss.Color) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.selectedButtonStyle = qm.windowStyle.
			Foreground(fg).Background(bg).
			BorderBackground(qm.windowStyle.GetBackground()).
			Width(10).Align(lipgloss.Center).Bold(true).
			Underline(true).
			BorderStyle(lipgloss.RoundedBorder())
	}
}

// WithWhiteSpaceColor() definuje barvu pozadí za oknem
// Používá se lipgloss.Place(.., lipgloss.WithWhitespaceBackground(bg))
func WithWhiteSpaceColor(bg lipgloss.Color) func(*QuitModel) {
	return func(qm *QuitModel) {
		qm.whiteSpaceBg = bg
	}
}

// Init() standardní definice Init() pro bubbletea
func (m QuitModel) Init() tea.Cmd {
	return nil
}

// Update() je standardní definice pro bubbletea
// Návratové proměné jsou rozšířené o bubbletea.Msg
//
// Použití v hlavním modelu - na začátku funkce Update() zavolat:
//
//	m.quit, cmd, msg = m.quit.Update(msg)
//
// Pokud je okno zobrazeno, model si přebere bubbletea.KeyMsg pro klávesové zkratky
// a nepošle je dál. Pokud okno není zobrazeno, model je pošle zpátky
func (m QuitModel) Update(msg tea.Msg) (QuitModel, tea.Cmd, tea.Msg) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.screenHeight = msg.Height
		m.screenWidth = msg.Width

		return m, nil, msg

	case tea.KeyMsg:
		if !m.displayed {
			if msg.String() == m.keys.Show1 || msg.String() == m.keys.Show2 || msg.String() == m.keys.Show3 {
				m.displayed = true
				return m, nil, nil
			}
			return m, nil, msg
		}

		switch msg.String() {

		case m.keys.Yes1, m.keys.Yes2, m.keys.Yes3:
			return m, tea.Quit, nil

		case m.keys.No1, m.keys.No2, m.keys.No3:
			m.displayed = false
			return m, nil, nil

		case m.keys.Next1, m.keys.Next2, m.keys.Next3, m.keys.Next4, m.keys.Next5:
			if m.selectedButton == 0 {
				m.selectedButton = 1
			} else {
				m.selectedButton = 0
			}

		case m.keys.SelectButton1, m.keys.SelectButton2, m.keys.SelectButton3:
			if m.selectedButton == 0 {
				return m, tea.Quit, nil
			} else {
				m.displayed = false
				return m, nil, msg
			}

		default:
			return m, nil, nil
		}
	}

	return m, nil, msg
}

func (m QuitModel) viewButtons() string {
	var yes, no = m.yesStr, m.noStr
	if len(yes) > 10 {
		yes = m.yesStr[:10]
	}
	if len(no) > 10 {
		no = m.noStr[:10]
	}

	var yesButton, noButton string
	if m.selectedButton == 0 {
		yesButton = m.selectedButtonStyle.Render(yes)
		noButton = m.unselectedButtonStyle.Render(no)
	} else {
		yesButton = m.unselectedButtonStyle.Render(yes)
		noButton = m.selectedButtonStyle.Render(no)
	}

	s := lipgloss.JoinHorizontal(
		lipgloss.Center,
		yesButton,
		m.windowStyle.Height(3).Render("    "),
		noButton,
	)

	return m.windowStyle.Width(40).Align(lipgloss.Center).Render(s)
}

// View() je standardní funkce pro bubbletea, rozšířená o parametr background
//
// Použití v hlavním modelu - volat na konci View() a předat vygenerovaný výstup:
//
//	  func (m model) View() string {
//		   sampleText := m.sampleText.View()
//		   s := m.quit.View(sampleText)
//		   return s
//	  }
//
// Pokud je okno zobrazeno, vrátí funkce výstup jen s oknem, jinak vrátí background
func (m QuitModel) View(background string) string {
	if m.displayed {

		buttons := m.viewButtons()
		q := m.windowStyle.Padding(1, 2).Width(40).Align(lipgloss.Center).Render(m.questionStr)

		s := lipgloss.JoinVertical(lipgloss.Center, q, buttons)
		s = m.borderStyle.Render(s)

		s = lipgloss.Place(
			m.screenWidth, m.screenHeight, lipgloss.Center, lipgloss.Center, s,
			lipgloss.WithWhitespaceBackground(m.whiteSpaceBg),
		)
		return s
	}

	return background
}

// Display() funkce zobrazí okno
func (m QuitModel) Display() QuitModel {
	m.displayed = true

	return m
}
