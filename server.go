package gotile

import (
	"fmt"
	"strings"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"encoding/json"
	"os/exec"
	"io/ioutil"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"log"
  "net/http"
  "time"
  m "github.com/murphy214/mercantile"
  "strconv"

)


// server layer
type Server struct {
	Mbtiles []string
	Geobufs []Geobuf_Serve
}

// getting the vector layernames in a mbtiles filestring
func Get_Vector_Layers(filename string) []string {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		fmt.Println(err)
	}
	//defer db.Close()

	jsonstring := Get_Json_String(db)
	var vector_layers Vector_Layers
	_ = json.Unmarshal([]byte(jsonstring),&vector_layers)
	layernames := []string{}
	for _,i := range vector_layers.Vector_Layers {
		layernames = append(layernames,i.ID)
	}

	return layernames
}

func Get_Part_Layer(layername string) string {

 	return strings.Replace(`var layerColor = '#' + randomColor(lightColors);

  map.addLayer({
    'id': 'delaware_polygons-polygons',
    'type': 'fill',
    'source': 'delaware.mbtiles',
    'source-layer': 'delaware_polygons',
    'filter': ["==", "$type", "Polygon"],
    'layout': {},
    'paint': {
      'fill-opacity': 0.1,
      'fill-color': layerColor
    }
  });

  map.addLayer({
    'id': 'delaware_polygons-polygons-outline',
    'type': 'line',
    'source': 'delaware.mbtiles',
    'source-layer': 'delaware_polygons',
    'filter': ["==", "$type", "Polygon"],
    'layout': {
      'line-join': 'round',
      'line-cap': 'round'
    },
    'paint': {
      'line-color': layerColor,
      'line-width': 1,
      'line-opacity': 0.75
    }
  });

  map.addLayer({
    'id': 'delaware_polygons-lines',
    'type': 'line',
    'source': 'delaware.mbtiles',
    'source-layer': 'delaware_polygons',
    'filter': ["==", "$type", "LineString"],
    'layout': {
      'line-join': 'round',
      'line-cap': 'round'
    },
    'paint': {
      'line-color': layerColor,
      'line-width': 1,
      'line-opacity': 0.75
    }
  });

  map.addLayer({
    'id': 'delaware_polygons-pts',
    'type': 'circle',
    'source': 'delaware.mbtiles',
    'source-layer': 'delaware_polygons',
    'filter': ["==", "$type", "Point"],
    'paint': {
      'circle-color': layerColor,
      'circle-radius': 2.5,
      'circle-opacity': 0.75
    }
  });

  layers.polygons.push('delaware_polygons-polygons');
  layers.polygons.push('delaware_polygons-polygons-outline');
  layers.lines.push('delaware_polygons-lines');
  layers.pts.push('delaware_polygons-pts');
  `,"delaware_polygons",layername,-1)
}

