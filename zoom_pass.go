package gotile

import (
	"fmt"
	g "github.com/murphy214/geobuf"
	m "github.com/murphy214/mercantile"
	//"database/sql"
	"math"
	"os"
	"io/ioutil"
	"sync"
)


// calculating memory implications
func Calc_Memory(raw_filesize int,total_features,currentzoom int,maxzoom int) int {
	gosize := int(math.Pow(float64(maxzoom - currentzoom),2.0))
	gosize = gosize * raw_filesize + gosize * 4000
	return gosize / 1000 // kb
}

func Number_Features(currentzoom int,maxzoom int,number_features int) int {
	//fmt.Println((maxzoom - currentzoom),"dif",currentzoom,(maxzoom - currentzoom) * number_features)
	total := 0
	current := currentzoom
	for current < maxzoom {
		number_features = number_features * 3
		total += number_features
		current += 1
	}
	return total
}

// gets the total number of features for each layer
func (filemap *File_Map) Total_Features() int {
	total := 0
	for k,v := range filemap.File_Map {
		total += Number_Features(int(k.Z),filemap.Config.Maxzoom,v.Total_Features)
	}
	return total
}


// gets the total number of features for each layer
func (filemap *File_Map) Total_Size() int {
	total := 0
	for k,v := range filemap.File_Map {
		total += Calc_Memory(v.File.FileSize,v.Total_Features,int(k.Z),filemap.Config.Maxzoom)
	}
	return total
}

// zoom pass
func (filemap *File_Map) Zoom_Pass() {
	// iterating through each geobuf
	//c := make(chan []Vector_Tile)
    maxGoroutines := 3
    guard := make(chan struct{}, maxGoroutines)
	
	//fmt.Println(filemap.Total_Features(),filemap.Config.Currentzoom)
	//fmt.Println(filemap.Total_Size())
	boolval := false
	if filemap.Config.Drill_Zoom == filemap.Config.Currentzoom {
		boolval = true
	}
	//boolval := false
	//size := len(filemap.File_Map)
	//totalmem := 0
	var wg sync.WaitGroup
	filemap.Config.Logger.SizeZoom = len(filemap.File_Map)
	filemap.Config.Logger.CountZoom = 0 
	for k,v := range filemap.File_Map {
		//sizemem := Calc_Memory(v.File.FileSize,v.Total_Features,int(k.Z),filemap.Config.Maxzoom)
		//memorysize := Calc_Memory(v.File.FileSize,int(k.Z),filemap.Config.Maxzoom)
        guard <- struct{}{} // would block if guard channel is already filled
		wg.Add(1)
		go func(k m.TileID,v *g.Geobuf) {
			filemap.Zoom = int(k.Z)
			if boolval == false {
			//fmt.Println(Number_Features(int(k.Z),filemap.Config.Maxzoom,v.Total_Features))
		       	<-guard

				Make_Tile(k, // tileid
					v, // geobuf
					filemap.Config.Prefix, // prefix
					filemap.Config.PointMapping, // pointmapping integer
					filemap.Config.PercentMapping, // percent mapping integer
					filemap.Config.RDP, // rdpbool
					filemap.Config.Mbtiles,
					filemap.Config.Logger,
				)
				wg.Done()
			} else {

				//fmt.Println(sizemem)
				//mutex.Lock()
				//fmt.Println(memorysize,k)
				bytevals,err := ioutil.ReadFile(v.Filename)
				if nil != err {
					fmt.Println(err)
				}
				Make_Tile(k, // tileid
					v, // geobuf
					filemap.Config.Prefix, // prefix
					filemap.Config.PointMapping, // pointmapping integer
					filemap.Config.PercentMapping, // percent mapping integer
					filemap.Config.RDP, // rdpbool
					filemap.Config.Mbtiles,
					filemap.Config.Logger,
				)

				v.File.File.Close()



				os.Remove(v.Filename)
				if filemap.Config.Drill_Zoom != filemap.Config.Maxzoom {
					Make_Zoom_Drill(k,g.Read_FeatureCollection(bytevals),filemap.Config.Prefix,filemap.Config.Maxzoom,filemap.Config.Mbtiles,filemap.Config.Logger)
				}

				
				//totalmem += sizemem

				filemap.SS.Lock()
				delete(filemap.File_Map,k)
				filemap.SS.Unlock()

		       	<-guard

				//mutex.Unlock()
				wg.Done()

			}
			filemap.Config.Logger.CountZoom += 1
		}(k,v)


	}
	wg.Wait()
	//filemap.Config = Insert_Data3(newlist,db,filemap.Config)

}

// creating the tiles
func (filemap *File_Map) Make_Tiles() {
	//fmt.Printf("%+v%s",db,"Imheredbdv")
	filemap.Config.Currentzoom = filemap.Zoom
	filemap.Zoom_Pass()	
	if filemap.Zoom == filemap.Config.Maxzoom {

	
	} else {
		for filemap.Zoom <= filemap.Config.Maxzoom {
			if len(filemap.File_Map) == 0 {
				//fmt.Println("shit")
				filemap.Zoom = 10000
			} else {
				//fmt.Println("here1")
				filemap = filemap.Drill_Map()
				filemap.Config.Currentzoom = filemap.Zoom

				filemap.Zoom_Pass()	
			}



			//fmt.Printf("%+v,%+v\n",filemap.Config,filemap.Config.FirstFeature)
			//fmt.Println(vtlist)
			//fmt.Println(filemap.Zoom)
		}
	}

	filemap.Config.Mbtiles.Commit()

	os.RemoveAll(filemap.Config.Dir)
	
}

