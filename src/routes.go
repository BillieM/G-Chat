package gchat

import (
	"bytes"
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

	servePageTemplate(w, nil, "index")
}

func settingsFunc(w http.ResponseWriter, r *http.Request) {

	var backgroundColourBuffer bytes.Buffer
	var textColourBuffer bytes.Buffer

	serveComponentTemplate(&backgroundColourBuffer, "colourpicker", config.BackgroundColours)
	serveComponentTemplate(&textColourBuffer, "colourpicker", config.TextColours)

	servePageTemplate(w, map[string]any{
		"Config": config,
		"Message": map[string]any{
			"Time":     time.Now(),
			"Gender":   "♀",
			"Username": config.PlayerUsername,
			"Content":  "Example message!",
		},
		"BackgroundColourPicker": template.HTML(backgroundColourBuffer.String()),
		"TextColourPicker":       template.HTML(textColourBuffer.String()),
	}, "settings", "message")
}

func chatFunc(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	connUuid := uuid.NewString()
	activeConnections[connUuid] = c
	log.Printf("connection %s opened\n", connUuid)
	defer func() {
		delete(activeConnections, connUuid)
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
	}
}

func staticFunc(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "js") {
		w.Header().Set("Content-Type", "text/javascript")
	} else {
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
