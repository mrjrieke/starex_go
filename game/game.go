// Package provides all needed game management functions
package game

import (
	"fmt"
	"log"
	"os"
	"time"

	"math/rand"

	"gopkg.in/yaml.v3"

	"github.com/Jest0r/starex_go/galaxy"
	gui "github.com/Jest0r/starex_go/gui_opengl"
)

const (
	GameStateTitle    = 0
	GameStateMainMenu = 1
	GameStateCreate   = 2
	GameStateLoad     = 3
	GameStateActive   = 10
)

// Holding the configuration file data
type Config struct {
	Logging struct {
		Logdir         string
		Log_level      string
		Logfile_name   string
		Logging_stdout bool
		Logging_file   bool
	}
}

// reads the config from a given filename
func (c *Config) ReadConfig(filename string) {
	yamlFile, err := os.Open(filename)
	if err != nil {
		log.Printf("yamlFile.Get error #%v", err)
	}
	defer yamlFile.Close()

	d := yaml.NewDecoder(yamlFile)
	err = d.Decode(&c)
	if err != nil {
		log.Fatalf("Config file unmarshal error #%v", err)
	}
}

// Holding all needed Game data
type Game struct {
	Title  string
	Gui    gui.Gui
	config Config
	Galaxy galaxy.Galaxy
	State  int
}

// Game initialisation
func (g *Game) Init() {
	fmt.Println("Init")
	g.Title = "Starex"
	g.config.ReadConfig("cfg/config.yaml")

	// TODO: Load should also be done here, and not directly in the GUI
	g.Galaxy.Init()
	g.Gui.Init()

	rand.Seed(int64(time.Now().Nanosecond()))

	// either Create()...
	// seems like the limit for a solid 60fps display on my current HW is around 700k-1M stars
	// depnding on the blur steps :)
	g.Galaxy.Create(200_000, 20000, 2000)
	//g.Galaxy.Create(100, 20000, 2000)
	// ... or LoadFromFile()
	//g.Galaxy.LoadFromFile("saves/galaxy2")

	// --- testing KDTree stuff
	g.Gui.Galaxy = &g.Galaxy

	// ---- Some kdtree tests ------
	// -- star lookup
	// ---- Star lookup via KNN:
	// Duration on 200k stars: 100us per lookup
	// Duration on 500k stars: 125us per lookup
	// ---- Star lookup via hash:
	// Duration on 200k stars: 408ns per lookup
	// Duration on 500k stars: 385ns per lookup
	// --- This could be quicker via x/y/z key lookup, but also more space consuming
	// pretty constant.

	fmt.Println(len(g.Galaxy.Systems), g.Galaxy.SysCount)

	i, s := g.Galaxy.GetRandomSystem()
	g.Galaxy.HilightedSystems = append(g.Galaxy.HilightedSystems, i)
	fmt.Println("Random System:", i, s)

	fmt.Println("Nearest systems:")
	nearest := g.Galaxy.GetKNearestSystems(s, 2)
	for i := range nearest {
		fmt.Println(nearest[i])
	}
	fmt.Println("Range search:")
	inRange := g.Galaxy.GetSystemsInRadius(s, 1000)

	// for each coord in range
	for i := range inRange {
		// get corresponding system
		sys := g.Galaxy.GetSysByCoords(inRange[i])
		// color it
		sys.SetColor("#cc00cc", sys.CenterObject.Lum())
		//sys.SetColor("#cc00cc", 10000)
		//			fmt.Println(inRange[i])
	}
	// color random star (after coloring the range, as the star is part of the range)
	g.Galaxy.Systems[i].SetColor("#ff00ff", 100000)

	g.Gui.PrepareScene()

}

// Game mainloop. Loop is handled within the function
func (g *Game) Mainloop() {
	for !g.Gui.GameExitRequested {
		// -- finally a run of the GUI mainloop
		g.Gui.Mainloop()
	}
}

// Game destructor
func (g *Game) Cleanup() {
	fmt.Println("Exit. Cleaning up...")
	defer g.Gui.Cleanup()
	fmt.Println("Bye.")
}
