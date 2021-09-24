package galaxy

import (
	"math"
)

var TrigBuf TrigBuffer

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
	return n.X + n.Y + n.Z
}

// returns square distance to another coord as float32
func (c *CoordsI16) DistanceSq(a CoordsI16) float64 {
	md := c.AbsDist(a)
	xf := float64(md.X)
	yf := float64(md.Y)
	zf := float64(md.Z)
	//return math.Sqrt(xf*xf + yf*yf + zf*zf)
	return xf*xf + yf*yf + zf*zf
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
