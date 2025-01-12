package ChatroomTUI

import (
	"chatTUIv2_0/protocol"
	"chatTUIv2_0/styles"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	gap = "\n\n"
)

type listUpdate struct {
	users []string
}

type messageStruct struct {
	message string
	user    string
}

type fileStruct struct {
	user     string
	namefile string
	idFile   uint
	size     float32
	percent  float64
}

//FORMATO
//TIPO|USUARIO|...

//EJEMPLOS:
//START|USER
//END|USER
//MSG|USER|CONTENIDO...
//FILE|USER|NAMEFILE|TAM|PARTS
//ACKFILE|ID_FILE|NAMEFILE|TAM|PARTS
//PARTFILE|ID_FILE|USER/ALL|NAMEFILE|NO.PART|T.PARTS|CONTENIDO...
//REQFILE|ID_FILE|USER

var (
	userListTitle = styles.SenderStyle.Render("Conected Users:")
)

// TODO:
//  Channels:
//  - UpdateUserList
//  - ReceiveMessages

type ChatModel struct {
	viewportChat    viewport.Model
	viewportUsers   viewport.Model
	isviewusersOpen bool
	viewportFiles   viewport.Model
	isviewfilesOpen bool
	//Todo: change to a struct with string method
	messages  []string
	filesView list.Model
	files     []list.Item
	textarea  textarea.Model
	username  string
	send      chan<- protocol.MessageCommunication
	receive   <-chan protocol.MessageCommunication
	quit      bool
	focused   bool
}

func InitChat(username string, send chan<- protocol.MessageCommunication, rec <-chan protocol.MessageCommunication) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vpc := viewport.New(30, 5)
	vpc.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	vpu := viewport.New(30, 5)
	//change to a getUsers
	vpu.SetContent(
		styles.SenderStyle.Render("Loading..."))

	vpf := viewport.New(30, 5)
	vpf.SetContent("Nothing sent yet...")

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return ChatModel{
		textarea:        ta,
		messages:        []string{},
		viewportChat:    vpc,
		viewportFiles:   vpf,
		viewportUsers:   vpu,
		isviewfilesOpen: false,
		isviewusersOpen: true,
		username:        username,
		send:            send,
		receive:         rec,
		focused:         true,
	}
}

func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink, waitForActivity(m.receive))
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		textAreaCmd      tea.Cmd
		viewPortChatCmd  tea.Cmd
		viewPortUsersCmd tea.Cmd
	)

	m.textarea, textAreaCmd = m.textarea.Update(msg)
	m.viewportChat, viewPortChatCmd = m.viewportChat.Update(msg)
	if m.isviewusersOpen {
		m.viewportUsers, viewPortUsersCmd = m.viewportUsers.Update(msg)
	}

	switch msg := msg.(type) {
	case listUpdate:
		userListContent := userListTitle + "\n"
		for _, user := range msg.users {
			userListContent += " " + user + "\n"
		}
		m.viewportUsers.SetContent(userListContent)
		return m, tea.Batch(waitForActivity(m.receive))
	case messageStruct:
		m.messages = append(m.messages, styles.SenderStyle.Render(msg.user+": ")+msg.message)
		m.viewportChat.SetContent(lipgloss.NewStyle().Width(m.viewportChat.Width).Render(strings.Join(m.messages, "\n")))
		m.viewportChat.GotoBottom()
		return m, tea.Batch(waitForActivity(m.receive))
	case fileStruct:
		m.messages = append(m.messages, styles.SenderStyle.Render(msg.user+": 📄 ")+msg.namefile+" "+styles.HelpStyle.Render(fmt.Sprintf("%v", msg.size)))
		m.viewportChat.SetContent(lipgloss.NewStyle().Width(m.viewportChat.Width).Render(strings.Join(m.messages, "\n")))
		m.viewportChat.GotoBottom()
		return m, tea.Batch(waitForActivity(m.receive))
	case tea.WindowSizeMsg:
		var chatWidth, leftPannelWidth int
		if m.isviewusersOpen || m.isviewfilesOpen {
			leftPannelWidth = msg.Width / 4
			chatWidth = 3 * msg.Width / 4
			if m.isviewfilesOpen {
				m.viewportFiles.Width = leftPannelWidth
				m.viewportFiles.Height = msg.Height - 3*lipgloss.Height("\n") - lipgloss.Height(styles.UnactiveButtonStyle.Render("Send file"))
			}
			if m.isviewusersOpen {
				m.viewportUsers.Width = leftPannelWidth
				m.viewportUsers.Height = msg.Height - 3*lipgloss.Height("\n")
			}
		} else {
			chatWidth = msg.Width
		}

		m.viewportChat.Width = chatWidth
		m.textarea.SetWidth(chatWidth)
		m.viewportChat.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap) - 2*lipgloss.Height("\n")

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.viewportChat.SetContent(lipgloss.NewStyle().Width(m.viewportChat.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewportChat.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quit = true
			return m, tea.Quit
		case tea.KeyCtrlU:
			m.isviewfilesOpen = false
			m.isviewusersOpen = !m.isviewusersOpen
			if !m.isviewusersOpen {
				m.textarea.Focus()
			}
			return m, tea.Batch(tea.WindowSize(), textAreaCmd, viewPortChatCmd, viewPortUsersCmd)
		case tea.KeyCtrlF:
			m.isviewfilesOpen = !m.isviewfilesOpen
			m.isviewusersOpen = false
			if !m.isviewfilesOpen {
				m.textarea.Focus()
			}
			return m, tea.Batch(tea.WindowSize(), textAreaCmd, viewPortChatCmd, viewPortUsersCmd)
		case tea.KeyTab:
			if m.isviewusersOpen == false && m.isviewfilesOpen == false {
				break
			}
			if m.focused {
				m.focused = false
				m.textarea.Blur()
			} else {
				m.focused = true
				m.textarea.Focus()
			}
		case tea.KeyEnter:
			if m.focused == false {
				break
			}
			m.messages = append(m.messages, styles.SenderStyle.Render(m.username+" [You]: ")+m.textarea.Value())
			m.viewportChat.SetContent(lipgloss.NewStyle().Width(m.viewportChat.Width).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewportChat.GotoBottom()
		}
	}
	return m, tea.Batch(textAreaCmd, viewPortChatCmd, viewPortUsersCmd)
}

