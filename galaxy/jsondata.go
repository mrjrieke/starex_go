package galaxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
)

//-------------- HELPER FUNCTIONS -----------------

// get commulative probability map from json provided probabilities
func getCPM(origprobs []int) []float64 {
	// there might be a non-altering way, but I couldn't get it working

	cpm := make([]float64, len(origprobs))
	// if input is empty, return empty slice
	if len(origprobs) == 0 {
		return cpm
	}

	probs := make([]int, len(origprobs))
	copy(probs, origprobs)

	for i := 1; i < len(probs); i++ {
		probs[i] += probs[i-1]
	}
	sum_vals := probs[len(probs)-1]

	for i := 0; i < len(probs); i++ {
		cpm[i] = float64(probs[i]) / float64(sum_vals)
	}
	return cpm
}

// gives you a random index index given on the probabilities given in
// cpm (almost a cdf, it's int and not float)
// cdf = cumulative distribution function
func sample(cdf []float64) int {
	bucket := 0
	r := rand.Float64()
	for r > cdf[bucket] {
		bucket++
	}
	return bucket
}

func getJSONFile(fname string) []byte {
	jsonFile, err := os.Open(fname)
	if err != nil {
		fmt.Println("ERROR Opening json file", err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue
}

// ------------------- Stellar_data.json ---------------
type SizeTypeData struct {
	Type        string `json:"type"`
	Description string
	Mass_min    int64
	Mass_max    int64
	// probability of how likely that particular type is for it's SizeType
	Prob        int
	Luminosity  float64
	Temperature int32
	Radius      int32
	Color       string
	Color_code  string
}

type SizeType struct {
	Description string
	// Probabilites of how many of these exist in a system(starting with 0)
	NumProbs []int     `json:"num_probs"`
	NumCpm   []float64 // cummulative probability map
	Types    []SizeTypeData
	Probs    []int
	Cpm      []float64
}

func (st *SizeType) GetProbs() {
	st.Probs = make([]int, len(st.Types))
	for i, t := range st.Types {
		st.Probs[i] = t.Prob
	}
	fmt.Println(st.Description, " - Probs:", st.Probs)
}

/*
func (st *SizeType) GetRandomObject() StellarObject {
	retO := StellarObject{}
	retT := st.Types[sample(st.Cpm)]

	retO.Init(retT)

}
*/

type SizeTypes struct {
	Massive SizeType
	Huge    SizeType
	Big     SizeType
	Medium  SizeType
	// TODO: Everything from here on downwards is still broken
	Small    SizeType
	DS_ptype SizeType `json:"ds-ptype"`
	DS_stype SizeType `json:"ds-stype"`
}

func (st *SizeTypes) ReadSizeTypeData(fname string) {
	//allSizeTypes := SizeTypes{}

	byteValue := getJSONFile(fname)
	//	json.Unmarshal([]byte(byteValue), &allSizeTypes)
	json.Unmarshal([]byte(byteValue), st)

	// This is very ugly
	//stypelist := [7]*SizeType{&allSizeTypes.Massive, &allSizeTypes.Huge,
	//	&allSizeTypes.Big, &allSizeTypes.Medium, &allSizeTypes.Small,
	//	&allSizeTypes.DS_ptype, &allSizeTypes.DS_stype}
	stypelist := [7]*SizeType{&st.Massive, &st.Huge,
		&st.Big, &st.Medium, &st.Small,
		&st.DS_ptype, &st.DS_stype}

	// get CPM for number of objects in system
	for _, st := range stypelist {
		st.NumCpm = getCPM(st.NumProbs)
		st.GetProbs()
		st.Cpm = getCPM(st.Probs)
	}

	//	fmt.Println(allSizeTypes)
	//	st.Types = append(st.Types, result...)

	//	return allSizeTypes
}

/*
func (st *SizeTypes) GetRandomCenterObject() (*SizeType, int) {
	for _, st := range [3]*SizeType{&st.Huge, &st.Big, &st.Medium} {
		n := sample(st.NumCpm)
//		fmt.Println("nums:", sample(st.NumCpm))
		if n > 0 {
			return st, n
		}

	}
	return nil,0
}
*/

// ------------------- Star_data.json ---------------
type StarData struct {
	Sequence    int
	Description string
	Type        string `json:"type"`
	Temperature int32
	Radius      float64
	Mass        float64
	Luminosity  float64
	HabZone     float64 `json:"hab_zone"`
	Abundance   float64
	Color       string `json:"color_code"`
	Prob        int
	RandMin     int32 `json:"rand_min"`
	RandMax     int32 `json:"rand_max"`
}

type StarTypes struct {
	//	Description string
	// Probabilites of how many of these exist in a system(starting with 0)
	Types []StarData
	Probs []int
	Cpm   []float64
}

func (st *StarTypes) GetProbs() {
	st.Probs = make([]int, len(st.Types))
	for i, t := range st.Types {
		st.Probs[i] = t.Prob
	}
}

func (st *StarTypes) ReadStarData(fname string) {
	jsonFile, err := os.Open(fname)
	if err != nil {
		fmt.Println("ERROR Opening json file", err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	result := []StarData{}
	json.Unmarshal([]byte(byteValue), &result)

	st.Types = append(st.Types, result...)

	// get CPM for number of objects in system
	st.GetProbs()
	st.Cpm = getCPM(st.Probs)
	fmt.Println("Star Probs:", st.Probs, st.Cpm)

}
