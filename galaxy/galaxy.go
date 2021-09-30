package galaxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/Jest0r/starex_go/coords"

	"github.com/kyroy/kdtree"
	"github.com/kyroy/kdtree/kdrange"
)

const (
	STAR = 0
	WD   = 1
	SBH  = 2
	SNS  = 3
	W    = 3
	X    = 0
	Y    = 1
	Z    = 2
	IMBH = 0
	SGS  = 1
)

type Galaxy struct {
	Systems []*System
	//SysMap           map[int16]map[int16]map[int16]*System
	SysCount         int32
	HilightedSystems []int
	sysHash          map[int64]bool    // for quick collision checking
	sysMap           map[int64]*System // for quick coord lookup

	Tree kdtree.KDTree

	SysTarget       int32
	RadiusTarget    int16 // what radius are we aiming for
	ThicknessTarget int16

	Radius int16 // acual radius (highest absolute x||y value)

	ArmsMaxRad        float64
	ArmsOuterRad      int16
	ArmsInnerRad      int16
	ArmsEllipseFactor float64

	StellarSizeTypes SizeTypes

	StarTypes       StarTypes
	OStarTypes      StarTypes
	WDTypes         StarTypes
	IMBHTypes       StarTypes
	PlanetTypes     StarTypes
	PlanetTypesNear StarTypes
	PlanetTypesHZ   StarTypes
	PlanetTypesFar  StarTypes
}

func (g *Galaxy) PrintHashes() {
	//	fmt.Println(len(g.sysHash), g.sysHash)
	//	fmt.Println(len(g.sysMap), g.sysMap)
	for key := range g.sysMap {
		fmt.Println(key, g.sysHash[key], g.sysMap[key])
	}
}

func (g *Galaxy) Init() {
	// arms config
	g.ArmsMaxRad = 2 * math.Pi
	g.ArmsOuterRad = 150
	g.ArmsInnerRad = 1200
	g.ArmsEllipseFactor = 0.3

	g.sysHash = make(map[int64]bool)

	g.sysMap = make(map[int64]*System)

	g.Tree = kdtree.KDTree{}

	g.StellarSizeTypes.ReadSizeTypeData("data/stellar_data.json")

	g.OStarTypes.ReadStarData("data/o_star_data.json")
	g.WDTypes.ReadStarData("data/wd_data.json")

	// not sure if WD and O star data should go in here or not
	g.StarTypes.ReadStarData("data/o_star_data.json")
	g.StarTypes.ReadStarData("data/star_data.json")
	g.StarTypes.ReadStarData("data/wd_data.json")

	g.IMBHTypes.ReadStarData("data/imbh_data.json")

	//	Planets
	g.PlanetTypesNear.ReadStarData("data/planet_data_near.json")
	g.PlanetTypesHZ.ReadStarData("data/planet_data_hz.json")
	g.PlanetTypesFar.ReadStarData("data/planet_data_far.json")

	// same, but into a single array
	g.PlanetTypes.ReadStarData("data/planet_data_near.json")
	g.PlanetTypes.ReadStarData("data/planet_data_hz.json")
	g.PlanetTypes.ReadStarData("data/planet_data_far.json")

	g.Systems = []*System{}
	//	g.HilightedSystems = []int

	//	rand.Seed(0)
}

func (g *Galaxy) AddSystemAt(c coords.CoordsI16) *System {
	// get new object
	sys := new(System)
	// new coords struct
	sys.Coords = c

	// ---- record system into look up structures ----
	// append System to list of systems
	g.Systems = append(g.Systems, sys)
	// append System to the last added system to the Hashmap
	coordHash := g.getCoordsHash(sys.Coords)
	g.sysMap[coordHash] = sys
	// insert system into kdtree
	g.Tree.Insert(sys)
	// -------

	// create center objects
	g.CreateCenterObject(sys)
	sys.SetColor(sys.CenterObject.Color(), sys.CenterObject.Lum())

	// rough method of finding the max radius
	// to be further refined - maybe by x*x + y*y > g.Radius**2
	if sys.Coords.X > g.Radius {
		g.Radius = sys.Coords.X
	}
	if sys.Coords.Z > g.Radius {
		g.Radius = sys.Coords.Z
	}

	g.SysCount++
	return sys
}

