package coords

import (
	"math"
)

//var TrigBuf TrigBuffer

// Simple 3D coordinate system in 16bit integer
type CoordsF64 struct {
	X float64
	Y float64
	Z float64
}

// Add another coord, return result
func (c *CoordsF64) Add(a CoordsF64) CoordsF64 {
	n := CoordsF64{}
	n.X = c.X + a.X
	n.Y = c.Y + a.Y
	n.Z = c.Z + a.Z
	return n
}

// Sub another coord, return result
func (c *CoordsF64) Sub(a CoordsF64) CoordsF64 {
	n := CoordsF64{}
	n.X = c.X - a.X
	n.Y = c.Y - a.Y
	n.Z = c.Z - a.Z
	return n
}

// Sub another coord, return result
func (c *CoordsF64) AbsDist(a CoordsF64) CoordsF64 {
	n := CoordsF64{}
	n.X = math.Abs(c.X - a.X)
	n.Y = math.Abs(c.Y - a.Y)
	n.Z = math.Abs(c.Z - a.Z)
	return n
}

// Sub another coord, return result
func (c *CoordsF64) ManhattanDist(a CoordsF64) float64 {
	n := CoordsF64{}
	n.X = math.Abs(c.X - a.X)
	n.Y = math.Abs(c.Y - a.Y)
	n.Z = math.Abs(c.Z - a.Z)
	return n.X + n.Y + n.Z
}

// returns square distance to another coord as float32
func (c *CoordsF64) DistanceSq(a CoordsF64) float64 {
	md := c.AbsDist(a)
	xf := float64(md.X)
	yf := float64(md.Y)
	zf := float64(md.Z)
	//return math.Sqrt(xf*xf + yf*yf + zf*zf)
	return xf*xf + yf*yf + zf*zf
}

func (c *CoordsF64) FromPolar(p CoordsPolar) {
	var Sin func(float64) float64
	var Cos func(float64) float64

	Sin = math.Sin
	Cos = math.Cos

	lenXY := p.L * Cos(p.B)
	c.Z = p.L * Sin(p.B)
	c.X = lenXY * Cos(p.A)
	c.Y = lenXY * Sin(p.A)
}

func (c *CoordsF64) ToPolar() CoordsPolar {
	p := CoordsPolar{}
	lenXY := math.Sqrt(float64(c.X*c.X + c.Y*c.Y))
	p.A = math.Atan2(float64(c.Y), float64(c.X))
	p.B = math.Atan2(float64(c.Z), lenXY)
	p.L = math.Sqrt(lenXY*lenXY + float64(c.Z*c.Z))
	return p
}
