package game

import (
	"fmt"
)

type Game struct {
	Title string
}

func (g *Game) Init() {
	fmt.Println("Init")
	g.Title = "Starex"
}


func (g *Game) Mainloop() {
	fmt.Println("Skipping Mainloop")
}

func (g *Game) Cleanup() {
	fmt.Println("Cleaning up...")
}

