// Package provides all needed game management functions
package game

import (
	"fmt"
	"log"
	"os"

	//	"math/rand"

	"gopkg.in/yaml.v3"

	"github.com/Jest0r/starex_go/galaxy"
	gui "github.com/Jest0r/starex_go/gui_opengl"
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
}

// Game initialisation
func (g *Game) Init() {
	fmt.Println("Init")
	g.Title = "Starex"
	g.config.ReadConfig("cfg/config.yaml")
	fmt.Println(g.config)

	// TODO: Load should also be done here, and not directly in the GUI
	g.Galaxy.Init()
	g.Gui.Init()
	g.Gui.Galaxy = &g.Galaxy

	// either Create()...
	g.Galaxy.Create(200000, 20000, 2000)
	// ... or LoadFromFile()
	//g.Galaxy.LoadFromFile("saves/galaxy2")

	fmt.Println(g.Galaxy.SysCount)

	g.Gui.PrepareScene()

}

// Game mainloop. Loop is handled within the function
func (g *Game) Mainloop() {
	fmt.Println("Skipping Mainloop")
	g.Gui.Mainloop()
}

// Game destructor
func (g *Game) Cleanup() {
	fmt.Println("Cleaning up...")
}