func Start_End() (string,string) {
	return `
<html><head>
  <meta charset="utf-8">
  <title>mbview - vector</title>
  <meta name="viewport" content="initial-scale=1,maximum-scale=1,user-scalable=no">
  <script src="https://api.tiles.mapbox.com/mapbox-gl-js/v0.34.0/mapbox-gl.js"></script>
  <link href="https://api.tiles.mapbox.com/mapbox-gl-js/v0.34.0/mapbox-gl.css" rel="stylesheet">
  <link href="https://www.mapbox.com/base/latest/base.css" rel="stylesheet">
  <style>
    body { margin:0; padding:0; }
    #map { position:absolute; top:0; bottom:0; width:100%; }
    .mbview_popup {
      color: #333;
      display: table;
      font-family: "Open Sans", sans-serif;
      font-size: 10px;
    }

    .mbview_feature:not(:last-child) {
      border-bottom: 1px solid #ccc;
    }

    .mbview_layer:before {
      content: '#';
    }

    .mbview_layer {
      display: block;
      font-weight: bold;
    }

    .mbview_property {
      display: table-row;
    }

    .mbview_property-value {
      display: table-cell;

    }

    .mbview_property-name {
      display: table-cell;
      padding-right: 10px;
    }
  </style>
</head>
<body>

<style>
#menu {
  position: absolute;
  top:10px;
  right:10px;
  z-index: 1;
  color: white;
  cursor: pointer;
}
#menu-container {
  position: absolute;
  display: none;
  top: 50px;
  right: 10px;
  z-index: 1;
  background-color: white;
  padding: 20px;
}
</style>

<div id="menu"><span class="icon menu big"></span></div>

<div id="menu-container">
  <h4>Filter</h4>
  <div id="menu-filter" onchange="menuFilter()" class="rounded-toggle short inline">
    <input id="filter-all" type="radio" name="rtoggle" value="all" checked="checked">
    <label for="filter-all">all</label>
    <input id="filter-polygons" type="radio" name="rtoggle" value="polygons">
    <label for="filter-polygons">polygons</label>
    <input id="filter-lines" type="radio" name="rtoggle" value="lines">
    <label for="filter-lines">lines</label>
    <input id="filter-pts" type="radio" name="rtoggle" value="pts">
    <label for="filter-pts">points</label>
  </div>
  <h4>Popup</h4>
  <div id="menu-popup" onchange="menuPopup()" class="rounded-toggle short inline">
    <input id="show-popup" type="checkbox" name="ptoggle" value="all" '="">
    <label for="show-popup">show attributes</label>
  </div>
</div>

<script>

// Show and hide hamburger menu as needed
var menuBtn = document.querySelector("#menu");
var menu = document.querySelector("#menu-container");
menuBtn.addEventListener('click', function() {
  popup.remove();
  if (menuBtn.className.indexOf('active') > -1) {
    //Hide Menu
    menuBtn.className = '';
    menu.style.display = 'none';
  } else {
    //Show Menu
    menuBtn.className = 'active';
    menu.style.display = 'block';

  }
}, false);

//Menu-Filter Module
function menuFilter() {
  if (document.querySelector("#filter-all").checked) {
    paint(layers.pts, 'visible');
    paint(layers.lines, 'visible');
    paint(layers.polygons, 'visible');
  } else if (document.querySelector("#filter-pts").checked) {
    paint(layers.pts, 'visible');
    paint(layers.lines, 'none');
    paint(layers.polygons, 'none');
  } else if (document.querySelector("#filter-lines").checked) {
    paint(layers.pts, 'none');
    paint(layers.lines, 'visible');
    paint(layers.polygons, 'none');
  } else if (document.querySelector("#filter-polygons").checked) {
    paint(layers.pts, 'none');
    paint(layers.lines, 'none');
    paint(layers.polygons, 'visible');
  }

  function paint(layers, val) {
    layers.forEach(function(layer) {
      map.setLayoutProperty(layer, 'visibility', val)
    });
  }
}

function menuPopup() {
  wantPopup = document.querySelector("#show-popup").checked;
}

</script>


<div id="map" class="mapboxgl-map"><div class="mapboxgl-canvas-container mapboxgl-interactive"><canvas class="mapboxgl-canvas" tabindex="0" aria-label="Map" style="position: absolute; width: 263px; height: 775px;" width="526" height="1550"></canvas></div><div class="mapboxgl-control-container"><div class="mapboxgl-ctrl-top-left"></div><div class="mapboxgl-ctrl-top-right"></div><div class="mapboxgl-ctrl-bottom-left"><div class="mapboxgl-ctrl"><a class="mapboxgl-ctrl-logo" target="_blank" href="https://www.mapbox.com/" aria-label="Mapbox logo"></a></div></div><div class="mapboxgl-ctrl-bottom-right"><div class="mapboxgl-ctrl mapboxgl-ctrl-attrib compact"><a href="https://www.mapbox.com/about/maps/" target="_blank">© Mapbox</a> <a href="http://www.openstreetmap.org/about/" target="_blank">© OpenStreetMap</a> <a class="mapbox-improve-map" href="https://www.mapbox.com/map-feedback/#/0/0/13" target="_blank">Improve this map</a></div></div></div></div>

<script>

var center = [0,0,6];
center = [center[0], center[1]];

mapboxgl.accessToken = 'pk.eyJ1IjoicnNiYXVtYW5uIiwiYSI6IjdiOWEzZGIyMGNkOGY3NWQ4ZTBhN2Y5ZGU2Mzg2NDY2In0.jycgv7qwF8MMIWt4cT0RaQ';
var map = new mapboxgl.Map({
  container: 'map',
  style: 'mapbox://styles/mapbox/dark-v9',
  center: center,
  zoom: 12,
  hash: true,
  maxZoom: 30
});

var layers = {
  pts: [],
  lines: [],
  polygons: []
}

var lightColors = [
  'FC49A3', // pink
  'CC66FF', // purple-ish
  '66CCFF', // sky blue
  '66FFCC', // teal
  '00FF00', // lime green
  'FFCC66', // light orange
  'FF6666', // salmon
  'FF0000', // red
  'FF8000', // orange
  'FFFF66', // yellow
  '00FFFF'  // turquoise
];

function randomColor(colors) {
  var randomNumber = parseInt(Math.random() * colors.length);
  return colors[randomNumber];
}

map.on('load', function () {
  

    map.addSource('delaware.mbtiles', {
      type: 'vector',
      tiles: [
        'http://localhost:5000/{z}/{x}/{y}'
      ],
      maxzoom: 20
    });`,`


  
});


function displayValue(value) {
  if (typeof value === 'undefined' || value === null) return value;
  if (typeof value === 'object' ||
      typeof value === 'number' ||
      typeof value === 'string') return value.toString();
  return value;
}

function renderProperty(propertyName, property) {
  return '<div class="mbview_property">' +
    '<div class="mbview_property-name">' + propertyName + '</div>' +
    '<div class="mbview_property-value">' + displayValue(property) + '</div>' +
    '</div>';
}

function renderLayer(layerId) {
  return '<div class="mbview_layer">' + layerId + '</div>';
}

function renderProperties(feature) {
  var sourceProperty = renderLayer(feature.layer['source-layer'] || feature.layer.source);
  var typeProperty = renderProperty('$type', feature.geometry.type);
  var properties = Object.keys(feature.properties).map(function (propertyName) {
    return renderProperty(propertyName, feature.properties[propertyName]);
  });
  return [sourceProperty, typeProperty].concat(properties).join('');
}

function renderFeatures(features) {
  return features.map(function (ft) {
    return '<div class="mbview_feature">' + renderProperties(ft) + '</div>';
  }).join('');
}

function renderPopup(features) {
  return '<div class="mbview_popup">' + renderFeatures(features) + '</div>';
}

var popup = new mapboxgl.Popup({
  closeButton: false,
  closeOnClick: false
});

var wantPopup = false;

console.log('layers', layers);
map.on('mousemove', function (e) {
  // set a bbox around the pointer
  var selectThreshold = 3;
  var queryBox = [
    [
      e.point.x - selectThreshold,
      e.point.y + selectThreshold
    ], // bottom left (SW)
    [
      e.point.x + selectThreshold,
      e.point.y - selectThreshold
    ] // top right (NE)
  ];

  var features = map.queryRenderedFeatures(queryBox, {
    layers: layers.polygons.concat(layers.lines.concat(layers.pts))
  }) || [];
  map.getCanvas().style.cursor = (features.length) ? 'pointer' : '';

  if (!features.length || !wantPopup) {
    popup.remove();
  } else {
    popup.setLngLat(e.lngLat)
      .setHTML(renderPopup(features))
      .addTo(map);
  }
});
</script>



</body></html>`	

}

