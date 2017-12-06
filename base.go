package gotile

import (
	"fmt"
	g "github.com/murphy214/geobuf"
	m "github.com/murphy214/mercantile"
	"os"
	"github.com/paulmach/go.geojson"
	"strings"
	"sync"
	"path/filepath"
	"time"
)


// File_Map that builds a data structure in dir
type File_Map struct {
	File_Map map[m.TileID]*g.Geobuf
	Zoom int
	Dir string
	SS sync.Mutex
	Increment int
	Config Config
}

type Geobuf_Output struct {
	TileID m.TileID
	Geobuf *g.Geobuf
}

func File_Name(tileid m.TileID,dir string) string {
	return dir + "/" + strings.Replace(m.Tilestr(tileid),"/","_",-1) + ".geobuf"
}

// creates a file geobuf
func Create_File_Geobuf(tileid m.TileID,dir string) *g.Geobuf {
	filename := File_Name(tileid,dir)
	_,err := os.Create(filename)
	if nil != err {
		fmt.Println(err)
	}
	geob := g.Geobuf_File(filename)
	geob.Filename = filename
	return geob
}

func Fix_Increment(sizes [][2]int,increment int) [][][2]int {
	vals := [][][2]int{}
	current := 0
	for current < len(sizes) {
		newcurrent := current + increment 
		if newcurrent > len(sizes) {
			newcurrent = len(sizes)
		}

		vals = append(vals,sizes[current:newcurrent])
		current = newcurrent
	}
	return vals
}

func Get_Geobuf_Paths(searchDir string)  map[int][]string {

    fileList := []string{}
   	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
        fileList = append(fileList, path)
        return nil
    })
   	mymap := map[int][]string{}
    for _, file := range fileList[1:] {
    	if file[len(file) - 7:] == ".geobuf" {
            tileid := m.Strtile(strings.Replace(strings.Split(file,".")[0][5:],"_","/",-1))
            mymap[int(tileid.Z)] = append(mymap[int(tileid.Z)],file)
    	}
    }
    return mymap
}

// maps an individual point
func Map_Point(feature_point *geojson.Feature,zoom int) map[m.TileID][]*geojson.Feature {
	pt := feature_point.Geometry.Point
	tileid := m.Tile(pt[0],pt[1],zoom)
	return map[m.TileID][]*geojson.Feature{tileid:[]*geojson.Feature{feature_point}}
}	

// maps a single point
func Map_Line(i *geojson.Feature,zoom int,k m.TileID)map[m.TileID][]*geojson.Feature {
	var partmap map[m.TileID][]*geojson.Feature
	if k.Z == 0 && k.Y == 0 && k.X == 0 {
		partmap = Env_Line(i, zoom)
	} else {
		partmap = Env_Line(i, int(k.Z+1))
		partmap = Lint_Children_Lines(partmap, k)
	}
	return partmap
}

// maps a single point
func Map_Polygon(i *geojson.Feature,zoom int,k m.TileID) map[m.TileID][]*geojson.Feature {
	var partmap map[m.TileID][]*geojson.Feature
	if k.Z == 0 && k.Y == 0 && k.X == 0 {
		partmap = Env_Polygon(i, zoom)
	} else {
		partmap = Children_Polygon(i, k)
	}
	return partmap
}

func Within_Child(childbds m.Extrema,featbds m.Extrema) bool {
	if (childbds.N >= featbds.N) && (childbds.E >= featbds.E) && 
	(childbds.S <= featbds.S) && (childbds.W <= featbds.W) {
		return true
	} else {
		return false 
	} 
	return false
}


// maps an individual feature
func Map_Feature(feat *geojson.Feature,zoom int,k m.TileID) map[m.TileID][]*geojson.Feature {
	featbds := Get_Bds(feat.Geometry)
	// checking for simple within
	if feat.Geometry.Type != "Point" {
		for _,child := range m.Children(k) {
			childbds := m.Bounds(child)
			if Within_Child(childbds,featbds) == true {
				return map[m.TileID][]*geojson.Feature{child:[]*geojson.Feature{feat}}
			}
		}
	}

	if feat.Geometry.Type == "Point" {
		return Map_Point(feat,zoom)
	} else if feat.Geometry.Type == "LineString" {
		return Map_Line(feat,zoom,k)
	} else if feat.Geometry.Type == "Polygon" {
		return Map_Polygon(feat,zoom,k)
	}
	return map[m.TileID][]*geojson.Feature{}
}

func Get_Delta(bds m.Extrema) (float64,float64) {
	return (bds.E - bds.W),(bds.N - bds.S)
}
 

func Size_Comp(bds m.Extrema,featbds m.Extrema) bool {
	deltax,deltay := Get_Delta(bds)
	deltaxf,deltayf := Get_Delta(featbds)
	if ((deltax * .001) < deltaxf) || ((deltay * .001) < deltayf) {
		return true
	}
	return false
}
 



// maps an individual feature
func Map_Feature_Reduce(feat *geojson.Feature,zoom int,k m.TileID) map[m.TileID][]*geojson.Feature {
	featbds := Get_Bds(feat.Geometry)

	children := m.Children(k)
	bds := m.Bounds(children[0])

	//if AreaBds(bds) * .001 > AreaBds(featbds) {
	if Size_Comp(bds,featbds) == true {
		return map[m.TileID][]*geojson.Feature{} 
	}
	//}

	// checking for simple within
	if feat.Geometry.Type != "Point" {
		for _,child := range children {
			childbds := m.Bounds(child)
			if Within_Child(childbds,featbds) == true {
				return map[m.TileID][]*geojson.Feature{child:[]*geojson.Feature{feat}}
			}
		}
	}

	if feat.Geometry.Type == "Point" {
		return Map_Point(feat,zoom)
	} else if feat.Geometry.Type == "LineString" {
		return Map_Line(feat,zoom,k)
	} else if feat.Geometry.Type == "Polygon" {
		return Map_Polygon(feat,zoom,k)
	}
	return map[m.TileID][]*geojson.Feature{}
}



