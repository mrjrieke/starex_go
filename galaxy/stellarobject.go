package galaxy


const (
	UNDEFINED = 0
	MASSIVE = 10
	HUGE = 11
	BIG = 12
	MEDIUM = 13
	SMALL = 14

	SINGLE = 20
	DOUBLE = 21
	MULTIPLE = 22
)

type CenterObject struct {
	//	Init()
	//	AddCenterObject()
	coType     int
	color      string
	luminosity float64
	CentralSO  []*StellarObj
}

func (co *CenterObject) AddCenterObjectSingle(so StellarObj) {
	co.coType = SINGLE
	co.CentralSO = []*StellarObj {&so}
//	co.coType = so.Type
	co.color = so.Color
	co.luminosity = so.Luminosity
}

func (co *CenterObject) AddCenterObjectDouble(so1 StellarObj, so2 StellarObj) {
	co.coType = DOUBLE
	co.CentralSO = []*StellarObj {&so1,&so2}
	// TODO - change this!
	if so1.Luminosity > so2.Luminosity {
		co.color = so1.Color
	} else {
		co.color = so2.Color
	}
	co.luminosity = so1.Luminosity + so2.Luminosity
}

func (co *CenterObject) AddCenterObjectMulti(mso []StellarObj) {
	co.coType = MULTIPLE

	var maxlum float64
	var maxi int
	_ = maxi
	// create empty slice
	co.CentralSO = []*StellarObj {}
	// add to it
	for i, so := range mso {
		co.CentralSO = append(co.CentralSO, &so)
		co.luminosity += so.Luminosity
		if so.Luminosity > maxlum {
			maxlum = so.Luminosity
			maxi = i
		}
	}
	// TODO - this should be a summary of all colors, not the max
	co.color = co.CentralSO[maxi].Color

}

func (co *CenterObject) Color() string {
	return co.color
}
func (co *CenterObject) Lum() float64 {
	return co.luminosity
}
func (co *CenterObject) Type() int {
	return co.coType
}


type StellarObj struct {
	sizeCategory   int8
	Type       string
	Category   int8
	Color      string
	Luminosity float64
	starData *StarData
}

func (so *StellarObj) Init(st StarData) {
	so.starData = &st
	so.Color = st.Color
	so.Luminosity = st.Luminosity
	so.Type = st.Type
}