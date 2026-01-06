package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/joho/godotenv/autoload"
)

type State int

const (
	input_state State = iota
	loading_state
	view_state
)

type Styles struct {
	borderColor lipgloss.Color
	input       lipgloss.Style
	view        lipgloss.Style
	spinner     lipgloss.Style
}

func DefaultStyles() *Styles {
	s := new(Styles)

	s.borderColor = lipgloss.Color("36")
	s.input = lipgloss.NewStyle().BorderForeground(s.borderColor).BorderStyle(lipgloss.NormalBorder()).Padding(1).Width(80)
	s.view = lipgloss.NewStyle().BorderForeground(s.borderColor).BorderStyle(lipgloss.RoundedBorder()).Padding(1).Width(80).Height(20)
	s.spinner = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(s.borderColor).Padding(1).Align(lipgloss.Center).Width(80).Height(20).AlignVertical(lipgloss.Center)

	return s
}

type Model struct {
	result string
	query  string
	title  string
	err    error

	spinner spinner.Model
	input   textinput.Model
	styles  *Styles

	state State

	width  int
	height int
}

type errMsg struct{ err error }

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
			if model.state == input_state {
				log.Printf("input: %s", model.input.Value())
				model.state = loading_state
				return model, tea.Batch(
					model.spinner.Tick,
					fetchCmd(model.input.Value()),
				)
			}
		}
	case spinner.TickMsg:
		if model.state == loading_state {
			var cmd tea.Cmd
			model.spinner, cmd = model.spinner.Update(msg)
			return model, cmd
		}
	case errMsg:
		model.err = msg.err
		model.result = model.err.Error()
		model.state = view_state
	case *Json:
		jsonData, _ := json.MarshalIndent(msg, "", " ")
		model.result = string(jsonData)
		model.state = view_state
	}
	model.input, cmd = model.input.Update(msg)
	return model, cmd
}

func (model Model) View() string {
	if model.width == 0 {
		return "Loading..."
	}

	switch model.state {
	case input_state:
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

	case loading_state:
		return lipgloss.Place(
			model.width,
			model.height,
			lipgloss.Center,
			lipgloss.Center,
			model.styles.spinner.Render(
				model.spinner.View(),
			),
		)

	case view_state:
		return lipgloss.Place(
			model.width,
			model.height,
			lipgloss.Center,
			lipgloss.Center,
			model.styles.view.Render(model.result),
		)

	}
	return ""
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

	newSpinner := spinner.New()
	newSpinner.Spinner = spinner.Monkey

	newState := input_state

	return &Model{
		input:   newInput,
		title:   "Website check",
		styles:  newStyles,
		state:   newState,
		spinner: newSpinner,
	}
}

func fetch(query string) (*Json, tea.Msg) {
	api_key := os.Getenv("API_KEY")

	url := "https://eregos.com/api/early"

	body := []byte(fmt.Appendf(nil, `{
	"query": "%s"
	}`, query))

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errMsg{err}
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", api_key)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, errMsg{err}
	}
	defer response.Body.Close()

	post := new(Json)

	derr := json.NewDecoder(response.Body).Decode(post)
	if derr != nil {
		return nil, errMsg{derr}
	}
	if response.StatusCode != http.StatusOK {
		return nil, errMsg{errors.New(response.Status)}
	}

	return post, nil
}

func fetchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		json, msg := fetch(query)
		if msg != nil {
			return msg
		}
		return json
	}
}

type Json struct {
	Host        string `json:"host"`
	Trustscore  int    `json:"trustscore"`
	Trustsignal struct {
		Domain     int `json:"domain"`
		Ownership  int `json:"ownership"`
		Encryption int `json:"encryption"`
		Website    int `json:"website"`
	} `json:"trustsignal"`
}
