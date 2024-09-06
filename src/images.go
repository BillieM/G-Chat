package gchat

import (
	"bytes"
	"errors"
	"fmt"
	"g-chat/src/data"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/image/draw"
	"xabbo.b7c.io/nx"
	"xabbo.b7c.io/nx/cmd/nx/util"
	gd "xabbo.b7c.io/nx/gamedata"
	"xabbo.b7c.io/nx/gamedata/origins"
	"xabbo.b7c.io/nx/imager"
)

const (
	figureSize                            string        = "l"
	figureEndpoint                        string        = "https://www.habbo.com/habbo-imaging/avatarimage?size=%s&figure=%s"
	minimumTimeBetweenPlayerImageRequests time.Duration = time.Hour * 1
	figureDataCacheTime                   time.Duration = time.Hour * 4
)

type playerImage struct {
	Type      playerImageType
	UnscaledX int
	UnscaledY int
}

type playerImageType string

var (
	FigureType playerImageType = "figure"
	AvatarType playerImageType = "avatar"
)

var (
	Figure playerImage = playerImage{
		Type:      FigureType,
		UnscaledX: 50,
		UnscaledY: 100,
	}

	Avatar playerImage = playerImage{
		Type:      AvatarType,
		UnscaledX: 50,
		UnscaledY: 50,
	}
)

func (p playerImage) String() string {
	return fmt.Sprint(p.Type)
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

func generatePlayerImage(player data.Player, figure nx.Figure, playerImageType playerImage) error {

	var headOnly bool

	if playerImageType == Avatar {
		headOnly = true
	}

	mgr := gd.NewManager(host)
	renderer := imager.NewAvatarImager(mgr)

	err := util.LoadGameData(mgr, "",
		gd.GameDataFigure, gd.GameDataFigureMap,
		gd.GameDataVariables, gd.GameDataAvatar)
	if err != nil {
		return fmt.Errorf("err loading game data: %w", err)
	}

	var parts []imager.AvatarPart

	// renderer.Parts is prone to panicking due to currently missing code to handle hats
	// this is a temp hack to get around it
	func() {
		defer func() {
			if excp := recover(); excp != nil {
				err = fmt.Errorf(
					"caught exception in render parts for player: %s, recovering: %+v", player.Username, excp,
				)
			}
		}()
		parts, err = renderer.Parts(figure)
	}()

	if err != nil {
		return fmt.Errorf("err rendering parts: %w", err)
	}

	libraries := map[string]struct{}{}

	for _, part := range parts {
		libraries[part.LibraryName] = struct{}{}
	}

	for lib := range libraries {
		err = mgr.LoadFigureParts(lib)
		if err != nil {
			return fmt.Errorf("err loading figure parts: %w", err)
		}
	}

	// hack to do 2 images for my char
	var imgs []bool = []bool{false}

	if playerImageType == Avatar && player.IsMe {
		imgs = append(imgs, true)
	}

	for _, isMe := range imgs {

		var meDir string = ""
		var headDirection = 2
		if isMe {
			headDirection = 4
			meDir = "/me"
		}

		avatar := imager.Avatar{
			Figure:        figure,
			Direction:     2,
			HeadDirection: headDirection,
			Actions:       []nx.AvatarState{nx.AvatarState(nx.ActStand)},
			Expression:    nx.AvatarState(nx.ActStand),
			HeadOnly:      headOnly,
		}

		anim, err := renderer.Compose(avatar)
		if err != nil {
			return fmt.Errorf("err composing avatar: %w", err)
		}

		// write image to buffer so we can do further processing (scaling/ cropping)
		// sadly we are double encoding/ decoding to avoid messing with nx code
		var imageBuffer bytes.Buffer

		encoder := imager.NewEncoderPNG()
		encoder.EncodeFrame(&imageBuffer, anim, 0, 0)

		// figures should be 100x200 - scale x2 + canvas size change
		// avatars should be 50x50 - canvas size change only

		unprocessedImg, err := png.Decode(&imageBuffer)

		if err != nil {
			return fmt.Errorf("err decoding png: %w", err)
		}

		unprocessedImgRect := image.Rect(0, 0, playerImageType.UnscaledX, playerImageType.UnscaledY)
		unscaledImg := image.NewRGBA(unprocessedImgRect)

		fixedSizeImgStartPoint := image.Point{
			X: (playerImageType.UnscaledX - unprocessedImg.Bounds().Max.X) / 2,
			Y: (playerImageType.UnscaledY - unprocessedImg.Bounds().Max.Y) / 2,
		}

		r := image.Rectangle{fixedSizeImgStartPoint, fixedSizeImgStartPoint.Add(unprocessedImgRect.Size())}
		draw.Draw(unscaledImg, r, unprocessedImg, unprocessedImgRect.Min, draw.Src)

		var outImg *image.RGBA

		if playerImageType == Figure {
			// scale image too if figure
			outImg = image.NewRGBA(image.Rect(0, 0, unscaledImg.Bounds().Max.X*2, unscaledImg.Bounds().Max.Y*2))
			draw.NearestNeighbor.Scale(outImg, outImg.Rect, unscaledImg, unscaledImg.Bounds(), draw.Over, nil)
		} else {
			outImg = unscaledImg
		}

		filePath := fmt.Sprintf("static/images/%ss%s/%s.png", playerImageType, meDir, player.Username)

		f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("err opening file: %w", err)
		}
		defer f.Close()

		// Encode to file:
		png.Encode(f, outImg)

		log.Printf("output: %s\n", filePath)
	}
	return nil

}
