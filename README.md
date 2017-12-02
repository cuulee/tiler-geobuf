# gotile-geobuf

I've been sort of breaking up my vector tile generation repos in the hopes that hopefully it will show some progression over time, and this change is quite signicant compared to the old gotile. 

The old gotile worked sort of like an trial and error methodology to getting the two different areas of the algorithm correct, being the portion where you map all the independent features to a tileid all at once, then at specific zoom level you send in each tile and creates vector tiles recursively from their. This was alright, but the features were a) still in memory and b) spatial data is super dynamic. 

The solution, a fast way to sequentially write and read features into a file context dynamically. Geobuf is that it is essentially a protocol buffer geojson that can be written to by appending a byte array to the end of a file and read sequentially using a var int manipulation. Therefore we can effectively read and write from a file context only bringing in one feature per iteration at a time. 

From here we can abstract from the previous mapping data structure map[m.TileID][]*geojson.Feature to a data structure like this map[m.TileID]*g.Geobuf{} and build out tight data structure for each tile to the point where eventually we will be able to send off an independent tile recursively independent of zoom which is going to be awesome. 

# Why?

While gotile is super fast to medium to smaller datasets, the annoying part has always been larger datasets at like the gb level. This file context interface rather than any data in memory for long, will alone make the algorithm much better for larger data sets. While minimal affects really on smaller datasets. 

Also as explained above, whenever we can assure all tiles below will only occupy a certain amount of memory max, bam we can send that in recursively for super fast vt creation, so basically this creates a more unified top layer structure to build out on top of. 

# Problems with implementation

Currently my issue has been the limit on the amount of file contents I'm opening. I'm pretty liberal with file creation (currently brute forces all parents I know, terrible) and opening I may have to abstract opening the file only when being written to avoid this. 

# Cleaner code 

Another reason for this is algorithm has a much more direct, less dynamic approach in which you could build things and test corner cases much much easier. While the code isn't much shorter it is a lot easier to understand (hopefully)
