package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/paulmach/go.geo"
	"github.com/qedus/osmpbf"
	"github.com/syndtr/goleveldb/leveldb"
)

type settings struct {
	PbfPath    string
	LevedbPath string
	Tags       map[string][]string
	BatchSize  int
}

func getSettings() settings {

	// command line flags
	leveldbPath := flag.String("leveldb", "/tmp", "path to leveldb directory")
	tagList := flag.String("tags", "", "comma-separated list of valid tags, group AND conditions with a +")
	batchSize := flag.Int("batch", 50000, "batch leveldb writes in batches of this size")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("invalid args, you must specify a PBF file")
	}

	// invalid tags
	if len(*tagList) < 1 {
		log.Fatal("Nothing to do, you must specify tags to match against")
	}

	// parse tag conditions
	conditions := make(map[string][]string)
	for _, group := range strings.Split(*tagList, ",") {
		conditions[group] = strings.Split(group, "+")
	}

	// fmt.Print(conditions, len(conditions))
	// os.Exit(1)

	return settings{args[0], *leveldbPath, conditions, *batchSize}
}

func main() {

	// configuration
	config := getSettings()

	// open pbf file
	file := openFile(config.PbfPath)
	defer file.Close()

	decoder := osmpbf.NewDecoder(file)
	err := decoder.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	if err != nil {
		log.Fatal(err)
	}

	db := openLevelDB(config.LevedbPath)
	defer db.Close()

	run(decoder, db, config)
}

func run(d *osmpbf.Decoder, db *leveldb.DB, config settings) {

	batch := new(leveldb.Batch)

	var nc, wc, rc uint64
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {

			case *osmpbf.Node:

				// inc count
				nc++

				// ----------------
				// write to leveldb
				// ----------------

				// write immediately
				// cacheStore(db, v)

				// write in batches
				cacheQueue(batch, v)
				if batch.Len() > config.BatchSize {
					cacheFlush(db, batch)
				}

				// ----------------
				// handle tags
				// ----------------

				if !hasTags(v.Tags) {
					break
				}

				v.Tags = trimTags(v.Tags)
				if containsValidTags(v.Tags, config.Tags) {
					onNode(v)
				}

			case *osmpbf.Way:

				// ----------------
				// write to leveldb
				// ----------------

				// flush outstanding batches
				if batch.Len() > 1 {
					cacheFlush(db, batch)
				}

				// inc count
				wc++

				if !hasTags(v.Tags) {
					break
				}

				v.Tags = trimTags(v.Tags)
				if containsValidTags(v.Tags, config.Tags) {

					// lookup from leveldb
					latlons, err := cacheLookup(db, v)

					// skip ways which fail to denormalize
					if err != nil {
						break
					}

					// compute centroid
					var centroid = computeCentroid(latlons)

					onWay(v, latlons, centroid)
				}

			case *osmpbf.Relation:

				// inc count
				rc++

				onRelation(v)

			default:

				log.Fatalf("unknown type %T\n", v)

			}
		}
	}

	// fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)
}

type jsonNode struct {
	ID   int64             `json:"id"`
	Type string            `json:"type"`
	Lat  float64           `json:"lat"`
	Lon  float64           `json:"lon"`
	Tags map[string]string `json:"tags"`
}

func onNode(node *osmpbf.Node) {
	marshall := jsonNode{node.ID, "node", node.Lat, node.Lon, node.Tags}
	json, _ := json.Marshal(marshall)
	fmt.Println(string(json))
}

type jsonWay struct {
	ID   int64             `json:"id"`
	Type string            `json:"type"`
	Tags map[string]string `json:"tags"`
	// NodeIDs   []int64             `json:"refs"`
	Centroid map[string]string   `json:"centroid"`
	Nodes    []map[string]string `json:"nodes"`
}

func onWay(way *osmpbf.Way, latlons []map[string]string, centroid map[string]string) {
	marshall := jsonWay{way.ID, "way", way.Tags /*, way.NodeIDs*/, centroid, latlons}
	json, _ := json.Marshal(marshall)
	fmt.Println(string(json))
}

