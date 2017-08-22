package util

import (
	"math"

	"engo.io/engo"
)

// SubtractPoints finds the difference of the X and Y values of p1 and p2
// This is needed even though engo has the engo.Point.Subtract function because
// this doesn't change the underlying value
func SubtractPoints(p1, p2 engo.Point) engo.Point {
	var newPoint engo.Point
	newPoint.X = p1.X - p2.X
	newPoint.Y = p1.Y - p2.Y
	return newPoint
}

// AddDegrees adds delta to degrees, and keeps the return value between 0 and 360
// It acts as if degrees is 'continious', and when you go over 360 you go to 0,
// when you go under 0 you go back to 360
func AddDegrees(degrees float32, delta float32) float32 {
	degrees += delta
	if degrees > 360 {
		factorOver := int(math.Ceil(float64(degrees))) / 360
		degrees = degrees - float32(factorOver*360)
	} else if degrees < 0 {
		factorUnder := int(math.Ceil(math.Abs(float64(degrees)))) / 360
		degrees += float32(factorUnder+1) * 360
	}
	return degrees
}
