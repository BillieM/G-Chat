package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	g "xabbo.b7c.io/goearth"
	in "xabbo.b7c.io/goearth/shockwave/in"
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
)

type ChatMessage struct {
	Username string
	Gender   string
	Colour   string
	Content  string
	Time     string
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
}

func onConnected(e g.ConnectArgs) {
	log.Printf("Game connected (%s)\n", e.Host)
	go initWebServer()
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

	tmpl := template.Must(template.ParseFiles("templates/msg.html"))
	var buf bytes.Buffer
	tmpl.Execute(&buf, ChatMessage{
		Username: user.Name,
		Gender:   genders[user.Gender],
		Content:  base64.StdEncoding.EncodeToString([]byte(msg)), // encode msg
		Time:     time.Now().Format("15:04"),
	})

	for _, conn := range activeConnections {
		err := conn.WriteMessage(1, buf.Bytes())
		if err != nil {
			log.Println(fmt.Errorf(
				"error with msg: %s, %w", msg, err,
			))
		}
	}

	log.Printf("msg successful %s: %s\n", user.Name, msg)
	// log.Printf("name:%s, custom: %s, figure: %s, gender: %s, pool figure: %s, type: %v, x: %v, y: %v, z: %v, badgecode: %s\n", user.Name, user.Custom, user.Figure, user.Gender, user.PoolFigure, user.Type, user.X, user.Y, user.Z, user.BadgeCode)
}

// web server code
func initWebServer() {
	http.HandleFunc("/", homeFunc)
	http.HandleFunc("/chat", chatFunc)

	http.ListenAndServe(":8080", nil)
}

func homeFunc(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func chatFunc(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	connUuid := uuid.NewString()
	activeConnections[connUuid] = c
	defer func() {
		delete(activeConnections, connUuid)
		c.Close()
	}()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "close 1001") {
				break
			}

			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
