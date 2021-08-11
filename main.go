package main

import (
	game "github.com/Jest0r/starex_go/game"
	//	gui "github.com/Jest0r/starex_go/gui_opengl"
	//	"github.com/g3n/engine/graphic"

	"runtime"
)

func main() {
	// neccessary, otherwise everything breaks.
	runtime.LockOSThread()

	var game game.Game
	//	var gui gui.Gui

	defer game.Cleanup()

	game.Init()
	game.Mainloop()

}
