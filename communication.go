package main

import (
	"chatTUIv2_0/protocol"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

const (
	_port = "127.0.0.1:8080"
	_sep  = "|"
)

type messageIDsSlice struct {
	IDs []string
}

func (messages *messageIDsSlice) remove(id string) {
	filteredMessageIDs := messages.IDs[:0]
	for _, mID := range filteredMessageIDs {
		if mID != id {
			filteredMessageIDs = append(filteredMessageIDs, mID)
		}
	}
	messages.IDs = filteredMessageIDs
}

func (messages *messageIDsSlice) add(id string) {
	messages.IDs = append(messages.IDs, id)
}

var (
	isEnding      = false
	messagesToAck = messageIDsSlice{
		IDs: make([]string, 0)}
)

//FORMATO
//TIPO|USUARIO|...

//EJEMPLOS:
//START|USER
//END|USER
//USERLIST|USER,USER,USER...
//MSG|ID_MSG|USER|CONTENIDO...
//PRIVMSG|ID_MSG|USER|DEST|CONTENIDO...
//FILE|ID_FILE|USER|NAMEFILE|TAM|PARTS
//ACKFILE|ID_FILE|NAMEFILE|TAM|PARTS
//PARTFILE|ID_FILE|USER/ALL|NAMEFILE|NO.PART|T.PARTS|CONTENIDO...
//REQFILE|ID_FILE|USER

func processCommunication(username string, sendChan chan<- protocol.MessageCommunication, receiveChan <-chan protocol.MessageCommunication) {
	conn, errHandshake := sendHandshake(username)
	if errHandshake != nil {
		sendChan <- protocol.MessageCommunication{
			TypeMessage: "error",
			Content:     "handshake failed:\n" + errHandshake.Error(),
		}
		return
	}
	internCommunication := make(chan protocol.MessageCommunication)
	go handleConnection(conn, internCommunication)
	go receiveCommandsFromUser(username, receiveChan, conn, internCommunication)
	sendMessagesToChatTUI(sendChan, internCommunication)
}

func handleConnection(conn net.Conn, internCommunication chan protocol.MessageCommunication) {
	buffer := make([]byte, 1024)
	for !isEnding {
		conn.SetReadDeadline(time.Time{})
		n, errRecv := conn.Read(buffer)
		if errRecv != nil {
			isEnding = true
			internCommunication <- protocol.MessageCommunication{TypeMessage: "error", Content: errRecv.Error()}
			return
		}
		messageUntrimed := string(buffer[:n])
		typeMsg := strings.SplitN(messageUntrimed, _sep, 2)
		switch typeMsg[0] {
		case "list":
			internCommunication <- protocol.MessageCommunication{
				TypeMessage: "list",
				Content:     typeMsg[1],
			}
		case "msg":
			// ..ID_MSG|USER|CONTENIDO
			messageParts := strings.SplitN(typeMsg[1], _sep, 3)
			internCommunication <- protocol.MessageCommunication{
				TypeMessage: "msg",
				User:        messageParts[1],
				Content:     messageParts[2],
			}
		case "privmsg":
			messageParts := strings.SplitN(typeMsg[1], _sep, 4)
			internCommunication <- protocol.MessageCommunication{
				TypeMessage: "privmsg",
				User:        messageParts[1] + _sep + messageParts[2],
				Content:     messageParts[3],
			}
		}
	}
}

// TODO: Receive messages from internCommunication
func sendMessagesToChatTUI(sendChan chan<- protocol.MessageCommunication, internCommunication <-chan protocol.MessageCommunication) {
	for {
		select {
		case msg := <-internCommunication:
			switch msg.TypeMessage {
			case "list":
				sendChan <- protocol.MessageCommunication{
					TypeMessage: "list",
					Content:     msg.Content,
				}
			case "msg":
				sendChan <- protocol.MessageCommunication{
					TypeMessage: "msg",
					User:        msg.User,
					Content:     msg.Content,
				}
			case "privmsg":
				sendChan <- protocol.MessageCommunication{
					TypeMessage: "privmsg",
					User:        msg.User,
					Content:     msg.Content,
				}
			case "error":
				sendChan <- protocol.MessageCommunication{
					TypeMessage: "error",
					Content:     msg.Content,
				}
				return
			}
		}
	}
}

func receiveCommandsFromUser(username string, receiveChan <-chan protocol.MessageCommunication, conn net.Conn, internCommunication chan<- protocol.MessageCommunication) {
	for {
		select {
		case msg := <-receiveChan:
			switch msg.TypeMessage {
			case "exit":
				sendExitUser(username, conn)
				conn.Close()
				isEnding = true
				return
			case "msg":
				errSendMsg := sendMessage(username, conn, msg.Content)
				if errSendMsg != nil {
					internCommunication <- protocol.MessageCommunication{
						TypeMessage: "error",
						Content:     "Sending message error: " + errSendMsg.Error(),
					}
					isEnding = true
					return
				}
			case "privmsg":
				errSendMsg := sendPrivMessage(msg.User, conn, msg.Content)
				if errSendMsg != nil {
					internCommunication <- protocol.MessageCommunication{
						TypeMessage: "error",
						Content:     "Sending message error: " + errSendMsg.Error(),
					}
					isEnding = true
					return
				}
			}
		}
	}
}

func sendHandshake(username string) (net.Conn, error) {
	conn, errConn := net.Dial("udp", _port)
	if errConn != nil {
		return nil, errConn
	}

	buffer := make([]byte, 1024)

	msg := "handshake" + _sep + username
	msgbyte := []byte(msg)
	errWriteDeadLine := conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
	if errWriteDeadLine != nil {
		return nil, errWriteDeadLine
	}
	_, errSend := conn.Write(msgbyte)
	if errSend != nil {
		return nil, errSend
	}

	for {
		errReadDeadLine := conn.SetReadDeadline(time.Now().Add(time.Second * 2))
		if errReadDeadLine != nil {
			return nil, errReadDeadLine
		}
		n, errRead := conn.Read(buffer)
		if errRead != nil {
			return nil, errRead
		}
		received := string(buffer[:n])
		parts := strings.SplitN(received, _sep, 2)
		if parts[0] != "start" {
			continue
		}
		if parts[1] == username {
			break
		}
	}

	return conn, nil
}

func sendExitUser(username string, conn net.Conn) {
	message := "exit" + _sep + username
	messageByte := []byte(message)
	conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
	_, errSend := conn.Write(messageByte)
	if errSend != nil {
		return
	}
}

func sendMessage(username string, conn net.Conn, content string) error {
	//send
	idMsg := getUniqueId()
	message := "msg" + _sep + idMsg + _sep + username + _sep + content
	messageByte := []byte(message)
	conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
	_, errSend := conn.Write(messageByte)
	if errSend != nil {
		return errSend
	}
	messagesToAck.add(idMsg)
	return nil
}

func sendPrivMessage(usernames string, conn net.Conn, content string) error {
	//send
	idMsg := getUniqueId()
	message := "privmsg" + _sep + idMsg + _sep + usernames + _sep + content
	messageByte := []byte(message)
	conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
	_, errSend := conn.Write(messageByte)
	if errSend != nil {
		return errSend
	}
	messagesToAck.add(idMsg)
	return nil
}

func getUniqueId() string {
	timestamp := time.Now().Format("02012006150405")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomDigits := r.Intn(1000)
	return fmt.Sprintf("%s%03d", timestamp, randomDigits)
}