func (g *Galaxy) getCoordsHash(coords coords.CoordsI16) int64 {
	return int64(int64(coords.Z)<<32 + int64(coords.X)<<16 + int64(coords.Y))
}

func (g *Galaxy) GetSysByCoords(c coords.CoordsI16) *System {
	cHash := g.getCoordsHash(c)

	if g.sysHash[cHash] {
		return g.sysMap[cHash]
	} else {
		panic("PANIC: can't find system hash!")
		fmt.Println("ERROR!")
		return nil
	}
}

func (g *Galaxy) GetRandomSystem() (int, *System) {
	ind := rand.Intn(int(g.SysCount - 1))
	return ind, g.Systems[ind]
}

func (g *Galaxy) GetKNearestSystems(s *System, n int) []kdtree.Point {
	points := g.Tree.KNN(s, 2)
	return points
}

func (g *Galaxy) GetSystemsInRadius(s *System, r int16) []coords.CoordsI16 {
	radSquare := float64(r) * float64(r)
	var inRange []coords.CoordsI16

	pointsInSquare := (g.Tree.RangeSearch(kdrange.New(
		float64(s.Coords.X-r), float64(s.Coords.X+r),
		float64(s.Coords.Y-r), float64(s.Coords.Y+r),
		float64(s.Coords.Z-r), float64(s.Coords.Z+r))))

	for _, pt := range pointsInSquare {
		sys := coords.CoordsI16{X: int16(pt.Dimension(0)), Y: int16(pt.Dimension(1)), Z: int16(pt.Dimension(2))}
		if s.Coords.DistanceSq(sys) <= radSquare {
			inRange = append(inRange, sys)
		}

	}

	fmt.Println("Get Systems in Radius -  Systems in Square / in Radius", len(pointsInSquare), len(inRange))

	return inRange
}

// Create The galaxy content.
// in here we will branch off into the randomizer and the various galaxy forms to create
func (g *Galaxy) Create(SysTarget int32, RTarget int16, TTarget int16) {
	starttime := time.Now()
	g.SysTarget = SysTarget
	g.RadiusTarget = RTarget
	g.ThicknessTarget = TTarget

	//	rand.Seed(0)

	// Activate sin/cos mapping if number of stars > 40000
	//	if g.SysTarget >= 50000 {
	//	TrigBuf.Activate(5)
	//	}

	// create coordinates for the 'standard spiral' form
	g.CreateFormSpiral1()
	//g.CreateForm2()

	// ------------------ next steps
	fmt.Println("Creation took ", time.Since(starttime))

	starttime = time.Now()

}

// -----------------------------------------------------------
// Different forms to create
// -----------------------------------------------------------
func (g *Galaxy) CreateFormSpiral1() {
	dctime := time.Now()
	g.AddDisc(0.25, 1, g.SysTarget/3)
	g.AddDisc(0.5, 0.25, g.SysTarget/6)
	fmt.Println("Disc creation took ", time.Since(dctime), "syscount", g.SysCount, "stars/ms:", float64(g.SysCount)/float64(time.Since(dctime).Milliseconds()))

	dctime = time.Now()
	g.AddArms(0.5, g.SysTarget/3, 2)
	g.AddArms(0.5, g.SysTarget/6, 4)
	fmt.Println("Arms creation took ", time.Since(dctime), "syscount", g.SysCount, "stars/ms:", float64(g.SysCount)/float64(time.Since(dctime).Milliseconds()))
}

func (g *Galaxy) CreateForm2() {
	g.AddShell(1, 0.3, g.SysTarget)
	fmt.Println("Created Form 2. Syscount", g.SysCount)
}

// -----------------------------------------------------------
// Component functions for all the different forms
// -----------------------------------------------------------

