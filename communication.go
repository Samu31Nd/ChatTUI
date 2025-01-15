package main

import (
	"chatTUIv2_0/protocol"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	_port = "127.0.0.1:8080"
	_sep  = "|"
)

var (
	endOfFile = false
)

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

func processCommunication(username string, sendChan chan<- protocol.MessageCommunication, receiveChan <-chan protocol.MessageCommunication) {
	counter := 0
	conn, errHandshake := sendHandshake(username)
	if errHandshake != nil {
		sendChan <- protocol.MessageCommunication{
			TypeMessage: "Error",
			Content:     "handshake failed:\n" + errHandshake.Error(),
		}
		return
	}
	go receiveCommands(username, receiveChan, conn)
	for !endOfFile {
		if counter%2 == 1 {
			sendChan <- protocol.MessageCommunication{
				TypeMessage: "File",
				User:        "Admin",
				IdOptional:  12,
				Content:     "file.txt,32000",
			}
		}
		counter++
		time.Sleep(time.Second * 1)
		sendChan <- protocol.MessageCommunication{
			TypeMessage: "List",
			Content:     "Admin,Elisa,Samu",
		}
		sendChan <- protocol.MessageCommunication{
			TypeMessage: "Msg",
			User:        "Admin",
			Content:     fmt.Sprintf("Echo Admin! %d", counter),
		}
	}
}

func receiveCommands(username string, receiveChan <-chan protocol.MessageCommunication, conn net.Conn) {
	for {
		select {
		case msg := <-receiveChan:
			switch msg.TypeMessage {
			case "exit":
				sendExitUser(username, conn)
				conn.Close()
				endOfFile = false
				return
			case "msg":
				sendMessage(username, conn, msg.Content)
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

func sendMessage(username string, conn net.Conn, content string) {
	message := "msg" + _sep + username + _sep + content
	messageByte := []byte(message)
	conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
	_, errSend := conn.Write(messageByte)
	if errSend != nil {
		return
	}
}
