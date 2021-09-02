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
)

type Galaxy struct {
	Systems  []System
	SysCount int32

	SysTarget       int32
	RadiusTarget    int16 // what radius are we aiming for
	ThicknessTarget int16

	Radius int16 // acual radius (highest absolute x||y value)

	ArmsMaxRad        float64
	ArmsOuterRad      int16
	ArmsInnerRad      int16
	ArmsEllipseFactor float64
	sysHash           map[int64]bool

	StellarSizeTypes SizeTypes
	StarTypes        StarTypes
	OStarTypes       StarTypes
	WDTypes          StarTypes
	IMBHTypes        StarTypes
}

func (g *Galaxy) Init() {
	// arms config
	g.ArmsMaxRad = 2 * math.Pi
	g.ArmsInnerRad = 800
	g.ArmsOuterRad = 150
	g.ArmsEllipseFactor = 0.3

	g.sysHash = make(map[int64]bool)

	g.StellarSizeTypes.ReadSizeTypeData("data/stellar_data.json")
	g.OStarTypes.ReadStarData("data/o_star_data.json")
	g.WDTypes.ReadStarData("data/wd_data.json")

	// not sure if WD and O star data should go in here or not
	g.StarTypes.ReadStarData("data/o_star_data.json")
	g.StarTypes.ReadStarData("data/star_data.json")
	g.StarTypes.ReadStarData("data/wd_data.json")

	//	fmt.Println("STAR TYPES", g.StarTypes)
	g.IMBHTypes.ReadStarData("data/imbh_data.json")

	//	fmt.Println("IMBH TYPES", g.IMBHTypes)

	g.Systems = []System{}

	rand.Seed(0)
}

// Create The galaxy content.
// in here we will branch off into the randomizer and the various galaxy forms to create
//func (g *Galaxy) Create() {
func (g *Galaxy) Create(SysTarget int32, RTarget int16, TTarget int16) {
	starttime := time.Now()
	g.SysTarget = SysTarget
	g.RadiusTarget = RTarget
	g.ThicknessTarget = TTarget

	rand.Seed(0)

	// Activate sin/cos mapping if number of stars > 40000
	//	if g.SysTarget >= 50000 {
	//	TrigBuf.Activate(5)
	//	}

	// create coordinates for the 'standard spiral' form
	g.CreateFormSpiral1()
	//g.CreateForm2()

	// ------------------ next steps
	// create kdtree()
	// create system contents_step1()
	// system content step 2 is only created once the system is 'used'
	fmt.Println("Creation took ", time.Since(starttime))

	starttime = time.Now()
	fmt.Println("Creating system (central) objects")

	for i := range g.Systems {
		//		sys.CenterObject = &CenterObjectSingle{}
		//		var sb StellarObjI
		g.CreateCenterObject(&g.Systems[i])
		g.Systems[i].SetColor(g.Systems[i].CenterObject.Color(), g.Systems[i].CenterObject.Lum())
		if g.Systems[i].Coords.X > g.Radius {
			g.Radius = g.Systems[i].Coords.X
		}
		if g.Systems[i].Coords.Z > g.Radius {
			g.Radius = g.Systems[i].Coords.Z
		}

		if g.Systems[i].Color.R == 200 {
			fmt.Println("CO:", g.Systems[i].CenterObject)
		}
	}

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
	sys := make([]System, numStars)
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
		coords := CoordsI16{}
		pc := CoordsPolar{}
		// ! ATTENTION ! radius is the sigma of th this could lead to values > maxint16 - should we check for this before? this could lead to values > maxint16 - should we check for this before? this could lead to values > maxint16 - should we check for this before?e normal variate. So 31.x% of all the stars are OUTSIDE the radius
		// radius is th
		pc.L = rand.NormFloat64() * halfRadius
		pc.A = rand.Float64() * twopi
		pc.B = rand.Float64() * twopi
		// convert to real coords
		// !QUESTION!  this could lead to values > maxint16 - should we check for this before?
		coords.FromPolar(pc)
		// flatten ball
		coords.Z = int16(float64(coords.Z) / flatten)

		// create an index of z<<32+x<<16+y (unique ID)
		map_idx := int64(int64(coords.Z)<<32 + int64(coords.X)<<16 + int64(coords.Y))

		// if there isn't a system at that location
		if !g.sysHash[map_idx] {
			// set it as occupied
			g.sysHash[map_idx] = true
			// add to list
			sys[s].Coords = coords
			/// and move on
			s++
		} else {
			// record dupe
			dupes++
		}

		// TODO: check dupes
	}
	fmt.Printf("%d Stars in disc created (%d). dupes: %d \n", s, len(sys), dupes)
	g.Systems = append(g.Systems, sys...)
	//	fmt.Println(len(g.Systems), g.Systems)
	g.SysCount += numStars
}

