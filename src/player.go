package gchat

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"g-chat/src/data"
	"log"
	"net/http"
	"time"

	"go.uber.org/ratelimit"
	"xabbo.b7c.io/nx"
)

const (
	userDataEndpoint                   string        = "https://origins.habbo.com/api/public/users?name=%s"
	minimumTimeBetweenUserDataRequests time.Duration = time.Hour * 12
)

var (
	playersProcessingChannel chan ClientPlayer = make(chan ClientPlayer, 100)
)

type APIPlayer struct {
	UniqueID                    string     `json:"uniqueId"`                    // "hhous-f062f8933a1aad1ca1732c233c9ab275"
	Name                        string     `json:"name"`                        // "billi"
	FigureString                string     `json:"figureString"`                // "hr-700-1.hd-540-1.ch-655-1276.lg-730-1.sh-600-1.ha-0-1"
	Motto                       string     `json:"motto"`                       // ""
	Online                      bool       `json:"online"`                      // true
	LastAccessTime              string     `json:"lastAccessTime"`              // "2024-08-31T15:00:03.000+0000"
	MemberSince                 string     `json:"memberSince"`                 // "2024-08-22T21:26:31.000+0000"
	ProfileVisible              bool       `json:"profileVisible"`              // true
	CurrentLevel                int        `json:"currentLevel"`                // 0
	CurrentLevelCompletePercent int        `json:"currentLevelCompletePercent"` // 0
	TotalExperience             int        `json:"totalExperience"`             // 0
	StarGemCount                int        `json:"starGemCount"`                // 0
	SelectBadges                []APIBadge `json:"selectedBadges"`              // []
	BouncerPlayerID             string     `json:"bouncerPlayerId"`             // null
}

type APIBadge struct {
	BadgeIndex  int    `json:"badgeIndex"`  // 1
	Code        string `json:"code"`        // "HC2"
	Name        string `json:"name"`        // "Habbo Club membership II"
	Description string `json:"description"` // "For 12 months of Habbo Club membership"
}

func (a APIPlayer) toUpdatePlayerUserDataParams(playerID int64) (data.UpdatePlayerUserDataParams, error) {
	updatePlayerUserDataParams := data.UpdatePlayerUserDataParams{
		Playerid: playerID,
	}

	if a.Motto != "" {
		updatePlayerUserDataParams.Motto = sql.NullString{String: a.Motto, Valid: true}
	}
	if a.MemberSince != "" {
		memberSinceTime, err := time.Parse("2006-01-02T15:04:05.000-0700", a.MemberSince)
		if err != nil {
			return updatePlayerUserDataParams, err
		}
		updatePlayerUserDataParams.Membersince = sql.NullTime{Time: memberSinceTime, Valid: true}
	}

	return updatePlayerUserDataParams, nil
}

func playerApiUpdate(player ClientPlayer) {
	dbPlayer, err := queries.GetPlayerByName(context.Background(), player.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("player %s does not exist in db, creating", player.Name)
			dbPlayer, err = queries.CreatePlayer(context.Background(), player.Name)
			if err != nil {
				log.Printf("error creating player in db for: %s: %v\n", player.Name, err)
				return
			}
		} else {
			log.Printf("error getting player from db for: %s: %v\n", player.Name, err)
			return
		}
	}

	var forceUpdate bool

	if myPlayer != nil && player.Index == myPlayer.Index && !dbPlayer.IsMe {
		forceUpdate = true
		dbPlayer, err = queries.UpdatePlayerSetIsMe(context.Background(), dbPlayer.Playerid)
		if err != nil {
			log.Printf("error setting is me in db for: %s: %v\n", player.Name, err)
			return
		}
	} else if myPlayer != nil && player.Index != myPlayer.Index && dbPlayer.IsMe {
		forceUpdate = true
		dbPlayer, err = queries.UpdatePlayerSetNotMe(context.Background(), dbPlayer.Playerid)
		if err != nil {
			log.Printf("error setting not me in db for: %s: %v\n", player.Name, err)
			return
		}
	}

	userDataUpdateRequired := forceUpdate ||
		time.Since(dbPlayer.Userdatalastrequested.Time) > minimumTimeBetweenUserDataRequests
	figureUpdateRequired := forceUpdate ||
		time.Since(dbPlayer.Figurelastrequested.Time) > minimumTimeBetweenPlayerImageRequests
	avatarUpdateRequired := forceUpdate ||
		time.Since(dbPlayer.AvatarLastRequested.Time) > minimumTimeBetweenPlayerImageRequests

	if userDataUpdateRequired {
		err := updateUserData(&dbPlayer)
		if err != nil {
			log.Printf("err updating user data for: %s: %v\n", dbPlayer.Username, err)
			return
		}
		log.Printf("updated user data for: %s successfully\n", player.Name)
	}

	if figureUpdateRequired || avatarUpdateRequired {

		figure, err := getAndUpdateFigure(&dbPlayer, player)
		if err != nil {
			log.Printf("err updating user data for: %s: %v\n", dbPlayer.Username, err)
			return
		}

		// generate player avatar
		if avatarUpdateRequired {
			updatePlayerAvatar(&dbPlayer, figure)
		}

		// get player figure
		if figureUpdateRequired {
			updatePlayerFigure(&dbPlayer, figure)
		}
	}

	addToPlayerListChannel <- AddToPlayerList{dbPlayer}
}

