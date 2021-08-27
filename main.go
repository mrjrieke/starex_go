package main

import (
	game "github.com/Jest0r/starex_go/game"
)

func main() {

	var game game.Game

	defer game.Cleanup()

	game.Init()
	game.Mainloop()

}
