package galaxy

import (
	"fmt"
	"math"
	"time"
)

var TrigBuf TrigBuffer

type TrigBuffer struct {
	Active     bool
	sin        []float64
	cos        []float64
	precision  int
	maxval     int
	twopi      float64
	multiplier float64
}

// buffers Sin and Cos values to a given precision
func (tb *TrigBuffer) Activate(prec int) {
	starttime := time.Now()
	tb.twopi = 2 * math.Pi
	// maximum 10^5 entries a' 8 bytes = 800.000 bytes
	maxPrecision := 5
	if prec > maxPrecision {
		tb.precision = maxPrecision
	} else {
		tb.precision = prec
	}
	tb.maxval = int(math.Pow10(int(tb.precision)))
	// make sin buffer
	tb.sin = make([]float64, tb.maxval)
	tb.cos = make([]float64, tb.maxval)
	increment := math.Pi * 2 / float64(tb.maxval)
	var s float64
	var c float64
	var x float64
	for i := 0; i < tb.maxval; i++ {
		s = math.Sin(x)
		c = math.Cos(x)
		x += increment
//		ind := int(x * float64(tb.maxval))
		tb.sin[i] = s 
		tb.cos[i] = c
//		fmt.Printf("%d, %f, %f |", ind, x, y)
	}
	//fmt.Println("\nincrement:", increment, len(tb.sin), "\n-----------------")
	tb.multiplier = float64(tb.maxval) / tb.twopi
	fmt.Println("Trigometric Buffer creation took ", time.Since(starttime))
	tb.Active = true
}

func (tb *TrigBuffer) Sin(alpha float64) float64 {
	ind := int(alpha * tb.multiplier)
	return tb.sin[ind]
}
func (tb *TrigBuffer) Cos(alpha float64) float64 {
	ind := int(alpha * tb.multiplier)
	return tb.cos[ind]
}

type CoordsPolar struct {
	L float64
	A float64
	B float64
}

// Simple 3D coordinate system in 16bit integer
type CoordsI16 struct {
	X int16
	Y int16
	Z int16
}

// Add another coord, return result
func (c *CoordsI16) Add(a CoordsI16) CoordsI16 {
	n := CoordsI16{}
	n.X = c.X + a.X
	n.Y = c.Y + a.Y
	n.Z = c.Z + a.Z
	return n
}

// Sub another coord, return result
func (c *CoordsI16) Sub(a CoordsI16) CoordsI16 {
	n := CoordsI16{}
	n.X = c.X - a.X
	n.Y = c.Y - a.Y
	n.Z = c.Z - a.Z
	return n
}

// Sub another coord, return result
func (c *CoordsI16) AbsDist(a CoordsI16) CoordsI16 {
	n := CoordsI16{}
	n.X = absInt16(c.X - a.X)
	n.Y = absInt16(c.Y - a.Y)
	n.Z = absInt16(c.Z - a.Z)
	return n
}

// Sub another coord, return result
func (c *CoordsI16) ManhattanDist(a CoordsI16) int16 {
	n := CoordsI16{}
	n.X = absInt16(c.X - a.X)
	n.Y = absInt16(c.Y - a.Y)
	n.Z = absInt16(c.Z - a.Z)
	return n.X+n.Y+n.Z
}

// returns distance to another coord as float32
func (c *CoordsI16) Distance(a CoordsI16) float64 {
	md := c.AbsDist(a)
	return math.Sqrt(float64(md.X*md.X + md.Y*md.Y + md.Z*md.Z))
}

func (c *CoordsI16) FromPolar(p CoordsPolar) {
	var Sin func(float64) float64
	var Cos func(float64) float64
	if TrigBuf.Active {
		Sin = TrigBuf.Sin
		Cos = TrigBuf.Cos
	} else {
		Sin = math.Sin
		Cos = math.Cos
	}


	//lenXY := p.L * math.Cos(p.B)
	//c.Z = int16(p.L * math.Sin(p.B))
	//c.X = int16(lenXY * math.Cos(p.A))
	//c.Y = int16(lenXY * math.Sin(p.A))
	lenXY := p.L * Cos(p.B)
	c.Z = int16(p.L * Sin(p.B))
	c.X = int16(lenXY * Cos(p.A))
	c.Y = int16(lenXY * Sin(p.A))
}

func (c *CoordsI16) ToPolar() CoordsPolar {
	p := CoordsPolar{}
	lenXY := math.Sqrt(float64(c.X*c.X + c.Y*c.Y))
	p.A = math.Atan2(float64(c.Y), float64(c.X))
	p.B = math.Atan2(float64(c.Z), lenXY)
	p.L = math.Sqrt(lenXY*lenXY + float64(c.Z*c.Z))
	return p
}

// returns absolute value of an int16 num
func absInt16(x int16) int16 {
	if x < 0 {
		return 0 - x
	}
	return x
}
