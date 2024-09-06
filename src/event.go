package gchat

import (
	"bytes"
	"fmt"
	"g-chat/src/data"
	"log"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	Shout   MessageType = "shout"
	Say     MessageType = "say"
	Whisper MessageType = "whisper"
)

type MessageToSendToHabbo struct {
	Chat        string      `json:"chat"`
	MessageType MessageType `json:"type"`
}

type Message struct {
	Username       string
	Gender         string
	ChatBackground string
	ChatText       string
	Content        string
	EncodedContent string
	Time           string
	MessageType    MessageType
	FigureExists   bool
	AvatarExists   bool
	FromMe         bool
}

type AddToPlayerList struct {
	data.Player
}

type ClearPlayerList struct{}

type PlayerLeaveRoom struct {
	Username string
}

type PlayerEnterRoom struct {
	Username string
}

var (
	messageChannel         chan Message         = make(chan Message)
	playerLeaveRoomChannel chan PlayerLeaveRoom = make(chan PlayerLeaveRoom)
	playerEnterRoomChannel chan PlayerEnterRoom = make(chan PlayerEnterRoom)
	addToPlayerListChannel chan AddToPlayerList = make(chan AddToPlayerList)
	clearPlayerListChannel chan ClearPlayerList = make(chan ClearPlayerList)
)

func socketEventSender() {
	log.Println("socket event sender initialized")
	for {
		select {
		case e := <-messageChannel:
			sendEvent(messageTemplate, e)
		case e := <-playerLeaveRoomChannel:
			sendEvent(exitNotificationTemplate, e)
			sendEvent(removeFromPlayerListTemplate, e)
		case e := <-playerEnterRoomChannel:
			sendEvent(enterNotificationTemplate, e)
		case e := <-addToPlayerListChannel:
			sendEvent(addToPlayerListTemplate, e)
		case e := <-clearPlayerListChannel:
			sendEvent(clearPlayerListTemplate, e)
		}
	}
}

func sendEvent(template appTemplate, data any) {

	var buf bytes.Buffer
	serveTemplate(&buf, template, data)

	activeConnections.Range(func(_ any, val any) bool {
		conn := val.(*websocket.Conn)
		err := conn.WriteMessage(1, buf.Bytes())
		if err != nil {
			log.Println(fmt.Errorf(
				"error with %s event: %v", template.templateName, err,
			))
		}
		return true
	})

	log.Printf("event: %s successfully sent to browser\n", template.templateName)
}

/*
theory

hx-swap-oob="delete" -> LeaveRoom event

*/