func onRelation(relation *osmpbf.Relation) {
	// do nothing (yet)
}

// determine if the node is for an entrance
// https://wiki.openstreetmap.org/wiki/Key:entrance
func isEntranceNode(node *osmpbf.Node) uint8 {
	if val, ok := node.Tags["entrance"]; ok {
		var norm = strings.ToLower(val)
		switch norm {
		case "main":
			return 2
		case "yes", "home", "staircase":
			return 1
		}
	}
	return 0
}

// determine if the node is accessible for wheelchair users
// https://wiki.openstreetmap.org/wiki/Key:entrance
func isWheelchairAccessibleNode(node *osmpbf.Node) uint8 {
	if val, ok := node.Tags["wheelchair"]; ok {
		var norm = strings.ToLower(val)
		switch norm {
		case "yes":
			return 2
		case "no":
			return 0
		default:
			return 1
		}
	}
	return 0
}

// write to leveldb immediately
func cacheStore(db *leveldb.DB, node *osmpbf.Node) {
	id, val := nodeToBytes(node)
	err := db.Put([]byte(id), []byte(val), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// queue a leveldb write in a batch
func cacheQueue(batch *leveldb.Batch, node *osmpbf.Node) {
	id, val := nodeToBytes(node)
	batch.Put([]byte(id), []byte(val))
}

// flush a leveldb batch to database and reset batch to 0
func cacheFlush(db *leveldb.DB, batch *leveldb.Batch) {
	err := db.Write(batch, nil)
	if err != nil {
		log.Fatal(err)
	}
	batch.Reset()
}

func cacheLookup(db *leveldb.DB, way *osmpbf.Way) ([]map[string]string, error) {

	var container []map[string]string

	for _, each := range way.NodeIDs {
		stringid := strconv.FormatInt(each, 10)

		data, err := db.Get([]byte(stringid), nil)
		if err != nil {
			log.Println("denormalize failed for way:", way.ID, "node not found:", stringid)
			return container, err
		}

		container = append(container, bytesToLatLon(data))
	}

	return container, nil
}

// decode bytes to a 'latlon' type object
func bytesToLatLon(data []byte) map[string]string {

	var latlon = make(map[string]string)

	var lat64 = math.Float64frombits(binary.BigEndian.Uint64(data[0:8]))
	var lon64 = math.Float64frombits(binary.BigEndian.Uint64(data[8:16]))
	latlon["lat"] = strconv.FormatFloat(lat64, 'f', 7, 64)
	latlon["lon"] = strconv.FormatFloat(lon64, 'f', 7, 64)

	// check for the bitmask byte which indicates things like an
	// entrance and the level of wheelchair accessibility
	if len(data) > 16 {
		latlon["entrance"] = fmt.Sprintf("%d", (data[16]&0xC0)>>6)
		latlon["wheelchair"] = fmt.Sprintf("%d", (data[16]&0x30)>>4)
	}

	return latlon
}

// encode a node as bytes (between 16 & 17 bytes used)
func nodeToBytes(node *osmpbf.Node) (string, []byte) {

	var bufval bytes.Buffer

	// encode lat/lon as two 64 bit floats packed in to 16 bytes
	var bufLatLon = make([]byte, 16)
	binary.BigEndian.PutUint64(bufLatLon[:8], math.Float64bits(node.Lat))
	binary.BigEndian.PutUint64(bufLatLon[8:], math.Float64bits(node.Lon))
	bufval.Write(bufLatLon)

	// generate a bitmask for relevant tag features
	var isEntrance = isEntranceNode(node)
	if isEntrance > 0 {
		// leftmost two bits are for the entrance, next two bits are accessibility
		// remaining 4 rightmost bits are reserved for future use.
		var bitmask = isEntrance << 6
		var isWheelchairAccessible = isWheelchairAccessibleNode(node)
		if isWheelchairAccessible > 0 {
			bitmask |= isWheelchairAccessible << 4
		}
		bufval.WriteByte(bitmask)
	}

	stringid := strconv.FormatInt(node.ID, 10)
	byteval := bufval.Bytes()

	return stringid, byteval
}

func openFile(filename string) *os.File {
	// no file specified
	if len(filename) < 1 {
		log.Fatal("invalid file: you must specify a pbf path as arg[1]")
	}
	// try to open the file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func openLevelDB(path string) *leveldb.DB {
	// try to open the db
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// extract all keys to array
// keys := []string{}
// for k := range v.Tags {
//     keys = append(keys, k)
// }

// check tags contain features from a whitelist
func matchTagsAgainstCompulsoryTagList(tags map[string]string, tagList []string) bool {
	for _, name := range tagList {

		feature := strings.Split(name, "~")
		foundVal, foundKey := tags[feature[0]]

		// key check
		if !foundKey {
			return false
		}

		// value check
		if len(feature) > 1 {
			if foundVal != feature[1] {
				return false
			}
		}
	}

	return true
}

// check tags contain features from a groups of whitelists
func containsValidTags(tags map[string]string, group map[string][]string) bool {
	for _, list := range group {
		if matchTagsAgainstCompulsoryTagList(tags, list) {
			return true
		}
	}
	return false
}

// trim leading/trailing spaces from keys and values
func trimTags(tags map[string]string) map[string]string {
	trimmed := make(map[string]string)
	for k, v := range tags {
		trimmed[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return trimmed
}

// check if a tag list is empty or not
func hasTags(tags map[string]string) bool {
	n := len(tags)
	if n == 0 {
		return false
	}
	return true
}

// select which entrance is preferable
func selectEntrance(entrances []map[string]string) map[string]string {

	// use the mapped entrance location where available
	var centroid = make(map[string]string)
	centroid["type"] = "entrance"

	// prefer the first 'main' entrance we find (should usually only be one).
	for _, entrance := range entrances {
		if val, ok := entrance["entrance"]; ok && val == "2" {
			centroid["lat"] = entrance["lat"]
			centroid["lon"] = entrance["lon"]
			return centroid
		}
	}

	// else prefer the first wheelchair accessible entrance we find
	for _, entrance := range entrances {
		if val, ok := entrance["wheelchair"]; ok && val != "0" {
			centroid["lat"] = entrance["lat"]
			centroid["lon"] = entrance["lon"]
			return centroid
		}
	}

	// otherwise just take the first entrance in the list
	centroid["lat"] = entrances[0]["lat"]
	centroid["lon"] = entrances[0]["lon"]
	return centroid
}

// compute the centroid of a way
func computeCentroid(latlons []map[string]string) map[string]string {

	// check to see if there is a tagged entrance we can use.
	var entrances []map[string]string
	for _, latlon := range latlons {
		if _, ok := latlon["entrance"]; ok {
			entrances = append(entrances, latlon)
		}
	}

	// use the mapped entrance location where available
	if len(entrances) > 0 {
		return selectEntrance(entrances)
	}

	// convert lat/lon map to geo.PointSet
	points := geo.NewPointSet()
	for _, each := range latlons {
		var lon, _ = strconv.ParseFloat(each["lon"], 64)
		var lat, _ = strconv.ParseFloat(each["lat"], 64)
		points.Push(geo.NewPoint(lon, lat))
	}

	// determine if the way is a closed centroid or a linestring
	// by comparing first and last coordinates.
	isClosed := false
	if points.Length() > 2 {
		isClosed = points.First().Equals(points.Last())
	}

	// compute the centroid using one of two different algorithms
	var compute *geo.Point
	if isClosed {
		compute = GetPolygonCentroid(points)
	} else {
		compute = GetLineCentroid(points)
	}

	// return point as lat/lon map
	var centroid = make(map[string]string)
	centroid["lat"] = strconv.FormatFloat(compute.Lat(), 'f', 7, 64)
	centroid["lon"] = strconv.FormatFloat(compute.Lng(), 'f', 7, 64)

	return centroid
}
