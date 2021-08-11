package UnitTests

import (
	"testing"

	"github.com/Jest0r/starex_go/galaxy"
)

func TestCoords(t *testing.T) {
	a := galaxy.CoordsI16{1,2,3}
	b := galaxy.CoordsI16{4,5,6}

	addResult := galaxy.CoordsI16{5,7,9} 
	subResult := galaxy.CoordsI16{-3,-3,-3} 
	mdResult := galaxy.CoordsI16{3,3,3}


	if a.Add(b) != addResult {
		t.Errorf("Error adding Coordinates. Should be %v, - is %v", addResult, a.Add(b))
	}

	if a.Sub(b) != subResult {
		t.Errorf("Error subtracting Coordinates. Should be %v, - is %v", subResult, a.Sub(b))
	}

	if a.AbsDist(b) != mdResult {
		t.Errorf("Error subtracting Coordinates. Should be %v, - is %v", mdResult, a.AbsDist(b))
	}

	if a.ManhattanDist(b) != 9 {
		t.Errorf("Error subtracting Coordinates. Should be %v, - is %v", 9, a.ManhattanDist(b))
	}

	if int(a.Distance(b)*100) != 519 {
		t.Errorf("Error subtracting Coordinates. Should be %v, - is %v", 519, a.Distance(b))
	}

	tp := a.ToPolar()
	if int16(tp.L*100) != 374 || int16(tp.A*100) != 110 || int16(tp.B*100) != 93 {
		t.Errorf("Error converting to Polar coords. Should be (float / 100 of) %v/%v/%v, - is %v", 374, 110, 94, a.ToPolar())
	}

	fpShould := galaxy.CoordsI16{0,1,20}
	fp := galaxy.CoordsI16{}
	fp.FromPolar(galaxy.CoordsPolar{21,2,1.5})
	if fp != fpShould {
		t.Errorf("Error converting to Polar coords. Should be %v, - is %v", fpShould, fp)
	}

	galaxy.TrigBuf.Activate(1000)
	fp.FromPolar(galaxy.CoordsPolar{21,2,1.5})

}