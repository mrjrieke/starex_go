package mob

import (
	"github.com/Jest0r/starex_go/coords"
)

type Ship struct {
	// limits
	MaxAcceleration float64
	MaxSpeed float64
	Range float64

	// Pos/movement
	pos coords.CoordsF64
	speedVec coords.CoordsF64
}

func (s *Ship) Pos() coords.CoordsF64 {
	return s.pos
}

func (s *Ship) SpeedVec() coords.CoordsF64 {
	return s.speedVec
}

func (s *Ship) Velocity() float64 {
	return s.speedVec.ToPolar().L
}



