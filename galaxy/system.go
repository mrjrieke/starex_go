package galaxy

import (
	"fmt"
	"math"
	"strconv"
)

const (
	LumMin  = 10.0
	LumExp  = 0.20
	LumMult = 15
)

type Color struct {
	R int32
	G int32
	B int32
	A int32
}

type System struct {
	CenterObject StellarObject
	//	Coords   Coordinates `json:"coords"`
	Lum      float64 `json:"lum"`
	Colorstr string  `json:"color"`
	Coords   CoordsI16
	Color    Color
}

func (s *System) print() {
	fmt.Printf("System - Coords %v", s.Coords)
}

func (s *System) PlaceCenterObject() {

}

func (s *System) SetColor(colorstr string, lum float64) {
	if len(colorstr) < 5 {
		s.Color.R = 89
		//s.Color.G=22
		//s.Color.B=89
		s.Color.G = 0
		s.Color.B = 89
		//s.Lum=1000
		s.Lum = 0.001
		s.Color.A = 50
	} else {
		r, _ := strconv.ParseInt(colorstr[1:3], 16, 16)
		g, _ := strconv.ParseInt(colorstr[3:5], 16, 16)
		b, _ := strconv.ParseInt(colorstr[5:7], 16, 16)
		s.Color.R = int32(r)
		s.Color.G = int32(g)
		s.Color.B = int32(b)
		s.Lum = lum

		alpha := math.Pow(lum, LumExp) * LumMult
		alpha = math.Max(alpha, LumMin)
		s.Color.A = int32(math.Min(alpha, 255))
	}
}
