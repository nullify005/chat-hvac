package cli

import (
	"log"

	"github.com/nullify005/chat-hvac/pkg/config"
	"github.com/nullify005/chat-hvac/pkg/slack"
)

func Config(path string) *config.Config {
	c, err := config.New(path)
	if err != nil {
		log.Fatalln(err)
	}
	return c
}

func Watch(botToken, appToken, channel, intesis, device string) {
	s, err := slack.New(botToken, appToken, channel, intesis, device)
	if err != nil {
		log.Fatalln(err)
	}
	s.ListenForMention()
}
