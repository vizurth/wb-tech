package point

import "math"

type Point struct {
	x, y float64
}

func New(x, y float64) *Point {
	return &Point{x, y}
}

func (p *Point) Distance(q *Point) float64 {
	return math.Sqrt(math.Pow(p.x-q.x, 2) + math.Pow(p.y-q.y, 2))
}
