package gchat

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/in"
)

type ChatMessage struct {
	Username       string
	Gender         string
	ChatBackground string
	ChatText       string
	Content        string
	Time           string
}

type HabboUser struct {
	Index      int
	Name       string
	Figure     string
	Gender     string
	Custom     string
	X, Y       int
	Z          float64
	PoolFigure string
	BadgeCode  string
	Type       int
}

func InitExt() {
	ext.Initialized(onInitialized)
	ext.Connected(onConnected)
	ext.Disconnected(onDisconnected)
	ext.Intercept(in.CHAT, in.CHAT_2, in.CHAT_3).With(handleHabboChat)
	ext.Intercept(in.OPC_OK).With(handleHabboEnterRoom)
	ext.Intercept(in.USERS).With(handleHabboUsers)
	ext.Intercept(in.LOGOUT).With(handleHabboRemoveUser)
	ext.Run()
}

func onInitialized(e g.InitArgs) {
	log.Println("G-Chat initialized")
	loadConfig()
	go initWebServer()
}

func onConnected(e g.ConnectArgs) {
	log.Printf("Game connected (%s)\n", e.Host)
}

func onDisconnected() {
	log.Println("Game disconnected")
}

func handleHabboEnterRoom(e *g.Intercept) {
	usersPacketCount = 0
	clear(users)
}

func handleHabboRemoveUser(e *g.Intercept) {
	s := e.Packet.ReadString()
	index, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	if user, ok := users[index]; ok {
		go sendEvent(Notification, map[string]string{
			"Content": fmt.Sprintf("%s left the room!", user.Name),
		})
		log.Printf("* %s left the room.", user.Name)
		delete(users, index)
	}
}

func handleHabboUsers(e *g.Intercept) {
	// Observations:
	// The first USERS packet sent upon entering the room (after OPC_OK)
	// is the list of users that are already in the room.
	// The second USERS packet contains a single user, yourself.
	// The following USERS packets indicate someone entering the room.
	usersPacketCount++
	for range e.Packet.ReadInt() {
		var user HabboUser
		e.Packet.Read(&user)
		if user.Type == 1 {
			if usersPacketCount >= 3 {
				go sendEvent(Notification, map[string]string{
					"Content": fmt.Sprintf("%s entered the room!", user.Name),
				})
				log.Printf("* %s entered the room\n", user.Name)
			}
			users[user.Index] = &user
		}
	}
}

func handleHabboChat(e *g.Intercept) {
	index := e.Packet.ReadInt() // skip entity index
	msg := e.Packet.ReadString()

	user, ok := users[index]

	if !ok {
		log.Println("error finding user, falling back to default")
		user = &HabboUser{
			Name:   "unknown",
			Gender: "unknown",
		}
	}

	colourPair := getUserColours(user.Name)

	go sendEvent(Message, ChatMessage{
		Username:       user.Name,
		Gender:         genders[user.Gender],
		ChatBackground: colourPair.BackgroundColour,
		ChatText:       colourPair.TextColour,
		Content:        base64.StdEncoding.EncodeToString([]byte(msg)), // encode msg
		Time:           time.Now().Format("15:04"),
	})
}
