package gotile

import (
	m "github.com/murphy214/mercantile"
	"github.com/paulmach/go.geojson"
	//"fmt"
)

// recursively drills until the max zoom is reached
func Make_Zoom_Drill(k m.TileID, v []*geojson.Feature, prefix string, endsize int) []Vector_Tile {
	outputsize := int(k.Z) + 1
	cc := make(chan map[m.TileID][]*geojson.Feature)
	for _, i := range v {
		go func(k m.TileID, i *geojson.Feature, cc chan map[m.TileID][]*geojson.Feature) {
			if i.Geometry.Type == "Polygon" {
				partmap := Children_Polygon(i, k) 
				cc <- partmap
			} else if i.Geometry.Type == "LineString" {
				partmap := Env_Line(i, int(k.Z+1))
				partmap = Lint_Children_Lines(partmap, k)
				cc <- partmap
			} else if i.Geometry.Type == "Point" {
				partmap := map[m.TileID][]*geojson.Feature{}
				pt := i.Geometry.Point
				tileid := m.Tile(pt[0], pt[1], int(k.Z+1))
				partmap[tileid] = append(partmap[tileid], i)
				cc <- partmap
			}
		}(k, i, cc)
	}

	// collecting all into child map
	childmap := map[m.TileID][]*geojson.Feature{}
	for range v {
		partmap := <-cc
		for kk, vv := range partmap {
			childmap[kk] = append(childmap[kk], vv...)
		}
	}

	// iterating through each value in the child map and waiting to complete
	//var wg sync.WaitGroup
	vtchan := make(chan Vector_Tile)
	for kkk, vvv := range childmap {
		//childmap = map[m.TileID][]*geojson.Feature{}
		//wg.Add(1)
		go func(kkk m.TileID, vvv []*geojson.Feature, prefix string,vtchan chan Vector_Tile) {
			vtchan <- Make_Tile_Geojson(kkk, vvv, prefix)
				//Make_Zoom_Drill(kkk, vvv, prefix, endsize)
			//wg.Done()

		}(kkk, vvv, prefix,vtchan)
	}
	
	vector_tiles := []Vector_Tile{}
	for range childmap {
		vt := <-vtchan
		vector_tiles = append(vector_tiles,vt)
	}

	//wg.Wait()
	if endsize != outputsize {
		ccc := make(chan []Vector_Tile)
		for kkk, vvv := range childmap {
			go func(kkk m.TileID, vvv []*geojson.Feature, prefix string,ccc chan []Vector_Tile) {
				ccc <- Make_Zoom_Drill(kkk,vvv,prefix,endsize)
			}(kkk,vvv,prefix,ccc)
		}
		// appending to the major vector tiles shit
		for range childmap {
			vts := <- ccc
			vector_tiles = append(vector_tiles,vts...)
		}
		return vector_tiles

	} else {


		return vector_tiles
	}
}