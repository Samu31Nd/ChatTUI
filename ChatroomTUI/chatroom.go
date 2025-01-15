package ChatroomTUI

import (
	"chatTUIv2_0/protocol"
	"chatTUIv2_0/styles"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"os"
	"strconv"
	"strings"
)

const (
	gap = "\n\n"
)

type openedViews uint

const (
	_onlyChat openedViews = iota
	_chatAndFiles
	_chatAndUsers
)

type focusedView int

const (
	exit focusedView = iota - 1
	fvViewChat
	fvViewUsers
	fvViewFiles
)

type listUpdate struct {
	users []string
}

type messageStruct struct {
	message string
	user    string
}

type ErrorMsg struct {
	msg error
}

type fileStruct struct {
	user     string
	namefile string
	idFile   uint
	size     uint64
	percent  float64
}

var (
	userListTitle = styles.SenderStyle.Render("Conected Users:")
)

type ChatModel struct {
	textarea        textarea.Model
	viewportChat    viewport.Model
	viewportUsers   viewport.Model
	viewportFiles   list.Model
	messages        []string
	files           []list.Item
	username        string
	openedViewports openedViews
	focusedViewport focusedView
	errorComp       error
	send            chan<- protocol.MessageCommunication
	receive         <-chan protocol.MessageCommunication
}

func InitChat(username string, send chan<- protocol.MessageCommunication, rec <-chan protocol.MessageCommunication) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vpc := viewport.New(30, 5)
	vpc.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)
	vpu := viewport.New(30, 5)
	vpu.SetContent(
		styles.SenderStyle.Render("Loading..."))

	files := make([]list.Item, 0)
	lf := list.New(files, itemDelegate{}, 100, 10)
	lf.SetShowHelp(false)
	lf.Title = "Files"
	lf.DisableQuitKeybindings()

	return ChatModel{
		textarea:        ta,
		viewportChat:    vpc,
		viewportUsers:   vpu,
		viewportFiles:   lf,
		messages:        []string{},
		files:           files,
		username:        username,
		openedViewports: _onlyChat,
		focusedViewport: fvViewChat,
		send:            send,
		receive:         rec,
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
		viewPortFilesCmd tea.Cmd
	)

	typeKeyMsg := false
	if _, ok := msg.(tea.KeyMsg); ok {
		typeKeyMsg = true
	}

	if !typeKeyMsg {
		m.viewportChat, viewPortChatCmd = m.viewportChat.Update(msg)
		m.viewportFiles, viewPortFilesCmd = m.viewportFiles.Update(msg)
		m.viewportUsers, viewPortUsersCmd = m.viewportUsers.Update(msg)
		m.textarea, textAreaCmd = m.textarea.Update(msg)
	} else {
		switch m.focusedViewport {
		case fvViewFiles:
			m.viewportFiles, viewPortFilesCmd = m.viewportFiles.Update(msg)
		case fvViewUsers:
			m.viewportUsers, viewPortUsersCmd = m.viewportUsers.Update(msg)
		case fvViewChat:
			m.viewportChat, viewPortChatCmd = m.viewportChat.Update(msg)
			m.textarea, textAreaCmd = m.textarea.Update(msg)
		default:
		}
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
		m.messages = append(m.messages, styles.SenderStyle.Render(msg.user+": 📄 ")+msg.namefile+" "+styles.HelpStyle.Render(fmt.Sprintf("%v", humanize.Bytes(msg.size))))
		m.viewportChat.SetContent(lipgloss.NewStyle().Width(m.viewportChat.Width).Render(strings.Join(m.messages, "\n")))
		m.viewportChat.GotoBottom()
		m.viewportFiles.InsertItem(0, item{
			name:            msg.namefile,
			size:            msg.size,
			percentDownload: 0,
		})
		return m, tea.Batch(waitForActivity(m.receive))
	case tea.WindowSizeMsg:
		chatWidth, leftPannelWidth := msg.Width, 0
		if m.openedViewports == _chatAndUsers || m.openedViewports == _chatAndFiles {
			leftPannelWidth = msg.Width / 4
			chatWidth = 3 * msg.Width / 4
			m.viewportFiles.SetWidth(leftPannelWidth)
			m.viewportFiles.SetHeight(msg.Height - 12*lipgloss.Height("\n") - lipgloss.Height(styles.UnactiveButtonStyle.Render("Send file")))
			m.viewportUsers.Width = leftPannelWidth
			m.viewportUsers.Height = msg.Height - 3*lipgloss.Height("\n")
		}

		m.viewportChat.Width = chatWidth
		m.textarea.SetWidth(chatWidth)
		m.viewportChat.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap) - 2*lipgloss.Height("\n")
		if len(m.messages) > 0 {
			m.viewportChat.SetContent(lipgloss.NewStyle().Width(m.viewportChat.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewportChat.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.focusedViewport = exit
			return m, tea.Quit
		case tea.KeyCtrlU:
			if m.openedViewports == _chatAndUsers {
				m.openedViewports = _onlyChat
				m.focusedViewport = fvViewChat
				m.textarea.Focus()
			} else {
				m.openedViewports = _chatAndUsers
			}
			return m, tea.Batch(tea.WindowSize(), textAreaCmd, viewPortChatCmd, viewPortUsersCmd)
		case tea.KeyCtrlF:
			if m.openedViewports == _chatAndFiles {
				m.openedViewports = _onlyChat
				m.focusedViewport = fvViewChat
				m.textarea.Focus()
			} else {
				m.openedViewports = _chatAndFiles
			}

			return m, tea.Batch(tea.WindowSize(), textAreaCmd, viewPortChatCmd, viewPortFilesCmd)
		case tea.KeyTab:
			if m.focusedViewport != fvViewChat {
				m.focusedViewport = fvViewChat
				m.textarea.Focus()
			} else {
				m.textarea.Blur()
				switch m.openedViewports {
				case _chatAndFiles:
					m.focusedViewport = fvViewFiles
				case _chatAndUsers:
					m.focusedViewport = fvViewUsers
				default:
				}
			}

		case tea.KeyEnter:
			switch m.focusedViewport {
			case fvViewChat:
				go m.sendMessage(m.textarea.Value())
				m.messages = append(m.messages, styles.SenderStyle.Render(m.username+" [You]: ")+m.textarea.Value())
				m.viewportChat.SetContent(lipgloss.NewStyle().Width(m.viewportChat.Width).Render(strings.Join(m.messages, "\n")))
				m.textarea.Reset()
				m.viewportChat.GotoBottom()
			default:
			}
		}
	case ErrorMsg:
		m.errorComp = msg.msg
		return m, tea.Quit
	}
	return m, tea.Batch(textAreaCmd, viewPortChatCmd, viewPortUsersCmd, viewPortFilesCmd)
}

