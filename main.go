package main

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	PaddingLeft(2).
	Margin(1, 2)

var loadingFinished = make(chan modelData, 1)

type viewOption string

const (
	activeViewEnterPassword viewOption = "ep"
	activeViewListItems     viewOption = "li"
	activeViewItem          viewOption = "i"
)

type model struct {
	selectedItem OnePasswordItem
	loading      bool
	spinner      spinner.Model
	cursor       int
	data         modelData
}

// modelData can't use the model itself because apparently channels have a size limit of 64kb
type modelData struct {
	err           error
	validationMsg string
	activeView    viewOption
	items         list.Model
	thePassword   textinput.Model
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
	)
}

func (i OnePasswordItem) Title() string {
	return i.TheTitle
}
func (i OnePasswordItem) Description() string { return i.AdditionalInformation }
func (i OnePasswordItem) FilterValue() string {
	value := strings.Builder{}
	for _, site := range i.Urls {
		value.WriteString(site.Href)
	}
	value.WriteString(i.TheTitle)
	value.WriteString(i.AdditionalInformation)
	return value.String()
}

func main() {
	m, err := initModel()
	if err != nil {
		log.Fatalf("error, when initModel() for main(). Error: %v", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err = p.Run(); err != nil {
		log.Fatalf("error, during program run. Error: %v", err)
	}
}

func initModel() (model, error) {
	items := list.New(nil, list.NewDefaultDelegate(), 0, 0) // will set width and height later
	items.Title = "Items"

	thePassword := textinput.New()
	thePassword.Placeholder = "the one and only password"
	thePassword.Focus()
	thePassword.CharLimit = 100
	thePassword.Width = 100

	m := model{
		data: modelData{
			items:       items,
			thePassword: thePassword,
			activeView:  activeViewEnterPassword,
		},
	}
	m.resetSpinner()
	return m, nil
}

func (m *model) resetSpinner() {
	s := spinner.New()
	s.Spinner = spinner.Globe
	m.spinner = s
}
