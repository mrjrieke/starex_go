package UnitTests

import (
	"testing"

	"github.com/Jest0r/starex_go/game"
)

func TestConfig(t *testing.T) {
	c := game.Config{}
	c.ReadConfig("./game_test.yaml")
	if c.Logging.Log_level != "DEBUG" {
		t.Errorf("ERROR in game.config. Should be 'DEBUG', %s ", c.Logging.Log_level)
	}
}

func TestGame(t *testing.T) {
}