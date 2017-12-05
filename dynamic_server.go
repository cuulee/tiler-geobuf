package gotile

import (
	"fmt"
	m "github.com/murphy214/mercantile"
	"github.com/paulmach/go.geojson"
	g "github.com/murphy214/geobuf"
	"sync"
)

func Get_Bds_Polygon(coords [][][]float64) (m.Extrema) {
	north := -1000.
	south := 1000.
	east := -1000.
	west := 1000.
	lat := 0.
	long := 0.

	// iterating through each outer ring
	for _, coord := range coords {
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
		}
	}

	return m.Extrema{S: south, W: west, N: north, E: east}

}

func Get_Bds_Line(coords [][]float64) (m.Extrema) {
	north := -1000.
	south := 1000.
	east := -1000.
	west := 1000.
	lat := 0.
	long := 0.

	// iterating through each outer ring
		// iterating through each point in a ring
	for _, i := range coords {
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
	}
	return m.Extrema{S: south, W: west, N: north, E: east}
}

func Get_Bds_Point(coords []float64) (m.Extrema) {
	return m.Extrema{S: coords[1], W: coords[0], N: coords[1], E: coords[0]}
}

func Get_Bds(geom *geojson.Geometry) m.Extrema {
	if geom.Type == "Point" {
		return Get_Bds_Point(geom.Point)
	} else if geom.Type == "LineString" {
		return Get_Bds_Line(geom.LineString)
	} else if geom.Type == "Polygon" {
		return Get_Bds_Polygon(geom.Polygon)
	}
	return m.Extrema{}
}


// structure for finding overlapping values
func Overlapping_1D(box1min float64,box1max float64,box2min float64,box2max float64) bool {
	if box1max >= box2min && box2max >= box1min {
		return true
	} else {
		return false
	}
	return false
}


// returns a boolval for whether or not the bb intersects
func Intersect(bdsref m.Extrema,bds m.Extrema) bool {
	if Overlapping_1D(bdsref.W,bdsref.E,bds.W,bds.E) && Overlapping_1D(bdsref.S,bdsref.N,bds.S,bds.N) {
		return true
	} else {
		return false
	}

	return false
}

func Get_Between(low int64,high int64) []int64 {
	current := low 
	total := []int64{current}
	for current < high {
		current += int64(1)
		total = append(total,current)
	}
	return total
}



// getting bounds extrema
func Get_Tiles_Size(bds m.Extrema, size int) []m.TileID {
	if ((bds.N == bds.S)&&(bds.E==bds.W)) {
		return []m.TileID{m.Tile(bds.E,bds.N,size)}
	}
	tileidne := m.Tile(bds.E,bds.N,size) // norrth east
	tileidsw := m.Tile(bds.W,bds.S,size) // south west
	yhigh,ylow := tileidsw.Y,tileidne.Y
	xlow,xhigh := tileidsw.X,tileidne.X

	betweenx := Get_Between(xlow,xhigh)
	betweeny := Get_Between(ylow,yhigh)
	//fmt.Println(betweeny)
	boxbds := m.Bounds(m.TileID{betweenx[0],betweeny[0],uint64(size)})
	totalids := []m.TileID{}



	if Size_Comp(boxbds,bds) == true {
		for _,x := range betweenx {
			for _,y := range betweeny {
			//fmt.Println(x,y)

				totalids = append(totalids,m.TileID{x,y,uint64(size)})
			}
		}
	}
	return totalids
}

func Get_Tiles(bds m.Extrema,minsize int,maxsize int) []m.TileID {
	current := minsize
	totalids := []m.TileID{}
	for current <= maxsize {
		totalids = append(totalids,Get_Tiles_Size(bds,current)...)
		current += 1
	}
	return totalids
}

type Tile_Struct struct {
	Pos [2]int
	TileIDs []m.TileID
}

func Make_Tiles(pos [2]int,geobuf *g.Geobuf,minsize int,maxsize int) Tile_Struct {
	bds := Get_Bds(geobuf.FeaturePos(pos).Geometry)
	return Tile_Struct{TileIDs:Get_Tiles(bds,minsize,maxsize),Pos:pos}
}


