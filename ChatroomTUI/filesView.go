package ChatroomTUI

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"io"
	"strings"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type item struct {
	name            string
	size            uint64
	percentDownload float64
}

func (i item) FilterValue() string { return i.name }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.name)
	sizeStr := fmt.Sprintf("\n%v\n", humanize.Bytes(i.size))
	fn := func(s ...string) string {
		return itemStyle.Render(strings.Join(s, " ") + sizeStr)
	}
	//itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> "+strings.Join(s, "")) + itemStyle.Render(sizeStr)
		}
	}

	_, err := fmt.Fprint(w, fn(str))
	if err != nil {
		return
	}
}
