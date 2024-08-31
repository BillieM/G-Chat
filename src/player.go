package gchat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/ratelimit"
)

const (
	figureSize                         string        = "l"
	userDataEndpoint                   string        = "https://origins.habbo.com/api/public/users?name=%s"
	figureEndpoint                     string        = "https://www.habbo.com/habbo-imaging/avatarimage?size=%s&figure=%s"
	minimumTimeBetweenUserDataRequests time.Duration = time.Hour
	minimumTimeBetweenFigureRequests   time.Duration = time.Hour
)

type APIPlayer struct {
	UniqueID                    string `json:"uniqueId"`                    // "hhous-f062f8933a1aad1ca1732c233c9ab275"
	Name                        string `json:"name"`                        // "billi"
	FigureString                string `json:"figureString"`                // "hr-700-1.hd-540-1.ch-655-1276.lg-730-1.sh-600-1.ha-0-1"
	Motto                       string `json:"motto"`                       // ""
	Online                      bool   `json:"online"`                      // true
	LastAccessTime              string `json:"lastAccessTime"`              // "2024-08-31T15:00:03.000+0000"
	MemberSince                 string `json:"memberSince"`                 // "2024-08-22T21:26:31.000+0000"
	ProfileVisible              bool   `json:"profileVisible"`              // true
	CurrentLevel                int    `json:"currentLevel"`                // 0
	CurrentLevelCompletePercent int    `json:"currentLevelCompletePercent"` // 0
	TotalExperience             int    `json:"totalExperience"`             // 0
	StarGemCount                int    `json:"starGemCount"`                // 0
	SelectBadges                []int  `json:"selectedBadges"`              // []
	BouncerPlayerID             string `json:"bouncerPlayerId"`             // null
}

func playerApiUpdate(player ClientPlayer) {
	dbPlayer, err := queries.GetPlayerByName(context.Background(), player.Name)
	if err != nil {
		log.Println("error getting player: %e", err)
	}

	timeLast

	if time.Since()
}

func playersApiUpdate(players []ClientPlayer) {
	rl := ratelimit.New(5) // per second

	for _, playerName := range players {
		rl.Take()
		go playerApiUpdate(playerName)
	}
}

func requestUserData(playerName string) (APIPlayer, error) {
	requestUrl := fmt.Sprintf(userDataEndpoint, playerName)
	var apiPlayer APIPlayer

	resp, err := http.Get(requestUrl)
	if err != nil {
		return apiPlayer, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&apiPlayer)
	if err != nil {
		return apiPlayer, err
	}

	return apiPlayer, nil
}

func requestFigure(apiPlayer APIPlayer) error {
	requestUrl := fmt.Sprintf(figureEndpoint, figureSize, apiPlayer.Name)

	resp, err := http.Get(requestUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("static/avatars/%s.png", apiPlayer.Name)

	err = os.WriteFile(filePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}
