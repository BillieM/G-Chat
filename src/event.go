package gchat

import (
	"bytes"
	"fmt"
	"log"
)

type EventType string

const (
	Notification EventType = "notification"
	Message      EventType = "message"
)

type WebSocketMessage struct {
	Chat        string      `json:"chat"`
	MessageType MessageType `json:"type"`
}

func sendEvent(eventType EventType, data any) {
	var buf bytes.Buffer
	serveComponentTemplate(&buf, string(eventType), data)

	for _, conn := range activeConnections {
		err := conn.WriteMessage(1, buf.Bytes())
		if err != nil {
			log.Println(fmt.Errorf(
				"error with %s: %w", eventType, err,
			))
		}
	}

	log.Printf("event: %s successfully sent\n", eventType)
}