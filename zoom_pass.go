package gotile

import (
	"fmt"
	g "github.com/murphy214/geobuf"
	m "github.com/murphy214/mercantile"
	"database/sql"

)

func (filemap *File_Map) Zoom_Pass(db *sql.DB) *sql.DB {
	// iterating through each geobuf
	c := make(chan Vector_Tile)
	for k,v := range filemap.File_Map {
		go func(k m.TileID,v *g.Geobuf,c chan Vector_Tile) {
			c <- Make_Tile(k,v,"shit")
		}(k,v,c)
	}

	// collecting tiles
	newlist := []Vector_Tile{}
	count := 0
	for range filemap.File_Map {
		out := <-c
		if len(out.Data) > 0 {
			newlist = append(newlist,out)
		}
		count += 1
		fmt.Printf("\r[%d/%d] Tiles Made",count,len(filemap.File_Map))
	}
	
	Insert_Data3(newlist,db,filemap.Config)
	return db

}

// creating the tiles
func (filemap *File_Map) Make_Tiles() {
	db := Create_Database_Meta(filemap.Config,filemap.Config.FirstFeature)
	//fmt.Printf("%+v%s",db,"Imheredbdv")
	filemap.Config.Currentzoom = filemap.Zoom

	for filemap.Zoom <= filemap.Config.Maxzoom {
		filemap.Config.Currentzoom = filemap.Zoom
		db = filemap.Zoom_Pass(db)

		filemap = filemap.Drill_Map()
		//fmt.Printf("%+v,%+v\n",filemap.Config,filemap.Config.FirstFeature)
		//fmt.Println(vtlist)
		//fmt.Println(filemap.Zoom)
	}




	Make_Index(db)
}

