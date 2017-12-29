# gotile-geobuf

Gotile is a project I've been working on in some form for a while, it uses an out of memory data structure to create vector tile sets in a mbtiles file and can be used either as a feature reducing data visualization tool OR a straight input-output vector tile set from a given geojson / geobuf file. 

The reason geobuf is used rather than the standard geojson file, is mainly for out of memory / sequential reading and writing as well as hooking things like bounding box values into each feature for easier mapping. The configuration structure is quite large and subject to change, however its also pretty featured, you can add layers to existing mbtiles files, and even create entire vector sets inputting a configuration and a geobuf in for each layer. I plan on eventually using this with toml or some other configuration centric file format. 

# Configuration Structure 



# Caveats 

Essentially my entire stack is written using only shape primitives (point, line, polygon) when a geojson file is converted into geobuf each multi geometry is compied as its own primitive with features copied over. This scorched earth approach may mean larger pbf sizes, but when its only one person sifting through a code base 3 geometry are 10x easier to manage than 6. These aren't necessarily hard to add, but add unneeded complexity for the stage this project is in. 

