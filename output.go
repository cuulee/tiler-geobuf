package gotile

import (
	"io/ioutil"
	//"sync"
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	"encoding/json"
	"reflect"
	"github.com/paulmach/go.geojson"
	"os"
	"log"
)

// Configuration shit
type Config struct {
	Type string // json or mbtiles
	Minzoom int // minimum zoom
	Maxzoom int // maximum zoom
	Increment int
	Dir string
	Prefix string // prefix
	Zooms []int // zooms (not needed)
	Currentzoom int // current zoom (not needed)
	OutputFilename string // output mbtiles filename (not needed)
	Memory float64 // memory (not needed)
	New_Output bool // output whether to delete the old output or keep it
	Json_Meta string // json metadata
	FirstFeature *geojson.Feature
}

// vector tile struct 
type Vector_Tile struct {
	Filename string
	Data []byte
	Tileid m.TileID
}
// epands the configuration structure
func Expand_Config(config Config) Config {
	count := config.Minzoom
	zooms := []int{}
	for count <= config.Maxzoom {
		zooms = append(zooms,count)
		count += 1
	}

	config.Zooms = zooms
	config.OutputFilename = config.Prefix + ".mbtiles"
	return config
}

var jsondata = `{
        "vector_layers": [
            {
                "id": "county",
                "description": "",
                "minzoom": 5,
                "maxzoom": 13,
                "fields": {
                    "area": "String",
                    "colorkey":"String"
                }
            }]
    }`


// vector layer json
type Vector_Layer struct {
	ID string `json:"id"`
	Description string `json:"description"`
	Minzoom int `json:"minzoom"`
	Maxzoom int `json:"maxzoom"`
	Fields map[string]string `json:"fields"`
}


type Vector_Layers struct {
	Vector_Layers []Vector_Layer `json:"vector_layers"`
}

// returns the string of the json meta data
func Make_Json_Meta(config Config,feat *geojson.Feature) string {
	// v vector_layers 
	var vector_layers Vector_Layers
	_ = json.Unmarshal([]byte(config.Json_Meta),&vector_layers)

	layer := Vector_Layer{ID:config.Prefix,Description:"",Minzoom:config.Minzoom,Maxzoom:config.Maxzoom}

	fields := Reflect_Fields(feat.Properties)
	layer.Fields = fields

	vector_layers.Vector_Layers = append(vector_layers.Vector_Layers,layer)

	b,_ := json.Marshal(vector_layers)
	return string(b)
}

// creating the slice that will be used to create the metadata table
func Make_Metadata_Slice(config Config,feat *geojson.Feature) [][]string {
	// getting the json blob metadata
	jsondata := Make_Json_Meta(config,feat)

	// creating values 
	values := [][]string{{"name",config.OutputFilename},{"type","overlay"},{"version","2"},{"description",config.OutputFilename},{"format","pbf"},{"json",jsondata}}

	return values
}

func Create_Metadata(db *sql.DB,config Config) (*sql.Stmt,*sql.Tx) {

	sqlStmt := `
	CREATE TABLE metadata (name text, value text);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
	fmt.Printf("Created metadata table: %s.\n",config.OutputFilename)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into metadata(value, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	return stmt,tx 
}

// preparing update statement for db thing
func Update_Metadata(db *sql.DB) (*sql.Stmt,*sql.Tx) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := db.Prepare("update metadata set value=? where name=?")
	if err != nil {
		log.Fatal(err)
	}

	return stmt,tx 
}

// gets the json db
func Get_Json_String(db *sql.DB) string {
	sqlStmt := `
	select value from metadata where name = "json";
	`
	var jsonstring string
	err := db.QueryRow(sqlStmt).Scan(&jsonstring)

	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
	return jsonstring
}


// creates the sqllite database and inserts metadata 
func Create_Database_Meta(config Config,feat *geojson.Feature) *sql.DB {
	if config.New_Output == true {
		os.Remove(config.OutputFilename)
	}

	db, err := sql.Open("sqlite3", config.OutputFilename)
	if err != nil {
		fmt.Println(err)
	}
	//defer db.Close()
	var stmt *sql.Stmt
	var tx *sql.Tx
	if config.New_Output == true {
		fmt.Printf("Creating and opening %s.\n",config.OutputFilename)
		stmt,tx = Create_Metadata(db,config)
		config.Json_Meta = `{"vector_layers": []}`
	} else if config.New_Output == false {
		stmt,tx = Update_Metadata(db)
		config.Json_Meta = Get_Json_String(db)
	}


	// creating metadata slice string
	values := Make_Metadata_Slice(config,feat)


	defer stmt.Close()
	for _,i := range values {
		_, err = stmt.Exec(i[1],i[0])
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()

	if config.New_Output == true {
		sqlStmt := `
		CREATE TABLE tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob);
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
		}

		tx, err = db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Inserted appropriate metadata: %s.\n",config.OutputFilename)

	}
	return db 
}