// create a flattened disc of systems, adds it to the systems in the Galaxy struct
// ATTENTION: relThickness is the SIGMA of the Normvariate, so 31.x% of all values are OUTSIDE of the value
// AMENDMENT - swapped to 2*SIGMA, so only ~4.55% of all values are outside
func (g *Galaxy) AddDisc(relRadius float64, relThickness float64, numStars int32) {
	// to adjust to 2*sigma
	radius := relRadius * float64(g.RadiusTarget)
	halfRadius := radius * 0.5
	thickness := relThickness * float64(g.ThicknessTarget)
	twopi := 2 * math.Pi
	fmt.Printf("Creating Disc. %d stars within %f ly radius and %f thickness\n", numStars, radius, thickness)

	var flatten float64 = radius / thickness

	var s int32
	dupes := 0
	for s < numStars {
		pos := coords.CoordsI16{}
		pc := coords.CoordsPolar{}
		// ! ATTENTION ! radius is the sigma of th this could lead to values > maxint16 - should we check for this before? this could lead to values > maxint16 - should we check for this before? this could lead to values > maxint16 - should we check for this before?e normal variate. So 31.x% of all the stars are OUTSIDE the radius
		// radius is th
		pc.L = rand.NormFloat64() * halfRadius
		pc.A = rand.Float64() * twopi
		pc.B = rand.Float64() * twopi
		// convert to real coords
		// !QUESTION!  this could lead to values > maxint16 - should we check for this before?
		pos.FromPolar(pc)
		// flatten ball
		pos.Z = int16(float64(pos.Z) / flatten)

		map_idx := g.getCoordsHash(pos)

		// if there isn't a system at that location
		if !g.sysHash[map_idx] {
			//			fmt.Println(map_idx, g.sysHash[map_idx])
			// set it as occupied
			g.sysHash[map_idx] = true
			// add to list
			//	sys[s].Coords = coords
			g.AddSystemAt(pos)
			/// and move on
			s++
		} else {
			// record dupe
			dupes++
		}

		// TODO: check dupes
	}
	//	fmt.Printf("%d Stars in disc created (%d). dupes: %d \n", s, len(sys), dupes)

	//	g.AddSystems(numStars, sys)
}

// create a set of spiral arms and adds it
func (g *Galaxy) AddArms(relRadius float64, numStars int32, numArms int32) {
	radius := relRadius * float64(g.RadiusTarget)
	twopi := 2 * math.Pi
	starsPerArm := numStars / numArms
	armsDiff := float64(g.ArmsInnerRad - g.ArmsOuterRad)

	fmt.Println("Target stars: ", numStars)
	// per arm
	var armNum int32 = 0
	var sysCount int32 = 0
	//	sys := make([]System, int(numStars))
	dupes := 0
	// TODO. make Arms denser inside, and 'lighter' on the edge, so it doesn't stop suddenly
	angleIncrease := radius / float64(starsPerArm)
	for armNum = 0; armNum < numArms; armNum++ {
		// not sure why that is 1 here and not 0...
		var i float64 = 1
		var armAngle float64
		armStartRad := g.ArmsMaxRad / float64(numArms) * float64(armNum)
		// paint stars
		for i < radius {
			if sysCount == numStars {
				fmt.Println("ERROR - too many stars! Terminating arm generation")
				break
			}
			pos := coords.CoordsI16{}
			// offset for the star
			randomSphere := coords.CoordsI16{}
			pc := coords.CoordsPolar{}

			// set where on the way to the edge we are
			pc.L += i
			pc.A = armAngle + armStartRad
			pos.FromPolar(pc)

			// get random ball within a given radius for the star
			randomSphereRadius := ((1 - pc.L/radius) * armsDiff) + float64(g.ArmsOuterRad)

			pc.L = rand.NormFloat64() * randomSphereRadius
			pc.A = rand.Float64() * twopi
			pc.B = rand.Float64() * twopi
			randomSphere.FromPolar(pc)

			// flatten the ball
			randomSphere.Z = int16(float64(randomSphere.Z) * g.ArmsEllipseFactor)

			// Add random sphere to
			pos = pos.Add(randomSphere)

			// ch3eck for dupes
			map_idx := g.getCoordsHash(pos)

			// if there isn't a system at that location
			if !g.sysHash[map_idx] {
				// set it as occupied
				g.sysHash[map_idx] = true
				// add to list
				//				sys[sysCount].Coords = coords
				//sys[i].Coords = coords
				//			sysCount += 1
				g.AddSystemAt(pos)
			} else {
				// record dupe
				dupes++
			}

			armAngle += (float64(g.ArmsMaxRad) / float64(numStars)) * float64(numArms)
			// i gets incremented regardlessly, otherwise we get stuck in infinite loops
			i += angleIncrease
			// add to list
		}
	}
	fmt.Printf("%d Stars in %d Arms created. dupes %d \n", sysCount, armNum, dupes)
	//	g.AddSystems(numStars, sys)
	//g.Systems = append(g.Systems, sys...)
	//g.SysCount += numStars

}

