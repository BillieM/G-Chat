package gchat

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"g-chat/src/data"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/ratelimit"
	"xabbo.b7c.io/nx"
	"xabbo.b7c.io/nx/cmd/nx/util"
	gd "xabbo.b7c.io/nx/gamedata"
	"xabbo.b7c.io/nx/gamedata/origins"
	"xabbo.b7c.io/nx/imager"
)

const (
	figureSize                            string        = "l"
	userDataEndpoint                      string        = "https://origins.habbo.com/api/public/users?name=%s"
	figureEndpoint                        string        = "https://www.habbo.com/habbo-imaging/avatarimage?size=%s&figure=%s"
	minimumTimeBetweenUserDataRequests    time.Duration = time.Hour * 12
	minimumTimeBetweenPlayerImageRequests time.Duration = time.Hour * 12
	figureDataCacheTime                   time.Duration = time.Hour * 4
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

type PlayerImageType string

const (
	Figure PlayerImageType = "figure"
	Avatar PlayerImageType = "avatar"
)

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

	if time.Since(dbPlayer.Userdatalastrequested.Time) > minimumTimeBetweenUserDataRequests {
		apiPlayer, err := requestUserData(player.Name)
		if err != nil {
			log.Printf("error requesting user data for %s: %v\n", player.Name, err)
			return
		}

		log.Printf("requested user data for: %s successfully\n", player.Name)

		updatePlayerUserDataParams, err := apiPlayer.toUpdatePlayerUserDataParams(dbPlayer.Playerid)

		if err != nil {
			log.Printf("error generating updatePlayerUserDataParams from apiPlayer for: %s: %v\n", player.Name, err)
			return
		}

		dbPlayer, err = queries.UpdatePlayerUserData(context.Background(), updatePlayerUserDataParams)

		if err != nil {
			log.Printf("error updating playerUserData in db for: %s: %v\n", player.Name, err)
			return
		}

		log.Printf("updated user data for: %s successfully\n", player.Name)
	}

	if time.Since(dbPlayer.Figurelastrequested.Time) > minimumTimeBetweenPlayerImageRequests ||
		time.Since(dbPlayer.AvatarLastRequested.Time) > minimumTimeBetweenPlayerImageRequests {

		figure, err := convertToFigure(player)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("converted figure for: %s successfully\n", player.Name)

		dbPlayer, err = queries.UpdatePlayerFigureString(context.Background(), data.UpdatePlayerFigureStringParams{
			Figurestring: sql.NullString{String: figure.String(), Valid: true},
			Playerid:     dbPlayer.Playerid,
		})
		if err != nil {
			log.Printf("error updating player figure string in db for: %s: %v\n", player.Name, err)
			return
		}

		// generate player avatar
		if time.Since(dbPlayer.AvatarLastRequested.Time) > minimumTimeBetweenPlayerImageRequests {
			go func() {
				err = generatePlayerImage(dbPlayer, figure, Avatar)
				if err != nil {
					log.Println(err)
					return
				}

				dbPlayer, err = queries.UpdatePlayerAvatar(context.Background(), dbPlayer.Playerid)
				if err != nil {
					log.Printf("error updating player avatar in db for: %s: %v\n", player.Name, err)
					return
				}

				log.Printf("updated avatar for: %s successfully\n", player.Name)
			}()
		}

		// get player figure
		if time.Since(dbPlayer.Figurelastrequested.Time) > minimumTimeBetweenPlayerImageRequests {
			go func() {
				err = generatePlayerImage(dbPlayer, figure, Figure)
				if err != nil {
					log.Println(err)
					return
				}

				dbPlayer, err = queries.UpdatePlayerFigure(context.Background(), dbPlayer.Playerid)
				if err != nil {
					log.Printf("error updating player figure in db for: %s: %v\n", player.Name, err)
					return
				}

				log.Printf("updated figure for: %s successfully\n", player.Name)
			}()
		}
	}
}

func playersApiUpdate(players []ClientPlayer) {
	rl := ratelimit.New(5) // per second
	for _, playerName := range players {
		rl.Take()
		playerApiUpdate(playerName)
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

/*
generateFigureString turns a numerical origins figure string into a figure string compatible
with the habbo figures api

based on: https://github.com/xabbo/nx/blob/dev/cmd/nx/cmd/figure/convert/convert.go
*/
func convertToFigure(player ClientPlayer) (nx.Figure, error) {
	originsFigure := player.Figure
	if len(originsFigure) != 25 {
		return nx.Figure{}, errors.New("invalid figure string, must be 25 characters in length")
	}

	for _, c := range originsFigure {
		if c < '0' || c > '9' {
			return nx.Figure{}, errors.New("invalid figure string, must consist only of numbers")
		}
	}

	gdm := gd.NewManager("www.habbo.com")

	err := gdm.Load(gd.GameDataFigure)
	if err != nil {
		return nx.Figure{}, fmt.Errorf("failed to load modern figure data: %w", err)
	}

	ofd, err := loadOriginsFigureData()
	if err != nil {
		return nx.Figure{}, fmt.Errorf("failed to load origins figure data: %w", err)
	}

	colorMap := origins.MakeColorMap(gdm.Figure())
	converter := origins.NewFigureConverter(ofd, colorMap)

	figure, err := converter.Convert(originsFigure)
	if err != nil {
		return nx.Figure{}, err
	}

	return figure, nil
}

func loadOriginsFigureData() (*origins.FigureData, error) {
	if time.Since(figureDataLastObtained) < figureDataCacheTime {
		return figureData, nil
	}

	res, err := http.Get("http://origins-gamedata.habbo.com/figuredata/1")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	figureData, err = origins.ParseFigureData(b)
	if err != nil {
		return nil, err
	}
	figureDataLastObtained = time.Now()
	return figureData, nil
}

func generatePlayerImage(player data.Player, figure nx.Figure, playerImageType PlayerImageType) error {

	var headOnly bool

	if playerImageType == Avatar {
		headOnly = true
	}

	filePath := fmt.Sprintf("static/%ss/%s.png", playerImageType, player.Username)

	mgr := gd.NewManager(host)
	renderer := imager.NewAvatarImager(mgr)

	err := util.LoadGameData(mgr, "Loading game data...",
		gd.GameDataFigure, gd.GameDataFigureMap,
		gd.GameDataVariables, gd.GameDataAvatar)
	if err != nil {
		return err
	}

	var parts []imager.AvatarPart
	// renderer.Parts is prone to panicking.. hack to get around it
	func() {
		defer func() {
			if excp := recover(); excp != nil {
				err = fmt.Errorf(
					"caught exception in render parts for player: %s, recovering: %+v\n", player.Username, excp,
				)
			}
		}()
		parts, err = renderer.Parts(figure)
	}()

	if err != nil {
		return err
	}

	libraries := map[string]struct{}{}

	for _, part := range parts {
		libraries[part.LibraryName] = struct{}{}
	}

	for lib := range libraries {
		err = mgr.LoadFigureParts(lib)
		if err != nil {
			return err
		}
	}

	avatar := imager.Avatar{
		Figure:        figure,
		Direction:     4,
		HeadDirection: 4,
		Actions:       []nx.AvatarState{nx.AvatarState(nx.ActStand)},
		Expression:    nx.AvatarState(nx.ActStand),
		HeadOnly:      headOnly,
	}

	anim, err := renderer.Compose(avatar)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := imager.NewEncoderPNG()
	encoder.EncodeFrame(f, anim, 0, 0)

	log.Printf("output: %s\n", filePath)
	return nil
}
