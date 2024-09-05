package gchat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// web server code
func initWebServer() {
	// pages
	http.HandleFunc("/", homeFunc)
	http.HandleFunc("/settings", settingsFunc)

	// components
	http.HandleFunc("/playercard/{username}", playerCardFunc)

	// websocket connect
	http.HandleFunc("/chat", chatFunc)

	// static
	http.HandleFunc("/static/", staticFunc)

	log.Printf("starting webserver on port: %v\n", webServerPort)
	http.ListenAndServe(fmt.Sprintf(":%v", webServerPort), nil)
}

func homeFunc(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	serveTemplate(w, indexTemplate, nil)
}

func settingsFunc(w http.ResponseWriter, r *http.Request) {

	var backgroundColourBuffer bytes.Buffer
	var textColourBuffer bytes.Buffer

	serveTemplate(&backgroundColourBuffer, colourPickerTemplate, config.BackgroundColours)
	serveTemplate(&textColourBuffer, colourPickerTemplate, config.TextColours)

	serveTemplate(w, settingsTemplate, map[string]any{
		"Config": config,
		"Message": map[string]any{
			"Time":    time.Now(),
			"Gender":  "♀",
			"Content": "Example message!",
		},
		"BackgroundColourPicker": template.HTML(backgroundColourBuffer.String()),
		"TextColourPicker":       template.HTML(textColourBuffer.String()),
	})
}

func playerCardFunc(w http.ResponseWriter, r *http.Request) {

	username := r.PathValue("username")

	dbPlayer, err := queries.GetPlayerByName(r.Context(), username)
	if err != nil {
		log.Printf("error getting player: %s from db for player card: %v\n", username, err)
		serveTemplate(w, playerCardTemplate, nil)
		return
	}

	serveTemplate(w, playerCardTemplate, map[string]any{
		"Username":     dbPlayer.Username,
		"Motto":        dbPlayer.Motto.String,
		"Figureexists": dbPlayer.Figureexists.Bool,
	})
}

func chatFunc(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	connUuid := uuid.NewString()
	activeConnections.Store(connUuid, c)
	log.Printf("connection %s opened\n", connUuid)
	defer func() {
		activeConnections.Delete(connUuid)
		c.Close()
		log.Printf("connection %s closed\n", connUuid)
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
		var messageToSendToHabbo MessageToSendToHabbo
		json.Unmarshal(message, &messageToSendToHabbo)
		handleSendHabboChat(messageToSendToHabbo)
	}
}

func staticFunc(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "js") {
		w.Header().Set("Content-Type", "text/javascript")
	} else if strings.HasSuffix(path, "css") {
		w.Header().Set("Content-Type", "text/css")
	}
	data, err := os.ReadFile(path[1:])
	if err != nil {
		fmt.Print(err)
	}
	_, err = w.Write(data)
	if err != nil {
		fmt.Print(err)
	}
}