func (m ChatModel) View() string {
	if m.errorComp != nil {
		err := m.errorComp.Error()
		go m.sendCloseConnection()
		return styles.ErrorStyle.Render("\nError encontrado: ") + err + "\n"
	}

	if m.focusedViewport == exit {
		go m.sendCloseConnection()
		return "\n  Goodbye " + m.username + "!\n"
	}

	var nextToggleView string

	chat := styles.FullChatViewStyle.Render(
		m.viewportChat.View() + gap +
			m.textarea.View())

	switch m.openedViewports {
	case _chatAndUsers:
		userList := styles.ViewportsStyle.Render(m.viewportUsers.View())
		modelView := lipgloss.JoinHorizontal(lipgloss.Top,
			userList,
			chat)
		if m.focusedViewport == fvViewChat {
			nextToggleView = "users"
		} else {
			nextToggleView = "chat"
		}
		return lipgloss.JoinVertical(lipgloss.Top,
			modelView,
			styles.HelpStyle.Render(fmt.Sprintf("\n  ctrl+u: close users list • ctrl+f: open files • tab: toggle %s • esc/ctrl+c: finish program", nextToggleView)),
		)
		////////////////////////////////
		////////////////////////////////
	case _chatAndFiles:
		var button string
		if m.focusedViewport == fvViewChat {
			button = styles.UnactiveButtonStyle.Render("Send file")
			nextToggleView = "files"
		} else {
			button = styles.ActiveButtonStyle.Render("Send file")
			nextToggleView = "chat"
		}
		fileList := styles.ViewportsStyle.Render(lipgloss.JoinVertical(lipgloss.Center, m.viewportFiles.View(), button))
		modelView := lipgloss.JoinHorizontal(lipgloss.Top,
			fileList,
			chat)
		return lipgloss.JoinVertical(lipgloss.Top,
			modelView,
			styles.HelpStyle.Render(fmt.Sprintf("\n  ctrl+u: open users list • ctrl+f: close files • tab: toggle %s • esc/ctrl+c: finish program", nextToggleView)),
		)
		////////////////////////////////
		////////////////////////////////
	case _onlyChat:
		return lipgloss.JoinVertical(lipgloss.Top,
			chat,
			styles.HelpStyle.Render(fmt.Sprintf("\n  ctrl+u: open users list • ctrl+f: open files • esc/ctrl+c: finish program")),
		)
	default:
		return ""
	}
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
					size:     uint64(sizeFile),
					idFile:   msg.IdOptional,
					percent:  0,
				}
			case "Error":
				return ErrorMsg{msg: errors.New(msg.Content)}
			default:
				return struct{}{}
			}
			//case <-time.After(time.Second * 2):
			//Aun nada
		}

	}
}

func (m ChatModel) sendCloseConnection() {
	m.send <- protocol.MessageCommunication{
		TypeMessage: "exit",
	}
}

func (m ChatModel) sendMessage(message string) {
	m.send <- protocol.MessageCommunication{
		TypeMessage: "msg",
		User:        m.username,
		Content:     message,
	}
}