// create a set of spiral arms and adds it
func (g *Galaxy) AddShell(relRadius float64, flatten float64, numStars int32) {
	radius := relRadius * float64(g.RadiusTarget)
	twopi := 2 * math.Pi
	armsDiff := float64(g.ArmsInnerRad - g.ArmsOuterRad)

	fmt.Println("Target stars: ", numStars)
	// per arm
	var armNum int32
	var sysCount int32
	//	sys := make([]System, int(numStars))
	dupes := 0
	angleIncrease := radius / float64(numStars)
	//var i int32 = 1
	var i float64 = 1
	var armAngle float64

	for i < radius {
		pos := coords.CoordsI16{}
		// offset for the star
		randomSphere := coords.CoordsI16{}
		pc := coords.CoordsPolar{}

		// set where on the way to the edge we are
		pc.L += i
		pc.B = rand.Float64() * twopi
		pc.A = float64(armAngle)

		pos.FromPolar(pc)
		pos.Z = int16(float64(pos.Z) * flatten)

		// get random ball within a given radius for the star
		randomSphereRadius := ((1 - pc.L/radius) * armsDiff) + float64(g.ArmsOuterRad)

		pc.L = rand.NormFloat64() * randomSphereRadius
		pc.A = rand.Float64() * twopi
		pc.B = rand.Float64() * twopi
		randomSphere.FromPolar(pc)

		// flatten the ball
		randomSphere.Z = int16(float64(randomSphere.Z))

		// Add random sphere to
		pos = pos.Add(randomSphere)

		// ch3eck for dupes
		map_idx := g.getCoordsHash(pos)

		// if there isn't a system at that location
		if !g.sysHash[map_idx] {
			// set it as occupied
			g.sysHash[map_idx] = true
			// add to list
			//sys[sysCount].Coords = coords
			g.AddSystemAt(pos)
			//sys[i].Coords = coords
			sysCount += 1
		} else {
			// record dupe
			dupes++
		}

		armAngle += (float64(g.ArmsMaxRad) / float64(numStars))
		// i gets incremented regardlessly, otherwise we get stuck in infinite loops
		i += angleIncrease
		// add to list
	}
	fmt.Printf("%d Stars in %d Arms created. dupes %d \n", sysCount, armNum, dupes)
	//	g.AddSystems(numStars, sys)
	//g.Systems = append(g.Systems, sys...)
	//g.SysCount += numStars

}

