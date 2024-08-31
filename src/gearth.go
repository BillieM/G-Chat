package gchat

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
)

type MessageType string

const (
	Shout   MessageType = "shout"
	Say     MessageType = "say"
	Whisper MessageType = "whisper"
)

type ClientMessage struct {
	Username       string
	Gender         string
	ChatBackground string
	ChatText       string
	Content        string
	EncodedContent string
	Time           string
	MessageType    MessageType
}

type ClientPlayer struct {
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
	ext.Intercept(in.CHAT).With(handleReceiveHabboSay)
	ext.Intercept(in.CHAT_2).With(handleReceiveHabboWhisper)
	ext.Intercept(in.CHAT_3).With(handleReceiveHabboShout)
	ext.Intercept(in.OPC_OK).With(handleHabboEnterRoom)
	ext.Intercept(in.USERS).With(handleHabboUsers)
	ext.Intercept(in.LOGOUT).With(handleHabboRemoveUser)
	ext.Intercept(out.WHISPER).With(handleWhisperTest)
	ext.Run()
}

func handleWhisperTest(e *g.Intercept) {
	fmt.Printf("intercept: %#v, packet: %#v, packet data: %s\n", e, e.Packet, string(e.Packet.Data))
}

func onInitialized(e g.InitArgs) {
	log.Println("G-Chat initialized")
	loadConfig()
	err := loadDB()
	if err != nil {
		log.Panicf("error loading database: %e", err)
	}
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
	var newPlayers []ClientPlayer
	usersPacketCount++
	for range e.Packet.ReadInt() {
		var player ClientPlayer
		e.Packet.Read(&player)
		if player.Type == 1 {
			if usersPacketCount >= 3 {
				go sendEvent(Notification, map[string]string{
					"Content": fmt.Sprintf("%s entered the room!", player.Name),
				})
				log.Printf("* %s entered the room\n", player.Name)
			}
			users[player.Index] = &player
			newPlayers = append(newPlayers, player)
		}
	}
	go playersApiUpdate(newPlayers)
}

func handleReceiveHabboSay(e *g.Intercept) {
	handleReceiveHabboChat(e, Say)
}

func handleReceiveHabboWhisper(e *g.Intercept) {
	handleReceiveHabboChat(e, Whisper)
}

func handleReceiveHabboShout(e *g.Intercept) {
	handleReceiveHabboChat(e, Shout)
}

func handleReceiveHabboChat(e *g.Intercept, messageType MessageType) {
	index := e.Packet.ReadInt() // skip entity index
	msg := e.Packet.ReadString()

	user, ok := users[index]

	if !ok {
		log.Println("error finding user, falling back to default")
		user = &ClientPlayer{
			Name:   "unknown",
			Gender: "unknown",
		}
	}

	colourPair := getUserColours(user.Name)

	fmt.Println(user.Figure)

	go sendEvent(Message, ClientMessage{
		Username:       user.Name,
		Gender:         genders[user.Gender],
		ChatBackground: colourPair.BackgroundColour,
		ChatText:       colourPair.TextColour,
		EncodedContent: base64.StdEncoding.EncodeToString([]byte(msg)), // encode msg
		Time:           time.Now().Format("15:04"),
		MessageType:    messageType,
	})
}

func handleSendHabboChat(data WebMessage) {
	switch data.MessageType {
	case Say:
		ext.Send(out.CHAT, data.Chat)
	case Whisper:
		ext.Send(out.WHISPER, data.Chat)
	case Shout:
		ext.Send(out.SHOUT, data.Chat)
	default:
		log.Println("invalid message type???")
	}
}
