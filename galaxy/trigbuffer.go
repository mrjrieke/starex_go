package galaxy

import (
	"fmt"
	"math"
	"time"
)


type TrigBuffer struct {
	Active     bool
	sin        []float64
	cos        []float64
	precision  int
	maxval     int
	twopi      float64
	multiplier float64
}

// buffers Sin and Cos values to a given precision
func (tb *TrigBuffer) Activate(prec int) {
	starttime := time.Now()
	tb.twopi = 2 * math.Pi
	// maximum 10^5 entries a' 8 bytes = 800.000 bytes
	maxPrecision := 5
	if prec > maxPrecision {
		tb.precision = maxPrecision
	} else {
		tb.precision = prec
	}
	tb.maxval = int(math.Pow10(int(tb.precision)))
	// make sin buffer
	tb.sin = make([]float64, tb.maxval)
	tb.cos = make([]float64, tb.maxval)
	increment := math.Pi * 2 / float64(tb.maxval)
	var s float64
	var c float64
	var x float64
	for i := 0; i < tb.maxval; i++ {
		s = math.Sin(x)
		c = math.Cos(x)
		x += increment
//		ind := int(x * float64(tb.maxval))
		tb.sin[i] = s 
		tb.cos[i] = c
//		fmt.Printf("%d, %f, %f |", ind, x, y)
	}
	//fmt.Println("\nincrement:", increment, len(tb.sin), "\n-----------------")
	tb.multiplier = float64(tb.maxval) / tb.twopi
	fmt.Println("Trigometric Buffer creation took ", time.Since(starttime))
	tb.Active = true
}

func (tb *TrigBuffer) Sin(alpha float64) float64 {
	ind := int(alpha * tb.multiplier)
	return tb.sin[ind]
}
func (tb *TrigBuffer) Cos(alpha float64) float64 {
	ind := int(alpha * tb.multiplier)
	return tb.cos[ind]
}
