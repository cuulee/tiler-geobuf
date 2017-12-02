package gotile

// This part of the package consists of code to trim or envelope the polygons
// It also contains the subsequent code for creating polygon children.

import (
	//"fmt"
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	"github.com/paulmach/go.geojson"
	"math"
)

// function for getting the extrema of an alignment
// it also converts the points from [][][]float64 > pc.Polygon
// in other words the clipping data structure
func get_extrema_coords(coords [][][]float64) (m.Extrema, pc.Polygon) {
	north := -1000.
	south := 1000.
	east := -1000.
	west := 1000.
	lat := 0.
	long := 0.
	polygon := pc.Polygon{}

	// iterating through each outer ring
	for _, coord := range coords {
		cont := pc.Contour{}
		// iterating through each point in a ring
		for _, i := range coord {
			lat = i[1]
			long = i[0]

			if lat > north {
				north = lat
			}
			if lat < south {
				south = lat
			}
			if long > east {
				east = long
			}
			if long < west {
				west = long
			}
			//fmt.Print(long, lat, "\n")
			cont.Add(pc.Point{long, lat})

		}
		polygon.Add(cont)
	}

	return m.Extrema{S: south, W: west, N: north, E: east}, polygon

}

// gets the size of a tileid
func get_size(tile m.TileID) pc.Point {
	bds := m.Bounds(tile)
	return pc.Point{bds.E - bds.W, bds.N - bds.S}
}

// raw 1d linspace like found in numpy
func linspace(val1 float64, val2 float64, number int) []float64 {
	delta := (val2 - val1) / float64(number)
	currentval := val1
	newlist := []float64{val1}
	for currentval < val2 {
		currentval += delta
		newlist = append(newlist, currentval)
	}

	return newlist
}

// gets the middle of a tileid
func get_middle(tile m.TileID) pc.Point {
	bds := m.Bounds(tile)
	return pc.Point{(bds.E + bds.W) / 2.0, (bds.N + bds.S) / 2.0}
}

func grid_bounds(c2pt pc.Point, c4pt pc.Point, size pc.Point) m.Extrema {
	return m.Extrema{W: c2pt.X - size.X/2.0, N: c2pt.Y + size.Y/2.0, E: c4pt.X + size.X/2.0, S: c4pt.Y - size.Y/2.0}
}

// Overlaps returns whether r1 and r2 have a non-empty intersection.
func Within(big pc.Rectangle, small pc.Rectangle) bool {
	return (big.Min.X <= small.Min.X) && (big.Max.X >= small.Max.X) &&
		(big.Min.Y <= small.Min.Y) && (big.Max.Y >= small.Max.Y)
}

// a check to see if each point of a contour is within the bigger
func WithinAll(big pc.Contour, small pc.Contour) bool {
	totalbool := true
	for _, pt := range small {
		boolval := big.Contains(pt)
		if boolval == false {
			totalbool = false
		}
	}
	return totalbool
}

// creating a list with all of the intersecting contours
// this function returns a list of all the constituent contours as well as
// a list of their keys
func Sweep_Contmap(bb pc.Rectangle, intcont pc.Contour, contmap map[int]pc.Contour) []int {
	newlist := []int{}
	for k, v := range contmap {
		// getting the bounding box
		bbtest := v.BoundingBox()

		// getting within bool
		withinbool := Within(bb, bbtest)

		// logic for if within bool is true
		if withinbool == true {
			withinbool = WithinAll(intcont, v)
		}

		// logic for when we know the contour is within the polygon
		if withinbool == true {
			newlist = append(newlist, k)
		}
	}
	return newlist
}

// getting the outer keys of contours that will be turned into polygons
func make_polygon_list(totalkeys []int, contmap map[int]pc.Contour, relationmap map[int][]int) []pc.Polygon {
	keymap := map[int]string{}
	for _, i := range totalkeys {
		keymap[i] = ""
	}

	// making polygon map
	polygonlist := []pc.Polygon{}
	for k, v := range contmap {
		_, ok := keymap[k]
		if ok == false {
			newpolygon := pc.Polygon{v}
			otherconts := relationmap[k]
			for _, cont := range otherconts {
				newpolygon.Add(contmap[cont])
			}

			// finally adding to list
			polygonlist = append(polygonlist, newpolygon)
		}
	}
	return polygonlist

}

