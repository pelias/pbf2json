
pbf2json creates a JSON stream of openstreetmap data from any PBF extract, you can pick-and-choose only the bits of the file you want, the library will take care of de-normalizing the relational data (nodes/ways) so you can put it straight in to your favourite document-store, inverted index or graph database.

![animated-gif](http://missinglink.embed.s3.amazonaws.com/pbf2json-2.gif)

### Run from pre-built binary

You don't need to have `Go` installed on your system to use one of the binary files in `./build`:

```bash
# 64-bit linux distributions
$ ./build/pbf2json.linux-x64
```

```bash
# 64-bit OSX distributions
$ ./build/pbf2json.darwin-x64
```

you can also run it on your raspberry pi!

```bash
# embedded devices
$ ./build/pbf2json.linux-arm
```

### Usage

To control which tags are output you must pass the `-tags=` flag to `pbf2json` and the PBF filepath:

```bash
$ ./build/pbf2json.linux-x64 -tags="amenity" /tmp/wellington_new-zealand.osm.pbf
```
```bash
{"id":170603342,"type":"node","lat":-41.289843000000005,"lon":174.7944402,"tags":{"amenity":"fountain","created_by":"Potlatch 0.5d","name":"Oriental Bay Fountain","source":"knowledge"},"timestamp":"0001-01-01T00:00:00Z"}
{"id":170605346,"type":"node","lat":-41.2861039,"lon":174.7711539,"tags":{"amenity":"fountain","created_by":"Potlatch 0.10c","source":"knowledge"},"timestamp":"0001-01-01T00:00:00Z"}
```

### Advanced Usage

Multiple tags can be specified with commas, records will be returned if they match one `OR` the other:

```bash
# all buildings and shops
-tags="building,shop"
```

Tags can also be grouped with the `+` symbol, records will only be returned if they match one `AND` the other:

```bash
# only records with BOTH housenumber and street specified
-tags="addr:housenumber+addr:street"
```

You can also combine the above 2 delimiters to get even more control over what get's returned:

```bash
# only highways and waterways which have a name
-tags="highway+name,waterway+name"
```

If you need to target only specific values for a tag you can specify exactly which values you wish to extract using the `~` symbol:

```bash
# only extract amenity tags which have the value of toilets or kindergarten
-tags="amenity~toilets,amenity~kindergarten"
```

### Denormalization

When processing the ways, the node refs are looked up for you and the lat/lon values are added to each way:

Since version `3.0` centroids are also computed for each way.

```bash
{
  "id": 301435061,
  "type": "way",
  "tags": {
    "addr:housenumber": "33",
    "addr:postcode": "N5 1TH",
    "addr:street": "Highbury Park",
    "building": "residential"
  },
  "centroid": {
    "lat": "51.554679",
    "lon": "-0.098485"
  },
  "nodes": [
    {
      "lat": "51.554663",
      "lon": "-0.098369"
    },
    {
      "lat": "51.554657",
      "lon": "-0.098529"
    },
    {
      "lat": "51.554656",
      "lon": "-0.098592"
    },
    {
      "lat": "51.554676",
      "lon": "-0.098590"
    },
    {
      "lat": "51.554680",
      "lon": "-0.098529"
    },
    {
      "lat": "51.554715",
      "lon": "-0.098529"
    },
    {
      "lat": "51.554720",
      "lon": "-0.098369"
    },
    {
      "lat": "51.554663",
      "lon": "-0.098369"
    }
  ],
  "timestamp": "0001-01-01T00:00:00Z"
}
```

### Leveldb

This library uses `leveldb` to store the lat/lon info about nodes so that it can denormalize the ways for you.

By default the leveldb path is set to `/tmp`, you can change where it stores the data with a flag:

```bash
$ ./build/pbf2json.linux-x64 -leveldb="/tmp/somewhere"
```

### Batched writes

Since version `3.0` writing of node info to leveldb is done in batches to improve performance.

By default the batch size is `50000`, you can change this with the following flag:

```bash
$ ./build/pbf2json.linux-x64 -batch="1000"
```

### NPM module

```javascript
var pbf2json = require('pbf2json'),
    through = require('through2');

var config = {
  file: '/tmp/wellington_new-zealand.osm.pbf',
  tags: 'addr:housenumber+addr:street',
  leveldb: '/tmp'
};

pbf2json.createReadStream( config )
 .pipe( through.obj( function( item, e, next ){
    console.log( item );
    next();
 }));
```

### Run the go code from source

Make sure `Go` is installed and configured on your system, see: https://gist.github.com/missinglink/4212a81a7d9c125b68d9

```bash
sudo apt-get install mercurial;
go get;
go run pbf2json.go;
```

### Compile source for all supported architecture

If you are doing a release and would like to compile for all supported architectures:

**note** if this is your first time doing this please read the notes in './compile.sh' to set it all up on your machine.

```bash
bash compile.sh;
```

### Compile source for a new architecture

If you would like to compile a version of this lib for an architecture which isn't currently supported you can:

```bash
go get;
go build pbf2json.go;
chmod +x pbf2json;
mv pbf2json build/pbf2json.{platform}-{arch};
```

Note you will need to change the variables {platform} and {arch} to match those returned by `nodejs` for your system:

```javascript
$ node
> var os=require('os')
> os.platform()
'linux'
> os.arch()
'x64'
```

Then submit a pull request, you are awesome ;)