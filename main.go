package main

import (
	"github.com/gorilla/websocket"
	g "xabbo.b7c.io/goearth"
)

/*
to implement

message sending
	- don't create 2 instances of a message sent by me
		- whether sent from web client or habbo itself
	- ability to choose whisper/ say/ shout
		- default to shout

shout/ whisper/ say received differentiation

avatars
	- clicking on avatar allows you to pick

mutexes to ensure concurrency stability

ability to view badges/ mottos in chat client

ability to assign a colour to a user from chat client
	- opens when clicking avatar
	- colour picker section
	- requests all available colours from backend
		- can then select the one you want

colour scheme creator
	- separate page, displays all colours
	- ability to add new, remove existing, or edit existing
	- ui:
		- (-) [rose-500] background: [colour slider] text: [colour slider] -> how it looks: [example message]
		- [Save]

*/

const (
	webServerPort = 8080
)

var (
	ext = g.NewExt(g.ExtInfo{
		Title:       "G-Chat",
		Description: "A web based habbo chat client",
		Author:      "Billie Merz",
		Version:     "1.0",
	})
	activeConnections map[string]*websocket.Conn = make(map[string]*websocket.Conn)
	upgrader                                     = websocket.Upgrader{}
	users                                        = map[int]*HabboUser{}
	usersPacketCount                             = 0
	genders           map[string]string          = map[string]string{
		"f":       "♀",
		"m":       "♂",
		"unknown": "",
	}
	config Config
)

type Config struct {
	AvailableColours map[string]ColourPair `json:"availableColours"`
	PlayerColours    map[string]ColourPair `json:"playerColours"`
}

type ColourPair struct {
	BackgroundColour string `json:"backgroundColour"`
	TextColour       string `json:"textColour"`
}

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

func main() {
	initExt()
}
