package gui_opengl


import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/engoengine/glm"
)

const (
	LumMin  = 10.0
	LumExp  = 0.20
	LumMult = 15
	W       = 3
	X       = 0
	Y       = 1
	Z       = 2
)

type Galaxy struct {
	meta       Metadata // could probably be discarded and only used locally in Galaxy.Import()
	stars      []System
	numSystems int
	radius     int
}

func (galaxy *Galaxy) Import(filepath string) {
	metafile := "/galaxy.meta"
	starfile := "/galaxy.json"

	metatext := scanFile(filepath + metafile)
	// Get data from scan with Bytes() or Text()
	err := json.Unmarshal([]byte(metatext), &galaxy.meta)
	if err != nil {
		log.Fatal(err)
	}
	galaxy.numSystems = galaxy.meta.NumSystems
	fmt.Printf("File Version: %d, NumSystems %d, RandSeed %d\n", galaxy.meta.FileVersion, galaxy.meta.NumSystems, galaxy.meta.RandSeed)

	// start importing the Galaxy
	galaxy.stars = make([]System, galaxy.meta.NumSystems)
	file, err := os.Open(filepath + starfile)
	if err != nil {
		log.Fatal(err)
	}

	byteValue, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &galaxy.stars)

	for i, s := range galaxy.stars {
		if s.Coords.X > galaxy.radius {
			galaxy.radius = s.Coords.X
		}
		if s.Coords.Z > galaxy.radius {
			galaxy.radius = s.Coords.Z
		}
		galaxy.stars[i].setColor()
	}
	fmt.Println("Galaxy radius: ", galaxy.radius)

}

type System struct {
	Coords   Coordinates `json:"coords"`
	Lum      float32     `json:"lum"`
	Colorstr string      `json:"color"`
	Color    Color
}

func (s *System) setColor() {
	r, _ := strconv.ParseInt(s.Colorstr[1:3], 16, 16)
	g, _ := strconv.ParseInt(s.Colorstr[3:5], 16, 16)
	b, _ := strconv.ParseInt(s.Colorstr[5:7], 16, 16)
	s.Color.r = int32(r)
	s.Color.g = int32(g)
	s.Color.b = int32(b)

	alpha := math.Pow(float64(s.Lum), LumExp) * LumMult
	alpha = math.Max(alpha, LumMin)
	s.Color.a = int32(math.Min(alpha, 255))
}

type Coordinates struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type Color struct {
	r int32
	g int32
	b int32
	a int32
}

type OpenGLColor struct {
	r float32
	g float32
	b float32
	a float32
}

func step(edge float32, x float32) float32 {
	if x < edge {
		return 0.0
	}
	return 1.0
}

func mix(x glm.Vec4, y glm.Vec4, a float32) glm.Vec4 {
	v1 := x.Mul(1.0 - a)
	v2 := y.Mul(a)
	return v1.Add(&v2)
}

type Metadata struct {
	FileVersion int `json:"file_version"`
	NumSystems  int `json:"num_systems"`
	RandSeed    int `json:"rand_seed"`
}

func scanFile(filename string) string {
	// Open file and create scanner on top of it
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	// Scan for next token.
	success := scanner.Scan()
	if success == false {
		// False on error or EOF. Check error
		err = scanner.Err()
		if err == nil {
			log.Println("Scan completed and reached EOF")
		} else {
			log.Fatal(err)
		}
	}
	defer file.Close()
	return scanner.Text()
}
