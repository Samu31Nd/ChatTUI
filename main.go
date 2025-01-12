package main

import (
	"chatTUIv2_0/ChatroomTUI"
	"chatTUIv2_0/FormLogginTUI"
	"chatTUIv2_0/protocol"
)

func main() {
	username := FormLogginTUI.GetUsername()
	if username == "" {
		return
	}
	//to receive messages in TUI
	//and send messages in communication
	channelCommunicationClient := make(chan protocol.MessageCommunication)
	//to send messages in TUI
	//and receive messages in communication
	channelCommunicationServer := make(chan protocol.MessageCommunication)
	//PORT
	go processCommunication(channelCommunicationClient, channelCommunicationServer)
	//APP
	ChatroomTUI.StartChatSession(username, channelCommunicationServer, channelCommunicationClient)
}
