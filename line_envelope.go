package gotile

import (
	"encoding/json"
	"fmt"
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	"github.com/paulmach/go.geojson"
	"math"
	"sort"
	"strings"
)

// gets the slope of two pc.Points along a line
// if statement logic accounts for undefined corner case
func get_slope2(pt1 pc.Point, pt2 pc.Point) float64 {
	if pt1.X == pt2.X {
		return 1000000.0
	}
	return (pt2.Y - pt1.Y) / (pt2.X - pt1.X)
}

// pc.Point represents a pc.Point in space.
type Size2 struct {
	deltaX float64
	deltaY float64
}

// iteroplates the position of y based on x of the location between two pc.Points
// this function accepts m the slope to keep it from recalculating
// what could be several hundred/thousand times between two pc.Points
func interp2(pt1 pc.Point, pt2 pc.Point, x float64) pc.Point {
	m := get_slope2(pt1, pt2)
	y := (x-pt1.X)*m + pt1.Y
	return pc.Point{x, y}
}

type ResponseCoords2 struct {
	Coords [][][]float64 `json:"coords"`
}

// gets the coordstring into a slice the easiest way I'm aware of
func get_coords_json2(stringcoords string) [][][]float64 {
	stringcoords = fmt.Sprintf(`{"coords":%s}`, stringcoords)
	res := ResponseCoords2{}
	json.Unmarshal([]byte(stringcoords), &res)

	return res.Coords
}

// distance between two points
func distance_pts(oldpt pc.Point, pt pc.Point) Size2 {
	return Size2{math.Abs(pt.X - oldpt.X), math.Abs(pt.Y - oldpt.Y)}

}

// distance of bounds
func distance_bounds(bds m.Extrema) Size2 {
	return Size2{bds.E - bds.W, bds.N - bds.S}
}

// decides which plane something itersects with
func which_plane(oldpt pc.Point, pt pc.Point, oldbds m.Extrema) string {
	xs := []float64{oldpt.X, pt.X}
	sort.Float64s(xs)

	ys := []float64{oldpt.Y, pt.Y}
	sort.Float64s(ys)

	mybds := m.Extrema{W: xs[0], E: xs[1], S: ys[0], N: ys[1]}

	if (mybds.N >= oldbds.N) && (mybds.S <= oldbds.N) {
		return "north"
	} else if (mybds.N >= oldbds.S) && (mybds.S <= oldbds.S) {
		return "south"
	} else if (mybds.E >= oldbds.E) && (mybds.W <= oldbds.E) {
		return "east"
	} else if (mybds.E >= oldbds.W) && (mybds.W <= oldbds.W) {
		return "west"
	} else {
		return "NONE"
	}
}

// given a pc.Point checks to see if the given pt is within the correct bounds
func check_bounds(oldpt pc.Point, pt pc.Point, intersectpt pc.Point, oldbds m.Extrema) bool {
	if (intersectpt.X >= oldbds.W) && (intersectpt.X <= oldbds.E) && (intersectpt.Y >= oldbds.S) && (intersectpt.Y <= oldbds.N) && (check_bb(oldpt, pt, intersectpt) == true) {
		//fmt.Print(check_bb(oldpt, pt, intersectpt), "\n")
		return true
	} else {
		return false
	}
}

// finding the pc.Point that intersects with a given y
func opp_interp(pt1 pc.Point, pt2 pc.Point, y float64) pc.Point {
	m := get_slope2(pt1, pt2)
	x := ((y - pt1.Y) / m) + pt1.X
	return pc.Point{x, y}
}

// checks a boudning box
func check_bb(oldpt pc.Point, pt pc.Point, intersectpt pc.Point) bool {
	xs := []float64{oldpt.X, pt.X}
	sort.Float64s(xs)

	ys := []float64{oldpt.Y, pt.Y}
	sort.Float64s(xs)
	//fmt.Print(xs, ys, "\n")
	if (intersectpt.X >= xs[0]) && (intersectpt.X <= xs[1]) && (intersectpt.Y >= ys[0]) && (intersectpt.Y <= ys[1]) {
		return true
	} else {
		return false
	}

}

