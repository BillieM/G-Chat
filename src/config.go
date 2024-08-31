package gchat

import (
	"encoding/json"
	"os"

	"github.com/gorilla/websocket"
	g "xabbo.b7c.io/goearth"
)

const (
	webServerPort = 8080
)

var (
	ext = g.NewExt(g.ExtInfo{
		Title:       "G-Chat",
		Description: "A web based habbo chat client",
		Author:      "Billie M",
		Version:     "0.2",
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
	AvailableColours  map[string]ColourPair `json:"availableColours"`
	PlayerColours     map[string]ColourPair `json:"playerColours"`
	BackgroundColours []string              `json:"backgroundColours"`
	TextColours       []string              `json:"textColours"`
	PlayerUsername    string                `json:"playerUsername"`
}

type ColourPair struct {
	BackgroundColour string `json:"backgroundColour"`
	TextColour       string `json:"textColour"`
}

func loadConfig() {
	data, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &config)
}
