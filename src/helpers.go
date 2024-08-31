package gchat

func getUserColours(username string) ColourPair {
	userColour, ok := config.PlayerColours[username]
	if ok {
		return userColour
	}
	for _, colourPair := range config.AvailableColours {
		userColour = ColourPair{
			BackgroundColour: colourPair.BackgroundColour,
			TextColour:       colourPair.TextColour,
		}
		config.PlayerColours[username] = userColour
		return userColour
	}
	return ColourPair{}
}