// creates a within map or a mapping of each edge
func Create_Withinmap(contmap map[int]pc.Contour) []pc.Polygon {
	totalkeys := []int{}
	relationmap := map[int][]int{}
	for k, v := range contmap {
		bb := v.BoundingBox()
		keys := Sweep_Contmap(bb, v, contmap)
		relationmap[k] = keys
		totalkeys = append(totalkeys, keys...)
	}

	return make_polygon_list(totalkeys, contmap, relationmap)
}

// lints each polygon
// takes abstract polygon rings that may contain polygon rings
// and returns geojson arranged polygon sets
func Lint_Polygons(polygon pc.Polygon) []pc.Polygon {
	contmap := map[int]pc.Contour{}
	for i, cont := range polygon {
		contmap[i] = cont
	}
	return Create_Withinmap(contmap)

}

// from a pc.Polygon representation (clipping representation)
// to a [][][]float64 representation
func Convert_Float(poly pc.Polygon) [][][]float64 {
	total := [][][]float64{}
	for _, cont := range poly {
		contfloat := [][]float64{}
		for _, pt := range cont {
			contfloat = append(contfloat, []float64{pt.X, pt.Y})
		}
		total = append(total, contfloat)
	}
	return total
}

// output structure to ensure everything stays in a key value stroe
type Output struct {
	Total [][][][]float64
	ID    m.TileID
}

// given a polygon to be tiled envelopes the polygon in corresponding boxes
func Env_Polygon(polygon *geojson.Feature, size int) map[m.TileID][]*geojson.Feature {
	// getting bds
	bds, poly := get_extrema_coords(polygon.Geometry.Polygon)

	// dummy values you know
	intval := 0
	tilemap := map[m.TileID][]int{}

	// getting all four corners
	c1 := pc.Point{bds.E, bds.N}
	c2 := pc.Point{bds.W, bds.N}
	c3 := pc.Point{bds.W, bds.S}
	c4 := pc.Point{bds.E, bds.S}

	// getting all the tile corners
	c1t := m.Tile(c1.X, c1.Y, size)
	c2t := m.Tile(c2.X, c2.Y, size)
	c3t := m.Tile(c3.X, c3.Y, size)
	c4t := m.Tile(c4.X, c4.Y, size)

	//tilemap := map[m.TileID][]int32{}
	tilemap[c1t] = append(tilemap[c1t], intval)
	tilemap[c2t] = append(tilemap[c2t], intval)
	tilemap[c3t] = append(tilemap[c3t], intval)
	tilemap[c4t] = append(tilemap[c4t], intval)
	sizetile := get_size(c1t)

	//c1pt := get_middle(c1t)
	c2pt := get_middle(c2t)
	//c3pt := get_middle(c3t)
	c4pt := get_middle(c4t)

	gridbds := grid_bounds(c2pt, c4pt, sizetile)
	//fmt.Print(gridbds, sizetile, "\n")
	sizepoly := pc.Point{bds.E - bds.W, bds.N - bds.S}
	xs := []float64{}
	if c2pt.X == c4pt.X {
		xs = []float64{c2pt.X}
	} else {
		xs = []float64{c2pt.X, c4pt.X}

	}
	ys := []float64{}
	if c2pt.Y == c4pt.Y {
		ys = []float64{c2pt.Y}
	} else {
		ys = []float64{c2pt.Y, c4pt.Y}

	}
	if sizetile.X < sizepoly.X {
		number := int((gridbds.E - gridbds.W) / sizetile.X)
		xs = linspace(gridbds.W, gridbds.E, number+1)
	}
	if sizetile.Y < sizepoly.Y {
		number := int((gridbds.N - gridbds.S) / sizetile.Y)
		ys = linspace(gridbds.S, gridbds.N, number+1)
	}

	//totallist := []string{}

	for _, xval := range xs {
		// iterating through each y
		for _, yval := range ys {
			tilemap[m.Tile(xval, yval, size)] = append(tilemap[m.Tile(xval, yval, size)], intval)
		}
	}
	c := make(chan Output)
	for k := range tilemap {
		newpoly := poly
		go func(newpoly pc.Polygon, k m.TileID, c chan Output) {
			newpoly2 := newpoly.Construct(pc.INTERSECTION, Make_Tile_Poly(k))
			polys := Lint_Polygons(newpoly2)
			total := [][][][]float64{}
			for _, p := range polys {
				total = append(total, Convert_Float(p))

			}
			c <- Output{Total: total, ID: k}
		}(newpoly, k, c)
	}
	totalmap := map[m.TileID][]*geojson.Feature{}
	properties := polygon.Properties
	for range tilemap {
		output := <-c
		if len(output.Total) > 0 {
			for _, coord := range output.Total {
				newgeom := geojson.Geometry{Type: "Polygon"}
				newgeom.Polygon = coord
				newfeat := geojson.Feature{Geometry: &newgeom, Properties: properties}
				totalmap[output.ID] = append(totalmap[output.ID], &newfeat)
			}
		}
	}

	return totalmap

}

