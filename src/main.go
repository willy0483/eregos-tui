package main

import (
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	borderColor lipgloss.Color
	input       lipgloss.Style
}

func DefaultStyles() *Styles {
	s := new(Styles)

	s.borderColor = lipgloss.Color("36")
	s.input = lipgloss.NewStyle().BorderForeground(s.borderColor).BorderStyle(lipgloss.NormalBorder()).Padding(1).Width(80)

	return s
}

type Model struct {
	result string
	title  string

	input  textinput.Model
	styles *Styles

	width  int
	height int
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return model, tea.Quit
		case "enter":
			log.Printf("input: %s", model.input.Value())
			model.input.SetValue("")
			return model, nil
		}

	}
	model.input, cmd = model.input.Update(msg)
	return model, cmd
}

func (model Model) View() string {
	if model.width == 0 {
		return "Loading..."
	}
	return lipgloss.Place(
		model.width,
		model.height,
		lipgloss.Center,
		lipgloss.Center,

		lipgloss.JoinVertical(
			lipgloss.Center,
			model.title,
			model.styles.input.Render(model.input.View()),
		),
	)
}

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	model := New()
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func New() *Model {
	newStyles := DefaultStyles()

	newInput := textinput.New()
	newInput.Placeholder = "Enter url"
	newInput.Focus()
	newInput.Width = 74

	return &Model{
		input:  newInput,
		title:  "Website check",
		styles: newStyles,
	}
}