func (m ChatModel) View() string {
	if m.quit == true {
		return "\n  Goodbye " + m.username + "!\n"
	}

	chat := styles.ViewportsStyle.Render(
		m.viewportChat.View() + gap +
			m.textarea.View())
	var modelView string
	if m.isviewusersOpen {
		userList := styles.RightBorder.Render(m.viewportUsers.View())
		modelView = lipgloss.JoinHorizontal(lipgloss.Top,
			userList,
			chat)

		var actualView string

		if m.focused {
			actualView = "users"
		} else {
			actualView = "chat"
		}

		return lipgloss.JoinVertical(lipgloss.Top,
			modelView,
			styles.HelpStyle.Render(fmt.Sprintf("\n  ctrl+u: close users list • ctrl+f: open files • tab: toggle %s • esc/ctrl+c: finish program", actualView)),
		)
	}

	if m.isviewfilesOpen {

		var actualView, button string
		if m.focused {
			button = styles.UnactiveButtonStyle.Render("Send file")
			actualView = "files"
		} else {
			button = styles.ActiveButtonStyle.Render("Send file")
			actualView = "chat"
		}

		fileList := styles.RightBorder.Render(lipgloss.JoinVertical(lipgloss.Center, m.viewportFiles.View(), button))
		modelView = lipgloss.JoinHorizontal(lipgloss.Top,
			fileList,
			chat)
		return lipgloss.JoinVertical(lipgloss.Top,
			modelView,
			styles.HelpStyle.Render(fmt.Sprintf("\n  ctrl+u: open users list • ctrl+f: close files • tab: toggle %s • esc/ctrl+c: finish program", actualView)),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		chat,
		styles.HelpStyle.Render(fmt.Sprintf("\n  ctrl+u: open users list • ctrl+f: open files • esc/ctrl+c: finish program")),
	)
}

func StartChatSession(username string, sendChan chan<- protocol.MessageCommunication, receiveChan <-chan protocol.MessageCommunication) {
	chatModel := InitChat(username, sendChan, receiveChan)
	if _, err := tea.NewProgram(chatModel).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}

func waitForActivity(recv <-chan protocol.MessageCommunication) tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-recv:
			switch msg.TypeMessage {
			case "List":
				users := strings.Split(msg.Content, ",")
				return listUpdate{users: users}
			case "Msg":
				return messageStruct{
					message: msg.Content,
					user:    msg.User,
				}
			case "File":
				metadata := strings.Split(msg.Content, ",")
				sizeFile, _ := strconv.ParseFloat(metadata[1], 32)
				return fileStruct{
					user:     msg.User,
					namefile: metadata[0],
					size:     float32(sizeFile),
					idFile:   msg.IdOptional,
					percent:  0,
				}
			default:
				users := []string{
					"data",
					"onemore",
				}
				return listUpdate{users: users}
			}
		case <-time.After(time.Second * 2):
			return listUpdate{users: []string{"no one"}}
		}

	}
}
