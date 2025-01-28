package utils

func GetControllingReaction(config *Config, channelID string) string {
	for _, channel := range config.Channels {
		if channel.ID == channelID {
			return channel.BeaconReaction
		}
	}
	return ""
}

func GetCategoryForReaction(config *Config, channelID, reaction string) (string, bool) {
	if channelReactions, exists := config.ReactionCache[channelID]; exists {
		category, found := channelReactions[reaction]
		return category, found
	}
	return "", false
}
