package ChatroomTUI

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"
)

type user struct {
	name     string
	selected bool
}

func (u user) FilterValue() string { return u.name }

type userDelegate struct{}

func (u userDelegate) Height() int                             { return 1 }
func (u userDelegate) Spacing() int                            { return 0 }
func (u userDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (u userDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(user)
	if !ok {
		return
	}

	usr := i.name
	if i.selected {
		usr = lipgloss.NewStyle().Underline(true).Render(i.name)
	}
	fn := func(s ...string) string {
		return itemStyle.Render(strings.Join(s, " "))
	}
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, ""))
		}
	}

	_, err := fmt.Fprint(w, fn(usr))
	if err != nil {
		return
	}
}
