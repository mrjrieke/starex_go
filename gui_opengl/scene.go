package gui_opengl

import (
	"fmt"
	"math"

	"github.com/Jest0r/starex_go/galaxy"
)

type Scene struct {
	Points []float32
	Colors []float32
	Lums   []float64
}

// calculating simple histograms on galaxy data. Mainly for debugging purposes
func (sc *Scene) getHistograms(galaxy *galaxy.Galaxy) {
	lumHist := make([]int, 256)
	lumSCHist := make([]int, 256)
	lumDfltHist := make(map[float64]int)
	typeHist := make(map[string]int)
	var maxlum float64

	for _, t := range galaxy.StarTypes.Types {
		typeHist[t.Type] = 0
		lumDfltHist[t.Luminosity] = 0
	}

	// find maximum lumen
	for _, s := range galaxy.Systems {
		if s.Lum > maxlum {
			maxlum = s.Lum
		}

		typeHist[s.CenterObject.Type]++
		lumDfltHist[s.CenterObject.Luminosity]++
	}

	// Lum histogram
	for _, s := range galaxy.Systems {
		lumHist[int(s.Lum*255/maxlum)]++
	}
	fmt.Println(maxlum, lumHist)

	maxlum = 0
	// lum histogram in SC
	for _, l := range sc.Lums {
		if l > maxlum {
			maxlum = l
		}
	}
	fmt.Println("maxlum sc", maxlum)

	for _, l := range sc.Lums {
		lumSCHist[int(l*255/maxlum)]++
	}

	fmt.Println(typeHist)
	fmt.Println(lumDfltHist)
	fmt.Println(maxlum, lumSCHist)
	// ------

}

func (sc *Scene) LoadData(galaxy *galaxy.Galaxy, scaleFactor float32) {
	// array for the stars and for the colors separately, for rendering reasons
	sc.Points = make([]float32, (galaxy.SysCount)*3)
	sc.Colors = make([]float32, (galaxy.SysCount)*4)
	sc.Lums = make([]float64, (galaxy.SysCount))

	//	sc.getHistograms(galaxy)

	for i, s := range galaxy.Systems {
		sc.Lums[i] = s.Lum
	}
	sc.normalizeLum(0, 100_000, 100_000, 0.2)

	sc.getHistograms(galaxy)
	//_ = lum_hist

	for i, sys := range galaxy.Systems {
		// switch Y and Z from original data to map the OpenGL coord system.
		sc.Points[3*i] = float32(sys.Coords.X) / scaleFactor
		sc.Points[3*i+1] = float32(sys.Coords.Z) / scaleFactor
		sc.Points[3*i+2] = float32(sys.Coords.Y) / scaleFactor

		sc.Colors[4*i] = float32(galaxy.Systems[i].Color.R) / 255
		sc.Colors[4*i+1] = float32(galaxy.Systems[i].Color.G) / 255
		sc.Colors[4*i+2] = float32(galaxy.Systems[i].Color.B) / 255
		// using this for Lum, because for some reason a separate array is only half read
		sc.Colors[4*i+3] = float32(sc.Lums[i])
	}
}

// tries to map the huge variance of object Luminosity into a displayable format
// rules (so far):
// lums < minmvalue are bumped up to max value
// values are recalculated to a range minvalue - maxvalue
// values are raise to the power exponent (should be <1 to spread even more)
func (sc *Scene) normalizeLum(minvalue float64, maxvalue float64, newMax float64, exponent float64) {

	// find max
	var maxlum float64
	var lumDivisor float64

	for i, s := range sc.Lums {
		// boosting small lums up to minimal value
		if sc.Lums[i] < minvalue {
			sc.Lums[i] = minvalue
		}
		if sc.Lums[i] > maxvalue {
			sc.Lums[i] = maxvalue
		}
		// collecting max lum
		if s > maxlum {
			maxlum = s
		}
	}

	lumDivisor = maxlum / maxvalue
	fmt.Println("lum divisor", lumDivisor, "exponent", exponent)

	lumDivisor = newMax / maxlum

	for i, l := range sc.Lums {
		sc.Lums[i] = math.Pow((float64(l / lumDivisor)), exponent)
	}
}
