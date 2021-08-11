// Package provides all needed game management functions
package game

import (
	"fmt"
	"log"
	"os"

	//	"math/rand"

	"gopkg.in/yaml.v3"

	"github.com/Jest0r/starex_go/galaxy"
	"github.com/Jest0r/starex_go/gui_opengl"
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
	Gui gui_opengl.Gui
	config Config
	Galaxy galaxy.Galaxy
}

// Game initialisation
func (g *Game) Init() {
	fmt.Println("Init")
	g.Title = "Starex"
	g.config.ReadConfig("cfg/config.yaml")
	fmt.Println(g.config)

	g.Galaxy.Init(1000, 20000, 2000)
	g.Galaxy.Create()

	fmt.Println(g.Galaxy.SysCount)
	//	fmt.Printf("%v\n", g.galaxy.Systems[:33])
	//	fmt.Printf("%v\n", g.galaxy.Systems)
}

// Game mainloop. Loop is handled within the function
func (g *Game) Mainloop() {
	fmt.Println("Skipping Mainloop")
}

// Game destructor
func (g *Game) Cleanup() {
	fmt.Println("Cleaning up...")
}
