 package galaxy

 import (
	 "fmt"
 )

 type System struct {
	Coords  CoordsI16
	CenterObject StellarObject
 }

 func (s *System) print() {
	 fmt.Printf("System - Coords %v", s.Coords)
 }

 func (s *System) PlaceCenterObject() {
		
 }