// removes the old filemap
func (filemap File_Map) Remove_Filemap() {
	for k := range filemap.File_Map {
		filename := File_Name(k,filemap.Dir)
		os.Remove(filename)
	}
}

// adding a channeled temporay map
func (filemap *File_Map) Add_Map_First(tilemap map[m.TileID][]*geojson.Feature) {
	for k,v := range tilemap {
		// getting the boolval
		filemap.SS.Lock()
		_,boolval := filemap.File_Map[k]
		filemap.SS.Unlock()
		if boolval == false {
			filemap.SS.Lock()
			filemap.File_Map[k] = Create_File_Geobuf(k,filemap.Dir)
			filemap.SS.Unlock()
		}

		// adding each feature to value
		for _,feat := range v {
			filemap.File_Map[k].Write_Feature(feat)
		}

	}
}

// adding a channeled temporay map
func (filemap *File_Map) Add_Map(tilemap map[m.TileID][]*geojson.Feature) {
	for k,v := range tilemap {
		// getting the boolval

		// adding each feature to value
		for _,feat := range v {
			filemap.File_Map[k].Write_Feature(feat)
		}

	}
}


// drills the map one farther down then previously before
func (filemap *File_Map) Drill_Map() *File_Map {
	newfilemap := &File_Map{Dir:filemap.Dir,Zoom:filemap.Zoom+1,File_Map:map[m.TileID]*g.Geobuf{},Increment:filemap.Increment,Config:filemap.Config}
	//newfilemap.Add_Files(filemap)
	// iterating through each file in the filemap
	increment := 4
	geobufs := []Geobuf_Output{}
	for k,v := range filemap.File_Map {
		geobufs = append(geobufs,Geobuf_Output{TileID:k,Geobuf:v})
		if len(geobufs) == increment {
			Make_Geobufs(geobufs,newfilemap)
			geobufs = []Geobuf_Output{}
		}
	}
	Make_Geobufs(geobufs,newfilemap)

	return newfilemap
}

// creates an initial File_Map
func (filemap *File_Map) Add_Geobuf(geobuf *g.Geobuf,k m.TileID) {
	newlist := [][2]int{}
	vals := Fix_Increment(geobuf.Sizes,filemap.Increment)
	totalcount := 0
	for _,newlist := range vals {
		Map_Bulk(newlist,geobuf,filemap,k,false)
		totalcount += len(newlist)
		//fmt.Printf("\r%d Values Mapped.        ",totalcount)

	}
	// adding the final left over newlist
	Map_Bulk(newlist,geobuf,filemap,k,false)
}




// concurrently creates a filemap on a set of geobufs
func Make_Geobufs(geobufs []Geobuf_Output,filemap *File_Map) {
	if len(geobufs) != 0 {
		var wg sync.WaitGroup
		for _,out := range geobufs {
			wg.Add(1)
			go func(out Geobuf_Output,filemap *File_Map) {
				filemap.Add_Geobuf(out.Geobuf,out.TileID)
				out.Geobuf.File.File.Close()
				os.Remove(out.Geobuf.Filename)
				wg.Done()

			}(out,filemap)
		}
		wg.Wait()
	}
}

// makes a bulk set of newlist inds
func Map_Bulk(newlist [][2]int,geobuf *g.Geobuf,filemap *File_Map,k m.TileID,boolval bool) {
	c := make(chan map[m.TileID][]*geojson.Feature) 
	for _,pos := range newlist {
		go func(pos [2]int,c chan map[m.TileID][]*geojson.Feature) {
			c <- Map_Feature(geobuf.FeaturePos(pos),filemap.Zoom,k)
		}(pos,c)
	}

	// adding each temorary map to the filemap
	for range newlist {
		if boolval == true {
			filemap.Add_Map_First(<-c)
		} else {
			filemap.Add_Map_First(<-c)

		}
	}

}

// adds a serives of files to the map
func (newfilemap *File_Map) Add_Files(oldfilemap *File_Map) {
	for k := range oldfilemap.File_Map {
		children := m.Children(k)
		for _,child := range children {
			newfilemap.File_Map[child] = Create_File_Geobuf(child,newfilemap.Dir)
		}
	}
	
	//fmt.Println(newfilemap)
}


// creates an initial File_Map
func Create_Map(geobuf *g.Geobuf,config Config) *File_Map {
	config.StartTime = time.Now()
	newlist := [][2]int{}
	config = Expand_Config(config)
	filemap := &File_Map{Dir:config.Dir,Zoom:config.Minzoom,File_Map:map[m.TileID]*g.Geobuf{},Increment:config.Increment,Config:config}
	totalcount := 0
	k := m.TileID{0,0,0}
	firstbool := false
	for geobuf.Next() {
		// adding config first feature to config
		if firstbool == false {
			filemap.Config.FirstFeature = geobuf.FeaturePos(geobuf.File.Feat_Pos)
			firstbool = true
		}

		newlist = append(newlist,geobuf.File.Feat_Pos)
		if len(newlist) == filemap.Increment {
			Map_Bulk(newlist,geobuf,filemap,k,true)
			newlist = [][2]int{}
			totalcount += filemap.Increment
			fmt.Printf("\r%d Values Mapped.        ",totalcount)
		}

	}
	// adding the final left over newlist
	Map_Bulk(newlist,geobuf,filemap,k,true)
	return filemap
}

