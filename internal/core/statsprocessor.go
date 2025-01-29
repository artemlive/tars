package core

import (
	"log"

	"github.com/artemlive/tars/pkg/utils"
	"github.com/slack-go/slack"
)

type StatsProcessor struct {
	config *utils.Config
}

func NewStatsProcessor(config *utils.Config) *StatsProcessor {
	return &StatsProcessor{config: config}
}

func (sp *StatsProcessor) ShouldProcessMessage(channelID string, message slack.Message, beaconReaction string) bool {
	if beaconReaction == "" {
		return true
	}
	for _, reaction := range message.Reactions {
		if reaction.Name == beaconReaction {
			return true
		}
	}
	return len(message.Reactions) > 0
}

func (sp *StatsProcessor) UpdateStats(channelID string, reactions []slack.ItemReaction, stats map[string]int) {
	for _, reaction := range reactions {
		log.Printf("Processing reaction: %s", reaction.Name)
		category, exists := utils.GetCategoryForReaction(sp.config, channelID, reaction.Name)
		if exists {
			stats[category] += reaction.Count
		}
	}
}
