package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	input_state State = iota
	view_state
)

type Styles struct {
	borderColor lipgloss.Color
	input       lipgloss.Style
	view        lipgloss.Style
}

func DefaultStyles() *Styles {
	s := new(Styles)

	s.borderColor = lipgloss.Color("36")
	s.input = lipgloss.NewStyle().BorderForeground(s.borderColor).BorderStyle(lipgloss.NormalBorder()).Padding(1).Width(80)
	s.view = lipgloss.NewStyle().BorderForeground(s.borderColor).BorderStyle(lipgloss.RoundedBorder()).Padding(1).Width(80).Height(20)

	return s
}

type Model struct {
	result string
	query  string
	title  string
	err    error

	input  textinput.Model
	styles *Styles

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
				model.state = view_state
				return model, fetchCmd()
			}
		}
	case errMsg:
		model.err = msg.err
		model.result = model.err.Error()
		model.state = view_state
	case *Json:
		jsonData, _ := json.MarshalIndent(msg, "", " ")
		model.result = string(jsonData)
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

	newState := input_state

	return &Model{
		input:  newInput,
		title:  "Website check",
		styles: newStyles,
		state:  newState,
	}
}

func fetch() (*Json, tea.Msg) {

	url := "https://dummyjson.com/auth/login"

	body := []byte(`{
	"username": "emilys",
	"password": "emilyspass"
	}`)

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errMsg{err}
	}

	request.Header.Add("Content-Type", "application/json")

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

func fetchCmd() tea.Cmd {
	return func() tea.Msg {
		json, msg := fetch()
		if msg != nil {
			return msg
		}
		return json
	}
}

type Json struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Gender       string `json:"gender"`
	Image        string `json:"image"`
}
