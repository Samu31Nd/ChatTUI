package FormLogginTUI

import (
	"chatTUIv2_0/styles"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
)

const gap = 2

type FormLoginModel struct {
	focusIndex    int
	input         textinput.Model
	width, height int
	quit          bool
}

func NewFormModel() *FormLoginModel {
	var t textinput.Model
	t = textinput.New()
	t.Cursor.Style = styles.CursorStyleForm
	t.CharLimit = 32
	t.Placeholder = "Username"
	t.Focus()
	t.PromptStyle = styles.FocusedStyleForm
	t.TextStyle = styles.FocusedStyleForm
	t.Prompt = ""

	m := FormLoginModel{
		input: t,
		quit:  false,
	}
	return &m
}

func (m FormLoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormLoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quit = true
			return m, tea.Quit
		case tea.KeyEnter:
			if m.input.Value() == "" {
				return m, nil
			}
			m.quit = true
			return m, tea.Quit
		}
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m FormLoginModel) View() string {
	pad := strings.Repeat(" ", gap)
	if m.quit == true {
		return ""
	}
	return fmt.Sprintf(
		"\n%vInsert your username:\n"+
			"\n%v\n",
		pad,
		pad+m.input.View(),
	)
}

func GetUsername() string {
	p, err := tea.NewProgram(NewFormModel()).Run()

	if err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}

	name := p.(FormLoginModel).input.Value()

	return name
}
