package util

import (
	"encoding/json"
	"os"
)

type Config interface {
	keywords() []string
	operators() []string
}

type DefaultConfig struct {
	Tokenizer struct {
		Keywords  []string
		Operators []string
	}
}

func NewDefaultConfig(filename string) (DefaultConfig, error) {
	c := DefaultConfig{}

	configFile, err := os.ReadFile(filename)
	if err != nil {
		return c, err
	}

	err = json.Unmarshal(configFile, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (c *DefaultConfig) keywords() []string {
	return c.Tokenizer.Keywords
}

func (c *DefaultConfig) operators() []string {
	return c.Tokenizer.Operators
}
