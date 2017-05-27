package util

import "engo.io/engo"

// SubtractPoints finds the difference of the X and Y values of p1 and p2
func SubtractPoints(p1, p2 engo.Point) engo.Point {
	var newPoint engo.Point
	newPoint.X = p1.X - p2.X
	newPoint.Y = p1.Y - p2.Y
	return newPoint
}