// create a set of spiral arms and adds it
func (g *Galaxy) AddArms(relRadius float64, numStars int32, numArms int32) {
	radius := relRadius * float64(g.RadiusTarget)
	twopi := 2 * math.Pi
	starsPerArm := numStars / numArms
	armsDiff := float64(g.ArmsInnerRad - g.ArmsOuterRad)

	fmt.Println("Target stars: ", numStars)
	// per arm
	var armNum int32
	var sysCount int32
	sys := make([]System, int(numStars))
	dupes := 0
	angleIncrease := radius / float64(starsPerArm)
	for armNum = 0; armNum < numArms; armNum++ {
		var i float64 = 1
		var armAngle float64
		armStartRad := g.ArmsMaxRad / float64(numArms) * float64(armNum)
		// paint stars
		for i < radius {
			coords := CoordsI16{}
			// offset for the star
			randomSphere := CoordsI16{}
			pc := CoordsPolar{}

			// set where on the way to the edge we are
			pc.L += i
			pc.A = armAngle + armStartRad
			coords.FromPolar(pc)

			// get random ball within a given radius for the star
			randomSphereRadius := ((1 - pc.L/radius) * armsDiff) + float64(g.ArmsOuterRad)

			pc.L = rand.NormFloat64() * randomSphereRadius
			pc.A = rand.Float64() * twopi
			randomSphere.FromPolar(pc)

			// flatten the ball
			randomSphere.Z = int16(float64(randomSphere.Z) * g.ArmsEllipseFactor)

			// Add random sphere to
			coords = coords.Add(randomSphere)

			// ch3eck for dupes
			// create an index of z<<32+x<<16+y (unique ID)
			map_idx := int64(int64(coords.Z)<<32 + int64(coords.X)<<16 + int64(coords.Y))

			// if there isn't a system at that location
			if !g.sysHash[map_idx] {
				// set it as occupied
				g.sysHash[map_idx] = true
				// add to list
				sys[sysCount].Coords = coords
				//sys[i].Coords = coords
				sysCount += 1
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
	g.Systems = append(g.Systems, sys...)
	g.SysCount += numStars

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
	sys := make([]System, int(numStars))
	dupes := 0
	angleIncrease := radius / float64(numStars)
	//var i int32 = 1
	var i float64 = 1
	var armAngle float64

	for i < radius {
		coords := CoordsI16{}
		// offset for the star
		randomSphere := CoordsI16{}
		pc := CoordsPolar{}

		// set where on the way to the edge we are
		pc.L += i
		pc.B = rand.Float64() * twopi
		pc.A = float64(armAngle)

		coords.FromPolar(pc)
		coords.Z = int16(float64(coords.Z) * flatten)

		// get random ball within a given radius for the star
		randomSphereRadius := ((1 - pc.L/radius) * armsDiff) + float64(g.ArmsOuterRad)

		pc.L = rand.NormFloat64() * randomSphereRadius
		pc.A = rand.Float64() * twopi
		pc.B = rand.Float64() * twopi
		randomSphere.FromPolar(pc)

		// flatten the ball
		randomSphere.Z = int16(float64(randomSphere.Z))

		// Add random sphere to
		coords = coords.Add(randomSphere)

		// ch3eck for dupes
		// create an index of z<<32+x<<16+y (unique ID)
		map_idx := int64(int64(coords.Z)<<32 + int64(coords.X)<<16 + int64(coords.Y))

		// if there isn't a system at that location
		if !g.sysHash[map_idx] {
			// set it as occupied
			g.sysHash[map_idx] = true
			// add to list
			sys[sysCount].Coords = coords
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
	g.Systems = append(g.Systems, sys...)
	g.SysCount += numStars

}

func (g *Galaxy) CreateCenterObject(sys *System) {

	// no huge object, continue with big
	n := sample(g.StellarSizeTypes.Big.NumCpm)
	switch n {
	case 1:
		objidx := sample(g.StellarSizeTypes.Big.Cpm)
		sys.CenterObject = CenterObject{}
		sb := StellarObj{}

		switch objidx {
		case STAR:
			objidx = sample(g.StarTypes.Cpm)
			sb.Init(g.StarTypes.Types[objidx])
			sys.CenterObject.AddCenterObjectSingle(sb)
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
		sys.CenterObject = CenterObject{}
		sb := [2]StellarObj{}
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
		//n = sample(g.StellarSizeTypes.Medium.NumCpm)
		// Is center object a huge object?
		if n > 0 {
			//fmt.Println("Huge cpm:", g.StellarSizeTypes.Huge.Cpm)
			sys.CenterObject = CenterObject{}
			var sb StellarObj = StellarObj{}
//			objidx := sample(g.StellarSizeTypes.Huge.Cpm)
			//sb.Init(g.StarTypes.Types[objidx])
			sb.Init(g.IMBHTypes.Types[0])
			sys.CenterObject.AddCenterObjectSingle(sb)
			//			sb.Init(g.StellarSizeTypes.Big.Types[objidx])
			//			sb.Init(st.Color, st.Luminosity, st.Type)
			//sys.CenterObject = StellarObject{}
			//sys.CenterObject.InitHuge(g.StellarSizeTypes.Huge.Types[objidx])

			//	fmt.Println("- Huge Object:", sys.CenterObject)
			return
		} else {
		// that would mean single planets, no suns
		// get 1-2 lonely planets
		// TODO!
			sys.CenterObject = CenterObject{}
			var sb StellarObj = StellarObj{}
			sb.Init(g.IMBHTypes.Types[0])
			sys.CenterObject.AddCenterObjectSingle(sb)

		}
	default:
		// multiple big objects, chaotic system
		//		objidx := []int{sample(g.StellarSizeTypes.Big.Cpm), sample(g.StellarSizeTypes.Big.Cpm)}
		sys.CenterObject = CenterObject{}
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
	//	fmt.Println("Center Object 2", sys.CenterObject)
	/*
		for _, st := range [3]*SizeType{&st.Huge, &st.Big, &st.Medium} {
			n := sample(st.NumCpm)
			if n > 0 {
				return st, n
			}

		}
		return nil,0
	*/
}

func (galaxy *Galaxy) LoadFromFile(filepath string) {
	// local struct def
	type Metadata struct {
		FileVersion int   `json:"file_version"`
		NumSystems  int32 `json:"num_systems"`
		RandSeed    int   `json:"rand_seed"`
	}
	// temp structs to read json file
	type Coordinates struct {
		X int16 `json:"x"`
		Y int16 `json:"y"`
		Z int16 `json:"z"`
	}
	type Stars struct {
		Coords   Coordinates `json:"coords"`
		Lum      float64     `json:"lum"`
		Colorstr string      `json:"color"`
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
	galaxy.SysCount = meta.NumSystems
	fmt.Printf("File Version: %d, NumSystems %d, RandSeed %d\n", meta.FileVersion, meta.NumSystems, meta.RandSeed)

	file, err := os.Open(filepath + starfile)
	if err != nil {
		log.Fatal(err)
	}

	// start importing the Galaxy
	stars := make([]Stars, meta.NumSystems)
	galaxy.Systems = make([]System, meta.NumSystems)
	//	stars := []Stars{}

	byteValue, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &stars)

	for i, s := range stars {
		if s.Coords.X > galaxy.Radius {
			galaxy.Radius = s.Coords.X
		}
		if s.Coords.Z > galaxy.Radius {
			galaxy.Radius = s.Coords.Z
		}
		//		fmt.Println(s.Coords)
		galaxy.Systems[i].SetColor(s.Colorstr, s.Lum)
		galaxy.Systems[i].Coords.X = s.Coords.X
		galaxy.Systems[i].Coords.Y = s.Coords.Y
		galaxy.Systems[i].Coords.Z = s.Coords.Z
		//		stars[i].setColor()
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
