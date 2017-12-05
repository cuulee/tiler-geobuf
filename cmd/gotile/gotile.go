package main

import (
  "fmt"
  "os"
  t "github.com/murphy214/gotile-geobuf"
  "github.com/urfave/cli"
  "os/exec"
  "strings"
  g "github.com/murphy214/geobuf"
)

func main() {
  app := cli.NewApp()
  server := t.Server{}
  app.Action = func(c *cli.Context) error {

    for _,i := range c.Args() {
      if strings.HasSuffix(i,"mbtiles") {
        server.Mbtiles = append(server.Mbtiles)
      } else if strings.HasSuffix(i,"geobuf") {
        geobuf := g.Geobuf_File(i)
        config := t.Config_Dynamic{LayerName:i,Minzoom:0,Maxzoom:20}
        geobuf_server := t.New_Geobuf_Serve(geobuf,config)
        server.Geobufs = append(server.Geobufs,geobuf_server)
      } else if strings.HasSuffix(i,"geojson") {
        fmt.Println("geojson file given converting to geobuf")
        geobuf_filename := strings.Split(i,".")[0] + ".geobuf"
        exec.Command("geojson2geobuf",i,geobuf_filename).Run()
        geobuf := g.Geobuf_File(geobuf_filename)
        config := t.Config_Dynamic{LayerName:i,Minzoom:0,Maxzoom:20}
        geobuf_server := t.New_Geobuf_Serve(geobuf,config)
        server.Geobufs = append(server.Geobufs,geobuf_server)
      }
    }
    server.Create_Open_Html()
    server.Serve()
    //fmt.Println(c.Args())
    //infilename := c.Args().Get(0)
    //outfilename := c.Args().Get(1)
    //fmt.Println("Converting: ",infilename,"to geojson filename:", outfilename)

    //g.Write_Geojson_Out(infilename,outfilename)

    //fmt.Printf("Hello %q", c.Args().Get(0))
    return nil
  }

  app.Run(os.Args)
}