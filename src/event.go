package gchat

import (
	"bytes"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type EventType string

const (
	Notification EventType = "notification"
	Message      EventType = "message"
)

type WebMessage struct {
	Chat        string      `json:"chat"`
	MessageType MessageType `json:"type"`
}

func sendEvent(eventType EventType, data any) {
	var buf bytes.Buffer
	serveComponentTemplate(&buf, string(eventType), data)

	activeConnections.Range(func(_ any, val any) bool {
		conn := val.(*websocket.Conn)
		err := conn.WriteMessage(1, buf.Bytes())
		if err != nil {
			log.Println(fmt.Errorf(
				"error with %s: %w", eventType, err,
			))
		}
		return true
	})

	log.Printf("event: %s successfully sent to browser\n", eventType)
}
