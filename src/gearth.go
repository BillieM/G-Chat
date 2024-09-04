package gchat

import (
	"context"
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
	FigureExists   bool
	AvatarExists   bool
	FromMe         bool
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
	ext.Initialized(handleInitialized)
	ext.Connected(handleConnected)
	ext.Disconnected(handleDisconnected)
	ext.Intercept(in.CHAT).With(handleReceiveHabboSay)
	ext.Intercept(in.CHAT_2).With(handleReceiveHabboWhisper)
	ext.Intercept(in.CHAT_3).With(handleReceiveHabboShout)
	ext.Intercept(in.OPC_OK).With(handleHabboEnterRoom)
	ext.Intercept(in.USERS).With(handleHabboUsers)
	ext.Intercept(in.LOGOUT).With(handleHabboRemoveUser)

	createDirs()
	loadConfig()
	err := loadDB()
	if err != nil {
		log.Panicf("error loading database: %e", err)
	}
	go initWebServer()

	ext.Run()
}

func handleInitialized(e g.InitArgs) {
	log.Println("G-Chat initialized")
}

func handleConnected(e g.ConnectArgs) {
	log.Printf("Game connected (%s)\n", e.Host)
	ext.Send(out.GETFLATINFO)
	ext.Send(out.GETINTERST)
	ext.Send(out.GETROOMAD)
	ext.Send(out.G_HMAP)
	ext.Send(out.G_USRS)
	ext.Send(out.G_STAT)
}

func handleDisconnected() {
	log.Println("Game disconnected")
}

func handleHabboEnterRoom(e *g.Intercept) {
	playersPacketCount = 0
	clear(players)
}

func handleHabboRemoveUser(e *g.Intercept) {
	s := e.Packet.ReadString()
	index, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	if player, ok := players[index]; ok {
		go sendEvent(Notification, map[string]string{
			"Content": fmt.Sprintf("%s left the room!", player.Name),
		})
		log.Printf("* %s left the room.", player.Name)
		delete(players, index)
	}
}

func handleHabboUsers(e *g.Intercept) {
	// Observations:
	// The first USERS packet sent upon entering the room (after OPC_OK)
	// is the list of users that are already in the room.
	// The second USERS packet contains a single user, yourself.
	// The following USERS packets indicate someone entering the room.
	var newPlayers []ClientPlayer
	playersPacketCount++
	for range e.Packet.ReadInt() {
		var player ClientPlayer
		e.Packet.Read(&player)
		if player.Type == 1 {

			if playersPacketCount == 2 {
				myPlayer = &player
			}

			if playersPacketCount >= 3 {
				go sendEvent(Notification, map[string]string{
					"Content": fmt.Sprintf("%s entered the room!", player.Name),
				})
				log.Printf("* %s entered the room\n", player.Name)
			}

			players[player.Index] = &player
			newPlayers = append(newPlayers, player)
		} else {
			otherEntities[player.Index] = player.Name
			if player.Type != 2 {
				log.Printf("non 1 player 1 or 2 type, name: %s, type: %v\n", player.Name, player.Type)
			}
		}
		log.Printf("user packet: %v for user: %s\n", playersPacketCount, player.Name)
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

	player, ok := players[index]

	if !ok {
		otherEntity, ok := otherEntities[index]
		if ok {
			log.Printf("message from other entity (likely pet), ignoring: %s", otherEntity)
			return
		}

		log.Println("error finding user, falling back to default")
		player = &ClientPlayer{
			Name:   "unknown",
			Gender: "unknown",
		}
	}

	colourPair := getUserColours(player.Name)

	clientMessage := ClientMessage{
		Username:       player.Name,
		Gender:         genders[player.Gender],
		ChatBackground: colourPair.BackgroundColour,
		ChatText:       colourPair.TextColour,
		EncodedContent: base64.StdEncoding.EncodeToString([]byte(msg)), // encode msg
		Time:           time.Now().Format("15:04"),
		MessageType:    messageType,
	}

	dbPlayer, err := queries.GetPlayerByName(context.Background(), player.Name)
	if err != nil {
		log.Printf("error getting player from db for: %s: %v\n", player.Name, err)
	} else {
		clientMessage.AvatarExists = dbPlayer.AvatarExists.Bool
		clientMessage.FigureExists = dbPlayer.Figureexists.Bool
	}

	sendEvent(Message, clientMessage)
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