// this function gets the intersection pc.Point with a bb box
// it also returns a string of the axis it intersected with
func get_intersection_pt(oldpt pc.Point, pt pc.Point, oldbds m.Extrema) (pc.Point, string) {
	trypt := interp2(oldpt, pt, oldbds.W)
	axis := "west"
	//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)

	if check_bounds(oldpt, pt, trypt, oldbds) == false {
		trypt = interp2(oldpt, pt, oldbds.E)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)

		axis = "east"
	}
	if check_bounds(oldpt, pt, trypt, oldbds) == false {
		trypt = opp_interp(oldpt, pt, oldbds.S)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
		axis = "south"
	}
	if check_bounds(oldpt, pt, trypt, oldbds) == false {
		trypt = opp_interp(oldpt, pt, oldbds.N)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)

		axis = "north"
	}
	if axis == "north" {
		trypt = pc.Point{0, 0}
	}

	return trypt, axis
}

// gets an intersection point
func itersection_pt(oldpt pc.Point, pt pc.Point, oldbds m.Extrema, axis string) []float64 {
	//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
	if axis == "west" {
		trypt := interp2(oldpt, pt, oldbds.W)
		return []float64{trypt.X, trypt.Y}
	} else if axis == "east" {
		trypt := interp2(oldpt, pt, oldbds.E)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
		return []float64{trypt.X, trypt.Y}

	} else if axis == "south" {
		trypt := opp_interp(oldpt, pt, oldbds.S)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
		return []float64{trypt.X, trypt.Y}

	} else if axis == "north" {
		trypt := opp_interp(oldpt, pt, oldbds.N)
		return []float64{trypt.X, trypt.Y}
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
	}
	return []float64{0, 0}
}

// convert the lines representing tile coods into lines readable by
// nlgeojson
func convert_tile_coords(total [][]pc.Point) {
	count := 0
	var totalstring []string
	for _, line := range total {
		totalstring = []string{}
		for _, pt := range line {
			totalstring = append(totalstring, fmt.Sprintf("[%f,%f]", pt.X, pt.Y))
		}
		//fmt.Printf(`%d,"[%s]"`, count, strings.Join(totalstring, ","))
		//fmt.Print("\n")
		count += 1
	}

}

// is the number even
func Even(number int) bool {
	return number%2 == 0
}

// is the number odd?
func Odd(number int) bool {
	return !Even(number)
}

// gets the axis opp from the last 
// LOGIC := if were crossing one tile and going into another
// whatever plane we cross is opposite of the previous
func opp_axis(val string) string {
	if val == "north" {
		return "south"
	} else if val == "south" {
		return "north"
	} else if val == "west" {
		return "east"
	} else if val == "east" {
		return "west"
	}

	return val
}
// 
func Get_Centroid_TileID(coords [][]float64,zoom int) m.TileID {
	totalx,totaly := 0.0,0.0
	for _,pt := range coords {
		totalx += pt[0]
		totaly += pt[1]
	}
	centx,centy := totalx/float64(len(coords)),totaly/float64(len(coords))
	return m.Tile(centx,centy,zoom)
}	