// makes the tile polygon
func Make_Tile_Poly(tile m.TileID) pc.Polygon {
	bds := m.Bounds(tile)
	return pc.Polygon{{pc.Point{bds.E, bds.N}, pc.Point{bds.W, bds.N}, pc.Point{bds.W, bds.S}, pc.Point{bds.E, bds.S}}}
}

// area of bds (of a square)
func AreaBds(ext m.Extrema) float64 {
	return (ext.N - ext.S) * (ext.E - ext.W)
}

// given a polygon to be tiled envelopes the polygon in corresponding boxes
// from a polygon and a tileid return the tiles relating to the polygon 1 level lower
func Children_Polygon(polygon *geojson.Feature, tileid m.TileID) map[m.TileID][]*geojson.Feature {
	// getting bds
	bd, poly := get_extrema_coords(polygon.Geometry.Polygon)
	pt := poly[0][0]

	temptileid := m.Tile(pt.X, pt.Y, int(tileid.Z+1))
	bdtemp := m.Bounds(temptileid)

	// checking to see if the polygon lies entirely within a smaller childd
	if (bd.N <= bdtemp.N) && (bd.S >= bdtemp.S) && (bd.E <= bdtemp.E) && (bd.W >= bdtemp.W) {
		totalmap := map[m.TileID][]*geojson.Feature{}
		totalmap[temptileid] = append(totalmap[temptileid], polygon)
		return totalmap
	}

	// checking to see if the polygon is encompassed within a square
	bdtileid := m.Bounds(tileid)
	if (math.Abs(AreaBds(bdtileid)-AreaBds(bd)) < math.Pow(.000001,2.0)) && len(poly) == 1 && len(poly[0]) == 4 {
		//fmt.Print("here\n")
		totalmap := map[m.TileID][]*geojson.Feature{}

		tiles := m.Children(tileid)
		for _, k := range tiles {
			//poly := Make_Tile_Poly(k)
			bds := m.Bounds(k)
			poly := [][][]float64{{{bds.E, bds.N}, {bds.W, bds.N}, {bds.W, bds.S}, {bds.E, bds.S}}}
			newgeom := geojson.Geometry{Type: "Polygon", Polygon: poly}

			totalmap[k] = append(totalmap[k], &geojson.Feature{Geometry: &newgeom, Properties: polygon.Properties})
		}

		return totalmap

	}

	//fmt.Print("\r", len(polygon.Geometry.Polygon[0]))

	c := make(chan Output)
	// creating the 4 possible children tiles
	// and sending into a go function
	tiles := m.Children(tileid)
	for _, k := range tiles {
		newpoly := poly
		go func(newpoly pc.Polygon, k m.TileID, c chan Output) {
			newpoly2 := newpoly.Construct(pc.INTERSECTION, Make_Tile_Poly(k))
			polys := Lint_Polygons(newpoly2)
			total := [][][][]float64{}
			for _, p := range polys {
				total = append(total, Convert_Float(p))

			}
			c <- Output{Total: total, ID: k}
		}(newpoly, k, c)
	}
	totalmap := map[m.TileID][]*geojson.Feature{}
	properties := polygon.Properties
	for range tiles {
		output := <-c
		if len(output.Total) > 0 {
			for _, coord := range output.Total {
				newgeom := geojson.Geometry{Type: "Polygon"}
				newgeom.Polygon = coord
				newfeat := geojson.Feature{Geometry: &newgeom, Properties: properties}
				totalmap[output.ID] = append(totalmap[output.ID], &newfeat)
			}
		}
	}

	return totalmap

}