func playersAPIUpdater() {
	log.Println("players api updater initialized")
	rl := ratelimit.New(5) // per second
	for p := range playersProcessingChannel {
		rl.Take()
		go playerApiUpdate(p)
	}
}

func updateUserData(dbPlayer *data.Player) error {
	apiPlayer, err := requestUserData(dbPlayer.Username)
	if err != nil {
		log.Printf("err requesting user data for %s: %v\n", dbPlayer.Username, err)
		return err
	}

	updatePlayerUserDataParams, err := apiPlayer.toUpdatePlayerUserDataParams(dbPlayer.Playerid)

	if err != nil {
		log.Printf("err generating update player user data params for: %s: %v\n", dbPlayer.Username, err)
		return err
	}

	updatedDBPlayer, err := queries.UpdatePlayerUserData(context.Background(), updatePlayerUserDataParams)
	dbPlayer = &updatedDBPlayer

	if err != nil {
		log.Printf("err updating player user data for: %s: %v\n", dbPlayer.Username, err)
		return err
	}

	return nil
}

func getAndUpdateFigure(dbPlayer *data.Player, player ClientPlayer) (nx.Figure, error) {

	figure, err := convertToFigure(player)
	if err != nil {
		log.Printf("err converting figure for %s: %v\n", dbPlayer.Username, err)
		return nx.Figure{}, err
	}

	updatedDBPlayer, err := queries.UpdatePlayerFigureString(context.Background(), data.UpdatePlayerFigureStringParams{
		Figurestring: sql.NullString{String: figure.String(), Valid: true},
		Playerid:     dbPlayer.Playerid,
	})
	dbPlayer = &updatedDBPlayer

	if err != nil {
		log.Printf("error updating player figure string in db for: %s: %v\n", dbPlayer.Username, err)
		return nx.Figure{}, err
	}

	return figure, nil
}

func updatePlayerAvatar(dbPlayer *data.Player, figure nx.Figure) {
	err := generatePlayerImage(*dbPlayer, figure, Avatar)
	if err != nil {
		log.Printf("err generating player avatar for: %s: %v\n", dbPlayer.Username, err)
		return
	}

	updatedDBPlayer, err := queries.UpdatePlayerAvatar(context.Background(), dbPlayer.Playerid)
	dbPlayer = &updatedDBPlayer
	if err != nil {
		log.Printf("error updating player avatar in db for: %s: %v\n", dbPlayer.Username, err)
		return
	}

	log.Printf("updated avatar for: %s successfully\n", dbPlayer.Username)
}

func updatePlayerFigure(dbPlayer *data.Player, figure nx.Figure) {
	err := generatePlayerImage(*dbPlayer, figure, Figure)
	if err != nil {
		log.Printf("err generating player figure for: %s: %v\n", dbPlayer.Username, err)
		return
	}

	updatedDBPlayer, err := queries.UpdatePlayerFigure(context.Background(), dbPlayer.Playerid)
	dbPlayer = &updatedDBPlayer
	if err != nil {
		log.Printf("error updating player figure in db for: %s: %v\n", dbPlayer.Username, err)
		return
	}

	log.Printf("updated figure for: %s successfully\n", dbPlayer.Username)
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

func getDBPlayers(clientPlayers map[int]*ClientPlayer) ([]data.Player, error) {
	if len(clientPlayers) == 0 {
		return []data.Player{}, nil
	}

	var usernames []string

	for _, player := range clientPlayers {
		usernames = append(usernames, player.Name)
	}

	dataPlayers, err := queries.ListPlayersByUsernames(context.Background(), usernames)
	if err != nil {
		return []data.Player{}, err
	}

	return dataPlayers, nil
}