// functionifying this section so it doesnt get massive pretty decent break point
func Env_Line(line *geojson.Feature, zoom int) map[m.TileID][]*geojson.Feature {
	// intializes variables
	var oldpt pc.Point
	var tileid, oldtileid m.TileID
	var bds, oldbds m.Extrema
	var axis string
	var tilecoords [][]float64
	var intersectpt []float64
	tilemap := map[m.TileID][]*geojson.Feature{}

	// getting properties for later
	properties := line.Properties

	// iterating through each point
	ept := line.Geometry.LineString[0]
	oldpt = pc.Point{ept[0], ept[1]}
	oldtileid = m.Tile(oldpt.X, oldpt.Y, zoom)
	oldbds = m.Bounds(oldtileid)

	geoms := line.Geometry.LineString[1:]
	tilecoords = append(tilecoords, ept)

	for _, ept := range geoms {
		// getting pt,tileid and bounds
		pt := pc.Point{X: ept[0], Y: ept[1]}
		tileid = m.Tile(pt.X, pt.Y, zoom)
		bds = m.Bounds(tileid)

		// skipping first pt

		// shit goes down here
		// getting the distances between two coordinate points
		dist := distance_pts(oldpt, pt)

		// if the point delta we are straddling is between two tileids
		// i.e. has crossed one of planes
		if tileid != oldtileid {
			bnddist := distance_bounds(bds)

			// if one of the distances violates or is greater than the distance
			// for bounds it will be sent into a tile creation function
			if (bnddist.deltaX < dist.deltaX) || (bnddist.deltaY < dist.deltaY) {
				// send to tile generation function
				// an edge case I don't cover yet
				//tilemap = Add_Tilemap_Segment(oldpt,pt,zoom,tilemap)
			} else {
				// otherwise handle normally finding the intersection point and adding in the
				// the end of tile coords
				axis = which_plane(oldpt, pt, oldbds)
				intersectpt = itersection_pt(oldpt, pt, oldbds, axis)
				tilecoords = append(tilecoords, []float64{oldpt.X, oldpt.Y})

				tilecoords = append(tilecoords, intersectpt)
				oldtileid = Get_Centroid_TileID(tilecoords,zoom)
				// creating new geometry
				newgeom := geojson.Geometry{Type: "LineString"}
				newgeom.LineString = tilecoords
				tilemap[oldtileid] = append(tilemap[oldtileid], &geojson.Feature{Geometry: &newgeom, Properties: properties})

				// setting tile coords back to only the intersection point
				tilecoords = [][]float64{intersectpt}

			}
		} else {
			tilecoords = append(tilecoords, []float64{oldpt.X, oldpt.Y})

		}

		// stateful stuff
		oldpt = pt
		oldtileid = tileid
		oldbds = bds
	}

	// adding the last point
	tilecoords = append(tilecoords, []float64{oldpt.X, oldpt.Y})

	// adding the last feature
	newe := geojson.Geometry{Type: "LineString"}
	newe.LineString = tilecoords
	oldtileid = Get_Centroid_TileID(tilecoords,zoom)

	tilemap[oldtileid] = append(tilemap[oldtileid], &geojson.Feature{Geometry: &newe, Properties: properties})

	return tilemap
}

// translations. that probably aren't needed
func translate(val string) string {
	if "upper" == val {
		return "north"
	} else if "lower" == val {
		return "south"
	} else if "left" == val {
		return "west"
	} else if "right" == val {
		return "east"
	}
	return val
}

// translations. that probably aren't needed
func opp_translate(val string) string {
	if "north" == val {
		return "upper"
	} else if "south" == val {
		return "lower"
	} else if "west" == val {
		return "left"
	} else if "right" == val {
		return "east"
	}
	return val
}

func Get_string(align []pc.Point) string {
	newlist := []string{}
	for _, i := range align {
		newlist = append(newlist, fmt.Sprintf("[%f,%f]", i.X, i.Y))
	}
	return fmt.Sprintf("[%s]", strings.Join(newlist, ","))
}

// lints the children of a partmap
func Lint_Children_Lines(tilemap map[m.TileID][]*geojson.Feature, k m.TileID) map[m.TileID][]*geojson.Feature {
	childtiles := m.Children(k)
	newtilemap := map[m.TileID][]*geojson.Feature{}

	// iterating through each child
	for _, child := range childtiles {
		newtilemap[child] = tilemap[child]
	}

	return newtilemap
}
