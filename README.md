
### Run from pre-built binary

You don't need to have `Go` installed on your system to use the binary file in `./build`:

```bash
$ ./build/pbf2json 
Nothing to do, you must specify tags to match against
```

### Usage

To control which tags are output you must pass the `-tags=` flag to `pbf2json` and the PBF filepath:

```bash
$ ./build/pbf2json -tags="amenity" /tmp/wellington_new-zealand.osm.pbf
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
# only records with valid street addresses
-tags="addr:housenumber+addr:street"
```

You can also combine the above 2 delimiters to get even more control over what get's returned:

```bash
# only highways and waterways which have a name
-tags="highway+name,waterway+name"
```

### Denormalization

When processing the ways, the node refs are looked up for you and the lat/lon values are added to each way:

```bash
{
  "id": 257577170,
  "type": "way",
  "tags": {
    "building": "yes"
  },
  "nodes": [
    {
      "lat": "-41.317247",
      "lon": "174.794847"
    },
    {
      "lat": "-41.317356",
      "lon": "174.794804"
    },
    {
      "lat": "-41.317408",
      "lon": "174.795076"
    },
    {
      "lat": "-41.317298",
      "lon": "174.795115"
    },
    {
      "lat": "-41.317247",
      "lon": "174.794847"
    }
  ],
  "timestamp": "0001-01-01T00:00:00Z"
}
```

### Leveldb

This library used `leveldb` to store the lat/lon info about nodes so that it can denormalize the ways for you.

By default the leveldb path is set to `/tmp`, you can change where it stores the data with a flag:

```bash
$ ./build/pbf2json -leveldb="/tmp/somewhere"
```

### Run the go code from source

Make sure `Go` is installed and configured on your system, see: https://gist.github.com/missinglink/4212a81a7d9c125b68d9

```bash
sudo apt-get install mercurial;
go get;
go run osm2pbf.go;
```