// creating and opening the html 
func (server Server) Create_Open_Html() {
	layernames := []string{}
	for _,i := range server.Mbtiles {
		layernames = append(layernames,Get_Vector_Layers(i)...)
	}
	for _,i := range server.Geobufs {
		layernames = append(layernames,i.Config_Dynamic.LayerName)
	}

	layer_parts := []string{}
	for _,layername := range layernames {
		layer_parts = append(layer_parts,Get_Part_Layer(layername))
	}
	middle := strings.Join(layer_parts,"\n")
	start,end := Start_End()

	total := start + "\n" + middle +  "\n" + end

	ioutil.WriteFile("index.html",[]byte(total),0677)

	exec.Command("open","index.html").Run()

}

type Mb_Struct struct {
	Stmt *sql.Stmt
	DB *sql.DB 
	Mutex *sync.Mutex
}


func (server Server) Serve() {
	newlist := []Mb_Struct{}
	for _,i := range server.Mbtiles {
		filename := i
		db, err := sql.Open("sqlite3", filename)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		
		stmt, err := tx.Prepare("select tile_data from tiles where zoom_level = ? and tile_column = ? and tile_row = ?;")
		if err != nil {
			log.Fatal(err)
		}
		var mm sync.Mutex
		defer stmt.Close()
		newlist = append(newlist,Mb_Struct{Stmt:stmt,DB:db,Mutex:&mm})
	}
  var mutex sync.Mutex

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vals := strings.Split(r.URL.Path,"/")
		if len(vals) == 4 {

			z := vals[1]
			x := vals[2]
			y := vals[3]

			znew,_ := strconv.ParseInt(z,10, 64)
			xnew,_ := strconv.ParseInt(x,10, 64)
			ynew,_ := strconv.ParseInt(y,10, 64)

			ynew_sql := (1 << uint64(znew)) - 1 - ynew
      s := time.Now()			
      w.Header().Set("Access-Control-Allow-Origin", "*")
			totalbytes := []byte{}
			
			// iterating through each mbtiles server
			for _,val := range newlist {
				var data []byte 
				val.Mutex.Lock()
				err := val.Stmt.QueryRow(int(znew),int(xnew),ynew_sql).Scan(&data)
				val.Mutex.Unlock()
				if err != nil {
					fmt.Print(err,"\n")
				}
				totalbytes = append(totalbytes,data...)
			}

			// iterating through each geobuf server 
			for _,buf := range server.Geobufs {
        mutex.Lock()
        totalbytes = append(totalbytes,buf.Make_Tile(m.TileID{xnew,ynew,uint64(znew)})...)
        mutex.Unlock()
			}



			fmt.Fprintf(w,"%s", string(totalbytes))
			fmt.Println(time.Now().Sub(s))
		}
	})

	err := http.ListenAndServe(":5000", h)
	log.Fatal(err)

}




