package mob

import (
	"github.com/Jest0r/starex_go/coords"
)

type mob interface {
	Pos() coords.CoordsF64
	SpeedVec() coords.CoordsF64
	Velocity() float64
}

