package gchat

import (
	"bytes"
	"fmt"
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

type ILeaveRoom struct {
	Username string
}

type IEnterRoom struct {
	Username string
}

type OtherLeaveRoom struct {
	Username string
}

type OtherEnterRoom struct {
	Username string
}

var (
	messageChannel        chan Message        = make(chan Message)
	iLeaveRoomChannel     chan ILeaveRoom     = make(chan ILeaveRoom)
	iEnterRoomChannel     chan IEnterRoom     = make(chan IEnterRoom)
	otherLeaveRoomChannel chan OtherLeaveRoom = make(chan OtherLeaveRoom)
	otherEnterRoomChannel chan OtherEnterRoom = make(chan OtherEnterRoom)
)

func socketEventSender() {
	log.Println("socket event sender initialized")
	for {
		select {
		case e := <-messageChannel:
			sendEvent(messageTemplate, e)
		case e := <-iLeaveRoomChannel:
			sendEvent(appTemplate{}, e)
		case e := <-iEnterRoomChannel:
			sendEvent(appTemplate{}, e)
		case e := <-otherLeaveRoomChannel:
			sendEvent(otherLeaveRoomTemplate, e)
		case e := <-otherEnterRoomChannel:
			sendEvent(otherEnterRoomTemplate, e)

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