// reflects a tile value back and stuff
func Reflect_Fields(mymap map[string]interface{}) map[string]string {
	newmap := map[string]string{}
	for k,v := range mymap {

		vv := reflect.ValueOf(v)
		kd := vv.Kind()
		if (reflect.Float64 == kd) || (reflect.Float32 == kd) {
			//fmt.Print(v, "float", k)
			newmap[k] = "Number"
			//hash = Hash_Tv(tv)
		} else if (reflect.Int == kd) || (reflect.Int8 == kd) || (reflect.Int16 == kd) || (reflect.Int32 == kd) || (reflect.Int64 == kd) || (reflect.Uint8 == kd) || (reflect.Uint16 == kd) || (reflect.Uint32 == kd) || (reflect.Uint64 == kd) {
			//fmt.Print(v, "int", k)
			newmap[k] = "Number"
			//hash = Hash_Tv(tv)
		} else if reflect.String == kd {
			//fmt.Print(v, "str", k)
			newmap[k] = "String"
			//hash = Hash_Tv(tv)

		} else {
			fmt.Print(k,v,"\n")
		}
	}
	return newmap
}




// inserting data into shit
func Insert_Data2(newmap map[m.TileID]Vector_Tile,db *sql.DB,config Config) *sql.DB {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	
	stmt, err := tx.Prepare("insert into tiles(zoom_level, tile_column,tile_row,tile_data) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	count := 0
	total := 0
	count3 := 0


	sizenewmap := len(newmap)

	for k,v := range newmap {
		var data []byte

		k.Y = (1 << uint64(k.Z)) - 1 - k.Y 

		if config.New_Output == false {
			query := fmt.Sprintf("select tile_data from tiles where zoom_level = %d and tile_column = %d and tile_row = %d",k.Z,k.X,k.Y)
			err = tx.QueryRow(query).Scan(&data)
			if len(data) > 0 {
				v.Data = append(v.Data,data...)			
				_,err = tx.Exec(`update tiles set tile_data = ? where zoom_level = ? and tile_column = ? and tile_row = ?`,v.Data,k.Z,k.X,k.Y)
				if err != nil {
					fmt.Print(err,"\n")		
				}
			}
		} else {
			_, err = stmt.Exec(int(k.Z),int(k.X),int(k.Y),v.Data)
			if err != nil {
				fmt.Print(err,"\n")
			}
		}

		count += 1
		if count == 1000 {
			count = 0
			total += 1000
			fmt.Printf("\r[%d/%d] Compressing tiles and inserting into db.",total,sizenewmap)
		}

		count3 += 1
		//fmt.Print(count,"\n")
		//count += 1
	}



	tx.Commit()


	return db
	
}





// inserting data into shit
func Insert_Data3(newmap []Vector_Tile,db *sql.DB,config Config) *sql.DB {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
	}
	
	stmt, err := tx.Prepare("insert into tiles(zoom_level, tile_column,tile_row,tile_data) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	total := 0
	count3 := 0


	sizenewmap := len(newmap)

	for _,v := range newmap {
		k := v.Tileid
		var data []byte

		k.Y = (1 << uint64(k.Z)) - 1 - k.Y 
		if config.New_Output == false {
			query := fmt.Sprintf("select tile_data from tiles where zoom_level = %d and tile_column = %d and tile_row = %d",k.Z,k.X,k.Y)
			err = tx.QueryRow(query).Scan(&data)
			if len(data) > 0 {
				v.Data = append(v.Data,data...)			
				_,err = tx.Exec(`update tiles set tile_data = ? where zoom_level = ? and tile_column = ? and tile_row = ?`,v.Data,k.Z,k.X,k.Y)
				if err != nil {
					fmt.Print(err,"\n")		
				}
			}
		} else {
			_, err = stmt.Exec(int(k.Z),int(k.X),int(k.Y),v.Data)
			if err != nil {
				fmt.Print(err,"\n")
			}
		}	

		count3 += 1
		if count3 == 10 {
			count3 = 0
			total += 10
			fmt.Printf("\r[%d/%d] Compressing tiles and inserting into db.",total,sizenewmap)
		}

		//fmt.Print(count,"\n")
		//count += 1
	}



	tx.Commit()


	return db
	
}


func Make_Index(db *sql.DB) {
	defer db.Close()


	sqlStmt := `
	CREATE UNIQUE INDEX IF NOT EXISTS tile_index on tiles (zoom_level, tile_column, tile_row)
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		//db.Close()

		fmt.Print(err,"\n")

		//_, err = db.Exec(sqlStmt)
		//return
	}

}

// writes a json file
func Write_Json(totalmap map[m.TileID]Vector_Tile,jsonfilename string) {
	newmap := map[string][]byte{}
	for _,i := range totalmap {
		newmap[i.Filename] = i.Data
	}

	b,_ := json.Marshal(newmap)
	ioutil.WriteFile(jsonfilename,b,0666)

}
