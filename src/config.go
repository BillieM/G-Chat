package gchat

import (
	"database/sql"
	"encoding/json"
	"g-chat/src/data"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/nx/gamedata/origins"

	_ "github.com/mattn/go-sqlite3"
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
	activeConnections  sync.Map
	upgrader           = websocket.Upgrader{}
	myPlayer           *ClientPlayer
	players                              = map[int]*ClientPlayer{}
	otherEntities                        = map[int]string{}
	playersPacketCount                   = 0
	genders            map[string]string = map[string]string{
		"f":       "♀",
		"m":       "♂",
		"unknown": "",
	}
	config                 Config
	queries                *data.Queries
	host                   string = "origins.habbo.com"
	figureData             *origins.FigureData
	figureDataLastObtained time.Time
)

type Config struct {
	AvailableColours  map[string]ColourPair `json:"availableColours"`
	PlayerColours     map[string]ColourPair `json:"playerColours"`
	BackgroundColours []string              `json:"backgroundColours"`
	TextColours       []string              `json:"textColours"`
}

type ColourPair struct {
	BackgroundColour string `json:"backgroundColour"`
	TextColour       string `json:"textColour"`
}

func createDirs() {
	neededDirs := []string{
		"static/images/figures",
		"static/images/avatars",
		"static/images/avatars/me",
	}

	for _, dir := range neededDirs {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Panicf("error creating dir: %s: %v\n", dir, err)
		}
	}
}

func loadConfig() {
	data, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &config)
}

func loadDB() error {
	db, err := sql.Open("sqlite3", "./db/app.db")
	if err != nil {
		return err
	}

	queries = data.New(db)

	return nil
}
