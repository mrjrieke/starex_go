package gui_opengl

import (
	"math"

	"github.com/Jest0r/starex_go/galaxy"
)

type Scene struct {
	Points []float32
	Colors []float32
	Lums   []float64
}

func (sc *Scene) LoadData(galaxy *galaxy.Galaxy, scaleFactor float32) {
	// array for the stars and for the colors separately, for rendering reasons
	sc.Points = make([]float32, (galaxy.SysCount)*3)
	sc.Colors = make([]float32, (galaxy.SysCount)*4)
	sc.Lums = make([]float64, (galaxy.SysCount))

	for i, s := range galaxy.Systems {
		sc.Lums[i] = s.Lum
	}
	//fmt.Println(sc.Lums[:200])
	 sc.normalizeLum2(5)
	//sc.normalizeLum(5)
	//fmt.Println(sc.Lums[:200])

	lum_hist := make([]int, 256)
	_ = lum_hist

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

func (sc *Scene) normalizeLum(maxvalue float64) {
	// find max
	var maxlum float64
	var lumDivisor float64
	minLumIdx := float64(13) / float64(256)
	maxLumIdx := float64(199) / float64(256)
	for _, s := range sc.Lums {
		if s > maxlum {
			maxlum = s
		}
	}
	lumDivisor = maxlum / maxvalue
	for i, l := range sc.Lums {
		sc.Lums[i] = math.Pow((float64(l / lumDivisor)), 0.1)
		//sc.Lums[i] = float32(math.Pow((float64(l / maxlum)), 0.1))
		if sc.Lums[i] < minLumIdx {
			sc.Lums[i] = minLumIdx
		}
		if sc.Lums[i] > maxLumIdx {
			sc.Lums[i] = maxLumIdx
		}
		sc.Lums[i] = sc.Lums[i] * 256 / 199
	}
	/* current distribution:
								  v--- get rid of everything below (idnex 13)
	[17134 0 0 0 0 0 0 0 0 0 0 0 0 2 0 0 5 4 0 4 6 0 21 11 0 19207 0 1653 70 7 0 23145 26435 0 9188 972 102 11 0 30218 1491 8551 1189 1332 230 30
	 7351 1440 65 5183 5316 427 5443 397 212 144 3542 219 156 13 3472 305 71 45 1 921 39 2 16 0 5 191 3 195 3 1405 36 2 0 2 183 6 0 1 224 6 208 1
	 2 0 187 1 0 0 1 0 0 0 208 229 18440 113 31 18 0 1 224 768 8 0 0 27 11 0 2 2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 7 3 4 1 1 1 1 0 207
	  0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 323 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 308 0 0 0 0 0 0 244 0 0 0 0 351 0 0 0 0 0 0 0
	  0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 2 1]                                           ^
	  																		 get rid of everything above (index 255-48=207) -------+

	*/

}

func (sc *Scene) normalizeLum2(maxvalue float32) {
	// find max
	var maxlum float64
	for _, s := range sc.Lums {
		if s > maxlum {
			maxlum = s
		}
	}
	for i, l := range sc.Lums {
		sc.Lums[i] = math.Pow(float64(l), 0.1)
	}
}
