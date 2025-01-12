package main

import (
	"chatTUIv2_0/protocol"
	"fmt"
	"time"
)

func processCommunication(sendChan chan<- protocol.MessageCommunication, receiveChan <-chan protocol.MessageCommunication) {
	counter := 0
	for {
		if counter == 2 {
			sendChan <- protocol.MessageCommunication{
				TypeMessage: "File",
				User:        "Admin",
				IdOptional:  12,
				Content:     "file.txt,32000",
			}
		}
		counter++
		time.Sleep(time.Second * 2)
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
