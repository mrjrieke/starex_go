package galaxy

import (
	//"fmt"
)

type StellarObject struct {
	Type       string
	Category   int8
	Color      string
	Luminosity float64
	Sattelites []StellarObject
}

func (so *StellarObject) InitHuge(st SizeTypeData) {
	so.Color = st.Color
	so.Luminosity = st.Luminosity
	so.Type = st.Type
}

func (so *StellarObject) InitBig(st StarData) {
	so.Color = st.Color
	so.Luminosity = st.Luminosity
	so.Type = st.Type
}

