package galaxy

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	STAR = 0
	WD = 1
	SBH = 2
	SNS = 3
)

type Galaxy struct {
	Systems  []System
	SysCount int32

	SysTarget       int32
	RadiusTarget    int16
	ThicknessTarget int16

	ArmsMaxRad        float64
	ArmsOuterRad      int16
	ArmsInnerRad      int16
	ArmsEllipseFactor float64
	sysHash           map[int64]bool

	StellarSizeTypes SizeTypes
	StarTypes        StarTypes
	OStarTypes  	StarTypes
	WDTypes StarTypes
}

func (g *Galaxy) Init(SysTarget int32, RTarget int16, TTarget int16) {
	g.SysTarget = SysTarget
	g.RadiusTarget = RTarget
	g.ThicknessTarget = TTarget

	// arms config
	g.ArmsMaxRad = 2 * math.Pi
	g.ArmsInnerRad = 800
	g.ArmsOuterRad = 150
	g.ArmsEllipseFactor = 0.3
	fmt.Printf("Galaxy targets - Systems: %d, Diameter: %d, Thickness: %d\n", SysTarget, int32(RTarget)*2, TTarget)

	g.sysHash = make(map[int64]bool)

	g.StellarSizeTypes.ReadSizeTypeData("data/stellar_data.json")
	g.OStarTypes.ReadStarData("data/o_star_data.json")
	g.StarTypes.ReadStarData("data/star_data.json")
	g.WDTypes.ReadStarData("data/wd_data.json")

	fmt.Println(g.StellarSizeTypes)

	// testing the sample function
	/*
		samples := []int{0, 0, 0, 0, 0, 0}
		for i := 0; i < 1000; i++ {
			//samples[sample(g.StellarSizeTypes.Big.NumProbs)]++
			samples[sample(g.StellarSizeTypes.Big.NumCpm)]++
		}
		fmt.Println("Probabilities 'big'", samples, g.StellarSizeTypes.Big.NumProbs)
	*/
	// random init for reproducable tests - to be amended later
	rand.Seed(0)
}

// Create The galaxy content.
// there could be more different create() functions for different Galaxy forms
func (g *Galaxy) Create() {
	starttime := time.Now()
	g.Systems = []System{}

	//---- sin/cos test
	// this will be part of the save game
	rand.Seed(0)

	// Activate sin/cos mapping if number of stars > 40000

	if g.SysTarget >= 50000 {
		TrigBuf.Activate(5)
	}

	// create coordinates for the 'standard spiral' form
	g.CreateFormSpiral1()

	// ------------------ next steps
	// create kdtree()
	// create system contents_step1()
	// system content step 2 is only created once the system is 'used'
	fmt.Println("Creation took ", time.Since(starttime))

	starttime = time.Now()
	fmt.Println("Creating system (central) objects")

	for _, sys := range g.Systems {
		g.CreateCenterObject(sys)
	}
}

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

func (g *Galaxy) CreateCenterObject(sys System) {

	// no huge object, continue with big
	n := sample(g.StellarSizeTypes.Big.NumCpm)
	switch n {
	case 1:
		objidx := sample(g.StellarSizeTypes.Big.Cpm)
		sys.CenterObject = StellarObject{}

		
		switch objidx {
		case STAR:
			objidx = sample(g.StarTypes.Cpm)
			sys.CenterObject.InitBig(g.StarTypes.Types[objidx])
		case WD:
			objidx = sample(g.WDTypes.Cpm)
			sys.CenterObject.InitBig(g.WDTypes.Types[objidx])
		default:
			sys.CenterObject.InitHuge(g.StellarSizeTypes.Big.Types[objidx])
		}

		fmt.Println("--- Big Object / idx", sys.CenterObject, objidx, g.StellarSizeTypes.Big.Cpm)

		return
		// single center object
	case 2:
		// double center object
	case 0:
		// how many center objects of huge size?
		n = sample(g.StellarSizeTypes.Huge.NumCpm)
		// Is center object a huge object?
		if n > 0 {
			objidx := sample(g.StellarSizeTypes.Huge.Cpm)
			sys.CenterObject = StellarObject{}
			sys.CenterObject.InitHuge(g.StellarSizeTypes.Huge.Types[objidx])

			fmt.Println("- Huge Object:", sys.CenterObject)
			return
		}
		// else
		// get 1-2 lonely planets
	default:
		// multiple big objects, chaotic system

	}
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

// create a flattened disc of systems, adds it to the systems in the Galaxy struct
// ATTENTION: relThickness is the SIGMA of the Normvariate, so 31.x% of all values are OUTSIDE of the value
func (g *Galaxy) AddDisc(relRadius float64, relThickness float64, numStars int32) {
	sys := make([]System, numStars)
	radius := relRadius * float64(g.RadiusTarget)
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
		pc.L = rand.NormFloat64() * radius
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
	var i int32 = 1
	sys := make([]System, int(numStars))
	dupes := 0
	for armNum = 0; armNum < numArms; armNum++ {
		armAngle := 0
		armStartRad := g.ArmsMaxRad / float64(numArms) * float64(armNum)

		// paint stars
		//		sys := make([]System, numStars)
		for i < starsPerArm {
			coords := CoordsI16{}
			// offset for the star
			randomSphere := CoordsI16{}
			pc := CoordsPolar{}

			// set where on the way to the edge we are
			pc.L = radius / float64(starsPerArm) * float64(i)
			pc.A = float64(armAngle) + armStartRad
			pc.B = rand.Float64() * twopi
			//			fmt.Println("i", i, "radius", radius, "starsPerArm", starsPerArm, "armAngle", armAngle, "armStartRad", armStartRad, pc)

			coords.FromPolar(pc)

			// get random ball within a given radius for the star
			randomSphereRadius := ((1 - pc.L/radius) * armsDiff) + float64(g.ArmsOuterRad)

			pc.L = rand.NormFloat64() * randomSphereRadius
			pc.A = rand.Float64() * twopi
			pc.B = rand.Float64() * twopi
			randomSphere.FromPolar(pc)
			randomSphere.Z = int16(float64(randomSphere.Z) * g.ArmsEllipseFactor)

			coords.Add(randomSphere)
			//fmt.Println(coords)

			// ch3eck for dupes
			// create an index of z<<32+x<<16+y (unique ID)
			map_idx := int64(int64(coords.Z)<<32 + int64(coords.X)<<16 + int64(coords.Y))

			// if there isn't a system at that location
			if !g.sysHash[map_idx] {
				// set it as occupied
				g.sysHash[map_idx] = true
				// add to list
				sys[i].Coords = coords
			} else {
				// record dupe
				dupes++
			}
			// i gets incremented regardlessly, otherwise we get stuck in infinite loops
			i++
			// add to list
		}
	}
	fmt.Printf("%d Stars in Arms created(%d). dupes %d \n", i, len(sys), dupes)
	g.Systems = append(g.Systems, sys...)
	g.SysCount += numStars

}