// creates a tilemap
func Make_Tilemap(geobuf *g.Geobuf,startzoom int,endzoom int) map[m.TileID][][2]int  {
	c := make(chan Tile_Struct)

	totalmap := map[m.TileID][][2]int{}

	total := 0
	counter := 0
	for geobuf.Next() {
		pos := geobuf.File.Feat_Pos
		total += 1
		counter += 1
		go func(pos [2]int,c chan Tile_Struct) {
			c <- Make_Tiles(pos,geobuf,startzoom,endzoom) 

		}(pos,c)
		if counter == 10000 {
			// iterating through each feature in teh channel
			count := 0
			for count < counter {
				output := <-c
				pos := output.Pos
				for _,id := range output.TileIDs {
					totalmap[id] = append(totalmap[id],pos)
				}
				count += 1
				fmt.Printf("\r%d Features mapped.",total)
			}
			counter = 0	
		}
	}

	// iterating through each feature in teh channel
	count := 0
	for count < counter {
		output := <-c
		pos := output.Pos
		for _,id := range output.TileIDs {
			totalmap[id] = append(totalmap[id],pos)
		}
		count += 1
		fmt.Printf("\r[%d/%d]",count,total)
	}	
	return totalmap
}


func Make_Geojson_Tile(tileid m.TileID) *geojson.Feature {
	bds := m.Bounds(tileid)
	polygon := [][][]float64{{{bds.E, bds.N}, {bds.W, bds.N}, {bds.W, bds.S}, {bds.E, bds.S}}}
	return &geojson.Feature{Geometry:&geojson.Geometry{Polygon:polygon,Type:"Polygon"}}
}


type Config_Dynamic struct {
	Minzoom int
	Maxzoom int
	LayerName string
}

// geobuf serve struct
type Geobuf_Serve struct {
	Tile_Map map[m.TileID][][2]int
	Geobuf *g.Geobuf
	Config_Dynamic
	Cache_Map map[m.TileID][]byte
	Mutex *sync.Mutex
}

// create geobuf serve struct
// this prevents any higher maxzoom being used
func New_Geobuf_Serve(geobuf *g.Geobuf,config Config_Dynamic) Geobuf_Serve {
	var maxzoom int
	if config.Maxzoom > 14 {
		maxzoom = 14
	} else {
		maxzoom = config.Maxzoom
	}
	eh := map[m.TileID][]byte{}
	var mm sync.Mutex
	return Geobuf_Serve{Mutex:&mm,Cache_Map:eh,Tile_Map:Make_Tilemap(geobuf,config.Minzoom,maxzoom),Config_Dynamic:config,Geobuf:geobuf}
}

type Return_Struct struct {
	BoolVal bool
	Feature *geojson.Feature
}

// serves a tile from a geobuf server
func (geobuf_serve Geobuf_Serve) Make_Tile(tileid m.TileID) []byte {
	geobuf_serve.Mutex.Lock()
	bytevals,boolval := geobuf_serve.Cache_Map[tileid] 
	geobuf_serve.Mutex.Unlock()
	
	if boolval == true {
		return bytevals
	}


	parent := m.Parent(tileid)
	var tilecloak m.TileID
	if int(parent.Z) > 14 {
		center := m.Center(tileid)
		tilecloak = m.Tile(center[0],center[1],14)
	} else {
		tilecloak = tileid
	}

	positions := geobuf_serve.Tile_Map[tilecloak]
	//fmt.Println(tileid)
	//fmt.Println(parent)
	c := make(chan Return_Struct)
	feats := []*geojson.Feature{}
	counter := 0
	for _,pos := range positions {
		counter += 1
		go func(pos [2]int,c chan Return_Struct) {
			mymap := Map_Feature(geobuf_serve.Geobuf.FeaturePos(pos),int(tileid.Z),parent)
			//fmt.Println(mymap,tileid)
			feat,inbool := mymap[tileid]
			if inbool == false || len(feat) == 0 {
				c <- Return_Struct{BoolVal:false,Feature:&geojson.Feature{}}
			} else {
				c <- Return_Struct{BoolVal:true,Feature:feat[0]}
			}
		}(pos,c)
		if counter == 1000 {
			count := 0
			for count < counter {
				out := <-c

				if out.BoolVal == true {
					feats = append(feats,out.Feature)
				}				
				count += 1
			}
			counter = 0
		}
	}

	count := 0
	for count < counter {
		out := <-c

		if out.BoolVal == true {
			feats = append(feats,out.Feature)
		}				
		count += 1
	}
	eh :=  Make_Tile_Geojson2(tileid,feats,geobuf_serve.Config_Dynamic.LayerName).Data
	geobuf_serve.Mutex.Lock()
	geobuf_serve.Cache_Map[tileid] = eh
	geobuf_serve.Mutex.Unlock()

	return eh
}	


