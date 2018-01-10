package gotile

import (
	//"sync"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	"github.com/paulmach/go.geojson"
	"time"
	"os"
	util "github.com/murphy214/mbtiles-util"
	"log"
)

// Configuration shit
type Config struct {
	Minzoom int // minimum zoom
	Maxzoom int // maximum zoom
	Increment int // how many features are processing concurrently in mapping
	Dir string // the temporary directory that will be used
	Prefix string // prefix
	Zooms []int // zooms (not needed)
	Currentzoom int // current zoom (not needed)
	OutputFilename string // output mbtiles filename (not needed)
	Memory float64 // memory (not needed)
	New_Output bool // output whether to delete the old output or keep it
	Json_Meta string // json metadata
	FirstFeature *geojson.Feature // first feature (intermediate value not user input)
	Drill_Zoom int // the zoom in which to drill down recursively 
	StartTime time.Time // (intermediate value)
	TotalTiles int // (intermediate value)
	PointMapping int // the dimmension of reduction for points  
	PercentMapping float64 // the percent of reduction for lines and polygons
	RDP bool
	Mbtiles util.Mbtiles
	Logger *Logger
}

type Logger struct {
	StartTime time.Time
	TotalTiles int
	TilesPerSec float64
	StartTimeZoom time.Time
	CountZoom int
	SizeZoom int
	TotalZoom int
}

// adding a logger which will be used within each tile function
func (logger *Logger) Add(tileid m.TileID) {
	logger.TotalTiles += 1
	logger.TotalZoom += 1
	if logger.TotalTiles%10000 == 0 {
		tiles_per_sec := int(float64(logger.TotalTiles) / time.Now().Sub(logger.StartTime).Seconds())
		string_time := time.Now().Round(time.Second).String()

		log.Printf("%s | tiles: %dk | elapsed: %d |tps: %d | Zoom: %d [%d/%d]\n",string_time,logger.TotalTiles / 1000,int(time.Now().Sub(logger.StartTime).Seconds()),tiles_per_sec,tileid.Z,logger.CountZoom,logger.SizeZoom)
	}
}




// vector tile struct 
type Vector_Tile struct {
	Filename string
	Data []byte
	Tileid m.TileID
}

func Make_Logger(startime time.Time) *Logger {
	return &Logger{StartTime:startime}
}


// epands the configuration structure
func Expand_Config(config Config) Config {
	count := config.Minzoom
	zooms := []int{}
	for count <= config.Maxzoom {
		zooms = append(zooms,count)
		count += 1
	}

	if config.PointMapping == 0 {
		config.PointMapping = 4096
	}
	if config.Dir == "" {
		config.Dir = "temp"
	}

	config.Zooms = zooms

	// creating util mbtiles config
	util_config := util.Config{LayerName:config.Prefix,
					FileName:config.OutputFilename,
					LayerProperties:config.FirstFeature.Properties,
					MinZoom:config.Minzoom,
					MaxZoom:config.Maxzoom,

				}

	if config.New_Output == true {
		os.Remove(config.OutputFilename)
		config.Mbtiles = util.Create_DB(util_config)
	} else {
		config.Mbtiles = util.Update_DB(util_config)
	}

	config.Logger = Make_Logger(config.StartTime)

	return config
}
