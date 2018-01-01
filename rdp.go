package gotile

import "math"
import "github.com/paulmach/go.geojson"


// RDPSimplify is an in-place implementation of Ramer–Douglas–Peucker.
func RDPSimplify(points [][]float64, epsilon float64) [][]float64 {
	return points[:rdpCompress(points, epsilon)]
}

// simplifies a given geometry
func RDP_Line(line [][]float64,zoom int) [][]float64 {
	simpl := math.Pow(10,-(math.Ceil((float64(zoom) * math.Ln2 + math.Log(256.0 / 360.0 / 0.5)) / math.Ln10)))
	return RDPSimplify(line,simpl)
}

// simplifies a polygon
func RDP_Polygon(polygon [][][]float64,zoom int) [][][]float64 {
	newlist := [][][]float64{}
	for _,cont := range polygon {
		cont = RDP_Line(cont,zoom)
		if len(cont) > 3 {
			newlist = append(newlist,cont)
		}
	}
	return newlist
}

func RDP(geom *geojson.Geometry,zoom int) *geojson.Geometry {
	if geom.Type == "Point" {
		return geom
	} else if geom.Type == "LineString" {
		geom.LineString = RDP_Line(geom.LineString,zoom)
		if len(geom.LineString) <= 1 {
			geom.Type = ""
		}

	} else if geom.Type == "Polygon" {
		geom.Polygon = RDP_Polygon(geom.Polygon,zoom)
		if len(geom.Polygon) == 0 {
			geom.Type = ""
		}
	}
	return geom
}


func rdpCompress(points [][]float64, epsilon float64) int {
	end := len(points)

	if end < 3 {
		// return points
		return end
	}

	// Find the point with the maximum distance
	var (
		first = points[0]
		last  = points[end-1]

		flDist  = distance(first, last)
		flCross = first[0]*last[1] - first[1]*last[0]

		dmax  float64
		index int
	)

	for i := 2; i < end-1; i++ {
		d := perpendicularDistance(points[i], first, last, flDist, flCross)
		if d > dmax {
			dmax, index = d, i
		}
	}

	// If max distance is lte to epsilon, return segment containing
	// the first and last points.
	if dmax <= epsilon {
		// return []point{first, last}
		points[1] = last
		return 2
	}

	// Recursive call
	r1 := rdpCompress(points[:index+1], epsilon)
	r2 := rdpCompress(points[index:], epsilon)

	// Build the result list
	// return append(r1[:len(r1)-1], r2...)
	x := r1 - 1
	n := copy(points[x:], points[index:index+r2])

	return x + n
}

func distance(a, b []float64) float64 {
	x := a[0] - b[0]
	y := a[1] - b[1]
	return math.Sqrt(x*x + y*y)
}

func perpendicularDistance(p, fp, lp []float64, dist, cross float64) float64 {
	return math.Abs(cross+
		lp[0]*p[1]+p[0]*fp[1]-
		p[0]*lp[1]-fp[0]*p[1]) / dist
}
