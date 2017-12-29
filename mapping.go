package gotile

import (
	"math"
	//"fmt"
	m "github.com/murphy214/mercantile"
	g "github.com/murphy214/geobuf"
)

// mapping for point
type Mapper_Point struct {
	Extrema m.Extrema
	Size int
	DeltaX float64 
	DeltaY float64
	PointMap map[[2]int]string
}

// made to map other tiles
type Mapper_Other struct {
	Extrema m.Extrema
	Size float64
	DeltaX float64
	DeltaY float64
}

// outermost mapper
type Mapper struct {
	Point Mapper_Point
	Other Mapper_Other
}

func Map_Val(valout float64,delta float64) int {
	return int(math.Floor(valout / delta))
}

func Get_Delta_Mapping(min float64, max float64,number int) float64 {
	return (max - min) / float64(number)
}


// getting new mapper
func New_Mapper_Point(tileid m.TileID,size int) Mapper_Point {
	bds := m.Bounds(tileid)
	deltay := Get_Delta_Mapping(bds.S,bds.N,size)
	deltax := Get_Delta_Mapping(bds.W,bds.E,size)
	return Mapper_Point{DeltaX:deltax,DeltaY:deltay,Extrema:bds,Size:size,PointMap:map[[2]int]string{}}
}

// maps a single point
func (mapper Mapper_Point) MapPoint(point []float64) [2]int {
	return [2]int{Map_Val(point[0]-mapper.Extrema.W,mapper.DeltaX),Map_Val(point[1]-mapper.Extrema.S,mapper.DeltaY)}
}

// maps first from a list of bounding boxs
func (mapper *Mapper_Point) MapPoints_First(bb *g.BoundingBox) bool {
	boolval2 := false
	val := mapper.MapPoint([]float64{bb.BB.E,bb.BB.N})
	_,boolval := mapper.PointMap[val]
	if boolval == false {
		mapper.PointMap[val] = ""
		boolval2 = true
	}
	return boolval2
}

// mapping for other stuff
func New_Mapper_Other(tileid m.TileID,size float64) Mapper_Other {
	bds := m.Bounds(tileid)
	if size > 1.0 {
		size = size / 100.0
	}
	deltax := (bds.E - bds.W) * size
	deltay := (bds.N - bds.S) * size
	return Mapper_Other{Size:size,Extrema:bds,DeltaX:deltax,DeltaY:deltay}
}

// mapping for just one bounding box
func (mapping Mapper_Other) Percent(bb *g.BoundingBox) bool {
	deltay := (bb.BB.N - bb.BB.S) 
	deltax := (bb.BB.E - bb.BB.W) 
	if (mapping.DeltaX <= deltax) || mapping.DeltaY <= deltay {
		return true
	}
	return false
}

// where size is the size of a dimension of the square for which points to be reduced by
// where percent is a percent of line and polygons a
func New_Mapper(tileid m.TileID,size int,percent float64) *Mapper {
	return &Mapper{Point:New_Mapper_Point(tileid,size),Other:New_Mapper_Other(tileid,percent)}
}

// filters a feature about a mapper
func (mapper *Mapper) Filter(bb *g.BoundingBox) bool {
	if bb.Type == "Point" {
		return mapper.Point.MapPoints_First(bb)
	} else {
		return mapper.Other.Percent(bb)
	}
	return false
}



