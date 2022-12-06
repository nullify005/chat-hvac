package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AppToken string `yaml:"appToken"`
	BotToken string `yaml:"botToken"`
	Channel  string `yaml:"channel"`
	Intesis  string `yaml:"intesis"`
	Device   string `yaml:"device"`
}

func New(path string) (*Config, error) {
	c := &Config{}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	d := yaml.NewDecoder(strings.NewReader(string(body)))
	d.KnownFields(true)
	err = d.Decode(&c)
	return c, nil
}