func (g *Galaxy) CreateCenterObject(sys *System) {

	// no huge object, continue with big
	n := sample(g.StellarSizeTypes.Big.NumCpm)
	//n := 1
	switch n {
	case 1:
		objidx := sample(g.StellarSizeTypes.Big.Cpm)
		//objidx := STAR
		//sys.CenterObject = CenterObject{}
		//		fmt.Println("Before CO creation:", sys)
		sys.CenterObject = new(CenterObject)
		sb := StellarObj{}
		//sb := new(StellarObj)

		//		fmt.Println("After CO creation:", sys.CenterObject, sys)
		switch objidx {
		case STAR:
			objidx = sample(g.StarTypes.Cpm)
			sb.Init(g.StarTypes.Types[objidx])
			sys.CenterObject.AddCenterObjectSingle(sb)
			//			fmt.Println("after object added ", sys.CenterObject, sys)
		case WD:
			objidx = sample(g.WDTypes.Cpm)
			sb.Init(g.WDTypes.Types[objidx])
			sys.CenterObject.AddCenterObjectSingle(sb)
		default:
			sb.Init(g.IMBHTypes.Types[0])
			sys.CenterObject.AddCenterObjectSingle(sb)
		}
		// single center object
	case 2:
		objidx := []int{sample(g.StellarSizeTypes.Big.Cpm), sample(g.StellarSizeTypes.Big.Cpm)}
		sys.CenterObject = new(CenterObject)
		//sys.CenterObject = CenterObject{}
		sb := new([2]StellarObj)
		//sb := [2]StellarObj{}
		for i, oi := range objidx {
			switch oi {
			case STAR:
				subidx := sample(g.StarTypes.Cpm)
				sb[i].Init(g.StarTypes.Types[subidx])
			case WD:
				subidx := sample(g.WDTypes.Cpm)
				sb[i].Init(g.WDTypes.Types[subidx])
			default:
				sb[i].Init(g.IMBHTypes.Types[0])
			}

		}
		sys.CenterObject.AddCenterObjectDouble(sb[0], sb[1])

		// double center object
	case 0:
		// TODO - change that to a planet once they are there
		n = sample(g.StellarSizeTypes.Huge.NumCpm)
		// Is center object a huge object?
		if n > 0 {
			//	sys.CenterObject = CenterObject{}
			sys.CenterObject = new(CenterObject)
			var sb StellarObj = StellarObj{}
			objidx := sample(g.StellarSizeTypes.Huge.Cpm)
			switch objidx {
			case IMBH:
				sb.Init(g.IMBHTypes.Types[0])
			case SGS:
				objidx = sample(g.OStarTypes.Cpm)
				sb.Init(g.OStarTypes.Types[objidx])
			default:
				sb.Init(g.OStarTypes.Types[0])
			}
			sys.CenterObject.AddCenterObjectSingle(sb)
		} else {
			// that would mean single planets, no suns
			// get 1-2 lonely planets
			// TODO!
			// at the moment only 1!
			//	sys.CenterObject = CenterObject{}
			sys.CenterObject = new(CenterObject)
			var sb StellarObj = StellarObj{}
			objidx := sample(g.StellarSizeTypes.Medium.Cpm)
			sb.Init(g.PlanetTypes.Types[objidx])
			sys.CenterObject.AddCenterObjectSingle(sb)

		}
	default:
		// multiple big objects, chaotic system
		//sys.CenterObject = CenterObject{}
		sys.CenterObject = new(CenterObject)
		sb := []StellarObj{}
		for i := 0; i < n; i++ {
			oi := sample(g.StellarSizeTypes.Big.Cpm)
			sb = append(sb, StellarObj{})
			switch oi {
			case STAR:
				subidx := sample(g.StarTypes.Cpm)
				sb[i].Init(g.StarTypes.Types[subidx])
			case WD:
				subidx := sample(g.WDTypes.Cpm)
				sb[i].Init(g.WDTypes.Types[subidx])
			default:
				sb[i].Init(g.IMBHTypes.Types[0])
			}

		}
		sys.CenterObject.AddCenterObjectMulti(sb)
	}

}

func (galaxy *Galaxy) LoadFromFile(filepath string) {
	// local struct def
	type Metadata struct {
		FileVersion int   `json:"file_version"`
		NumSystems  int32 `json:"num_systems"`
		RandSeed    int   `json:"rand_seed"`
	}

	type Stars struct {
		//Coords   Coordinates `json:"coords"`
		Coords   coords.CoordsI16 `json:"coords"`
		Lum      float64          `json:"lum"`
		Colorstr string           `json:"color"`
	}

	meta := Metadata{}
	metafile := "/galaxy.meta"
	starfile := "/galaxy.json"

	metatext := scanFile(filepath + metafile)
	// Get data from scan with Bytes() or Text()
	err := json.Unmarshal([]byte(metatext), &meta)
	if err != nil {
		log.Fatal(err)
	}
	//	galaxy.SysCount = meta.NumSystems
	fmt.Printf("File Version: %d, NumSystems %d, RandSeed %d\n", meta.FileVersion, meta.NumSystems, meta.RandSeed)

	file, err := os.Open(filepath + starfile)
	if err != nil {
		log.Fatal(err)
	}

	// start importing the Galaxy
	stars := make([]Stars, meta.NumSystems)

	byteValue, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &stars)

	for _, s := range stars {
		// look for max size
		pos := coords.CoordsI16{X: s.Coords.X, Y: s.Coords.Y, Z: s.Coords.Z}

		if s.Coords.X > galaxy.Radius {
			galaxy.Radius = s.Coords.X
		}
		if s.Coords.Z > galaxy.Radius {
			galaxy.Radius = s.Coords.Z
		}
		// add system.
		//newsys := galaxy.AddSystemAt(s.Coords)
		newsys := galaxy.AddSystemAt(pos)
		newsys.SetColor(s.Colorstr, s.Lum)
		//galaxy.Systems[len(galaxy.Systems)-1].SetColor(s.Colorstr, s.Lum)
	}
	fmt.Println("Galaxy radius(Target/actual): ", galaxy.RadiusTarget, galaxy.Radius)
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
	if !success {
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
