package main

import (
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

	geo "github.com/paulmach/go.geo"
	"github.com/qedus/osmpbf"
	"github.com/syndtr/goleveldb/leveldb"
)

type settings struct {
	PbfPath    string
	LevedbPath string
	Tags       map[string][]string
	BatchSize  int
	WayNodes   bool
}

var emptyLatLons = make([]map[string]string, 0)

func getSettings() settings {

	// command line flags
	leveldbPath := flag.String("leveldb", "/tmp", "path to leveldb directory")
	tagList := flag.String("tags", "", "comma-separated list of valid tags, group AND conditions with a +")
	batchSize := flag.Int("batch", 50000, "batch leveldb writes in batches of this size")
	wayNodes := flag.Bool("waynodes", false, "should the lat/lons of nodes belonging to ways be printed")

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

	return settings{args[0], *leveldbPath, conditions, *batchSize, *wayNodes}
}

func main() {

	// configuration
	config := getSettings()

	// open pbf file
	file := openFile(config.PbfPath)
	defer file.Close()

	// perform two passes over the file, on the first pass
	// we record a bitmask of the interesting elements in the
	// file, on the second pass we extract the data

	// set up bimasks
	var masks = NewBitmaskMap()

	// set up leveldb connection
	var db = openLevelDB(config.LevedbPath)
	defer db.Close()

	// === first pass (indexing) ===
	idxDecoder := osmpbf.NewDecoder(file)
	err := idxDecoder.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	if err != nil {
		log.Fatal(err)
	}

	// index target IDs in bitmasks
	index(idxDecoder, masks, config)

	// no-op if no relation members of type 'way' present in mask
	if !masks.RelWays.Empty() {
		// === potential second pass (indexing) to index members of relations ===
		file.Seek(io.SeekStart, 0) // rewind file
		idxRelationsDecoder := osmpbf.NewDecoder(file)
		err = idxRelationsDecoder.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
		if err != nil {
			log.Fatal(err)
		}

		// index relation member IDs in bitmasks
		indexRelationMembers(idxRelationsDecoder, masks, config)
	}

	// === final pass (printing json) ===
	file.Seek(io.SeekStart, 0) // rewind file
	decoder := osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	if err != nil {
		log.Fatal(err)
	}

	// print json
	print(decoder, masks, db, config)
}

func index(d *osmpbf.Decoder, masks *BitmaskMap, config settings) {
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {

			case *osmpbf.Node:
				if hasTags(v.Tags) && containsValidTags(v.Tags, config.Tags) {
					masks.Nodes.Insert(v.ID)
				}

			case *osmpbf.Way:
				if hasTags(v.Tags) && containsValidTags(v.Tags, config.Tags) {
					masks.Ways.Insert(v.ID)
					for _, nodeid := range v.NodeIDs {
						masks.WayRefs.Insert(nodeid)
					}
				}

			case *osmpbf.Relation:
				if hasTags(v.Tags) && containsValidTags(v.Tags, config.Tags) {

					// record a count of which type of members
					// are present in the relation
					var count = make(map[int]int64)
					for _, member := range v.Members {
						count[int(member.Type)]++
					}

					// skip relations which contain 0 ways
					if count[1] == 0 {
						continue
					}

					masks.Relations.Insert(v.ID)
					for _, member := range v.Members {
						switch member.Type {
						case 0: // node
							masks.RelNodes.Insert(member.ID)
						case 1: // way
							masks.RelWays.Insert(member.ID)
						case 2: // relation
							masks.RelRelation.Insert(member.ID)
						}
					}
				}
			}
		}
	}
}

func indexRelationMembers(d *osmpbf.Decoder, masks *BitmaskMap, config settings) {
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Way:
				if masks.RelWays.Has(v.ID) {
					for _, nodeid := range v.NodeIDs {
						masks.RelNodes.Insert(nodeid)
					}
				}
				// support for super-relations
				// case *osmpbf.Relation:
				// 	if masks.RelRelation.Has(v.ID) {
				// 		for _, member := range v.Members {
				// 			switch member.Type {
				// 			case 0: // node
				// 				masks.RelNodes.Insert(member.ID)
				// 			case 1: // way
				// 				masks.RelWays.Insert(member.ID)
				// 			}
				// 		}
				// 	}
			}
		}
	}
}

func print(d *osmpbf.Decoder, masks *BitmaskMap, db *leveldb.DB, config settings) {

	batch := new(leveldb.Batch)
	finishedNodes := false
	finishedWays := false

	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {

			case *osmpbf.Node:

				// ----------------
				// write to leveldb
				// note: only write way refs and relation member nodes
				// ----------------
				if masks.WayRefs.Has(v.ID) || masks.RelNodes.Has(v.ID) {

					// write in batches
					CacheQueueNode(batch, v)
					if batch.Len() > config.BatchSize {
						CacheFlush(db, batch, true)
					}
				}

				// bitmask indicates if this is a node of interest
				// if so, print it
				if masks.Nodes.Has(v.ID) {

					// trim tags
					v.Tags = trimTags(v.Tags)
					onNode(v)
				}

			case *osmpbf.Way:

				// ----------------
				// write to leveldb
				// flush outstanding node batches
				// before processing any ways
				// ----------------
				if !finishedNodes {
					finishedNodes = true
					if batch.Len() > 1 {
						CacheFlush(db, batch, true)
					}
				}

				// ----------------
				// write to leveldb
				// note: only write relation member ways
				// ----------------
				if masks.RelWays.Has(v.ID) {

					// write in batches
					CacheQueueWay(batch, v)
					if batch.Len() > config.BatchSize {
						CacheFlush(db, batch, true)
					}
				}

				// bitmask indicates if this is a way of interest
				// if so, print it
				if masks.Ways.Has(v.ID) {

					// lookup from leveldb
					latlons, err := CacheLookupNodes(db, v)

					// skip ways which fail to denormalize
					if err != nil {
						break
					}

					// compute centroid
					centroid, bounds := ComputeCentroidAndBounds(latlons)

					// trim tags
					v.Tags = trimTags(v.Tags)

					if config.WayNodes {
						onWay(v, latlons, centroid, bounds)
					} else {
						onWay(v, emptyLatLons, centroid, bounds)
					}
				}

			case *osmpbf.Relation:

				// ----------------
				// write to leveldb
				// flush outstanding way batches
				// before processing any relation
				// ----------------
				if !finishedWays {
					finishedWays = true
					if batch.Len() > 1 {
						CacheFlush(db, batch, true)
					}
				}

				// bitmask indicates if this is a relation of interest
				// if so, print it
				if masks.Relations.Has(v.ID) {

					// fetch all latlons for all ways in relation
					var memberWayLatLons = findMemberWayLatLons(db, v)

					// no ways found, skip relation
					if len(memberWayLatLons) == 0 {
						log.Println("[warn] denormalize failed for relation:", v.ID, "no ways found")
						continue
					}

					// consult https://wiki.openstreetmap.org/wiki/DE:Relation:multipolygon on how osm handles outer members
					centroid, bounds := ComputeRelationCentroidAndBounds(memberWayLatLons)

					if centroid == nil || bounds == nil {
						// the relation is probably not a whole part of the osm dump
						log.Printf("[warn] could not find centroid and bounds of %d", v.ID)
						continue
					}

					// trim tags
					v.Tags = trimTags(v.Tags)

					// print relation
					onRelation(v, centroid, bounds)
				}

			default:

				log.Fatalf("[error] unknown type %T\n", v)

			}
		}
	}
}

// lookup all latlons for all ways in relation
func findMemberWayLatLons(db *leveldb.DB, v *osmpbf.Relation) map[osmpbf.Member][]map[string]string {
	var memberWayLatLons = make(map[osmpbf.Member][]map[string]string)

	for _, mem := range v.Members {
		// lookup from leveldb
		latlons, err := CacheLookupWayNodes(db, mem.ID)

		// skip way if it fails to denormalize
		if err != nil {
			break
		}

		memberWayLatLons[mem] = latlons
	}

	return memberWayLatLons
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
	Bounds   map[string]string   `json:"bounds"`
	Nodes    []map[string]string `json:"nodes,omitempty"`
}

func jsonBbox(bounds *geo.Bound) map[string]string {
	// render a North-South-East-West bounding box
	var bbox = make(map[string]string)
	bbox["n"] = strconv.FormatFloat(bounds.North(), 'f', 7, 64)
	bbox["s"] = strconv.FormatFloat(bounds.South(), 'f', 7, 64)
	bbox["e"] = strconv.FormatFloat(bounds.East(), 'f', 7, 64)
	bbox["w"] = strconv.FormatFloat(bounds.West(), 'f', 7, 64)

	return bbox
}

func onWay(way *osmpbf.Way, latlons []map[string]string, centroid map[string]string, bounds *geo.Bound) {
	bbox := jsonBbox(bounds)
	marshall := jsonWay{way.ID, "way", way.Tags /*, way.NodeIDs*/, centroid, bbox, latlons}
	json, _ := json.Marshal(marshall)
	fmt.Println(string(json))
}

type jsonRelation struct {
	ID       int64             `json:"id"`
	Type     string            `json:"type"`
	Tags     map[string]string `json:"tags"`
	Centroid map[string]string `json:"centroid"`
	Bounds   map[string]string `json:"bounds"`
}

func onRelation(relation *osmpbf.Relation, centroid map[string]string, bounds *geo.Bound) {
	bbox := jsonBbox(bounds)
	marshall := jsonRelation{relation.ID, "relation", relation.Tags, centroid, bbox}
	json, _ := json.Marshal(marshall)
	fmt.Println(string(json))
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

// decode bytes to a 'latlon' type object
func bytesToLatLon(data []byte) map[string]string {
	buf := make([]byte, 8)
	latlon := make(map[string]string, 4)

	// first 6 bytes are the latitude
	// buf = append(buf, data[0:6]...)
	copy(buf, data[:6])
	lat64 := math.Float64frombits(binary.BigEndian.Uint64(buf[:8]))
	latlon["lat"] = strconv.FormatFloat(lat64, 'f', 7, 64)

	// next 6 bytes are the longitude
	// buf = append(buf[:0], data[6:12]...)
	copy(buf, data[6:12])
	lon64 := math.Float64frombits(binary.BigEndian.Uint64(buf[:8]))
	latlon["lon"] = strconv.FormatFloat(lon64, 'f', 7, 64)

	// check for the bitmask byte which indicates things like an
	// entrance and the level of wheelchair accessibility
	if len(data) > 12 {
		latlon["entrance"] = fmt.Sprintf("%d", (data[12]&0xC0)>>6)
		latlon["wheelchair"] = fmt.Sprintf("%d", (data[12]&0x30)>>4)
	}

	return latlon
}

// encode a node as bytes (between 12 & 13 bytes used)
func nodeToBytes(node *osmpbf.Node) (string, []byte) {
	stringid := strconv.FormatInt(node.ID, 10)

	buf := make([]byte, 14)
	// encode lat/lon as 64 bit floats packed in to 8 bytes,
	// each float is then truncated to 6 bytes because we don't
	// need the additional precision (> 8 decimal places)

	binary.BigEndian.PutUint64(buf, math.Float64bits(node.Lat))
	binary.BigEndian.PutUint64(buf[6:], math.Float64bits(node.Lon))

	// generate a bitmask for relevant tag features
	isEntrance := isEntranceNode(node)
	if isEntrance == 0 {
		return stringid, buf[:12]
	}

	// leftmost two bits are for the entrance, next two bits are accessibility
	// remaining 4 rightmost bits are reserved for future use.
	bitmask := isEntrance << 6
	bitmask |= isWheelchairAccessibleNode(node) << 4
	buf[12] = bitmask

	return stringid, buf[:13]
}

func idSliceToBytes(ids []int64) []byte {
	buf := make([]byte, 8*len(ids))
	for i, id := range ids {
		binary.BigEndian.PutUint64(buf[8*i:], uint64(id))
	}
	return buf
}

func bytesToIDSlice(bytes []byte) []int64 {
	if len(bytes)%8 != 0 {
		log.Fatal("invalid byte slice length: not divisible by 8")
	}

	ids := make([]int64, len(bytes)/8)
	for i := 0; i < len(bytes)/8; i++ {
		ids[i] = int64(binary.BigEndian.Uint64(bytes[8*i:]))
	}
	return ids
}

// encode a way as bytes (repeated int64 numbers)
func wayToBytes(way *osmpbf.Way) (string, []byte) {
	// prefix the key with 'W' to differentiate it from node ids
	stringid := "W" + strconv.FormatInt(way.ID, 10)
	return stringid, idSliceToBytes(way.NodeIDs)
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

func getEntrance(latlons []map[string]string) (bool, map[string]string, *geo.Bound) {
	var entrances []map[string]string
	for _, latlon := range latlons {
		if _, ok := latlon["entrance"]; ok {
			entrances = append(entrances, latlon)
		}
	}

	// convert lat/lon map to geo.PointSet
	points := geo.NewPointSet()
	for _, each := range latlons {
		var lon, _ = strconv.ParseFloat(each["lon"], 64)
		var lat, _ = strconv.ParseFloat(each["lat"], 64)
		points.Push(geo.NewPoint(lon, lat))
	}

	// use the mapped entrance location where available
	if len(entrances) > 0 {
		return true, selectEntrance(entrances), points.Bound()
	}

	return false, nil, nil
}

// ComputeCentroidAndBounds compute the centroid of a way and its bbox for polygons and lines
func ComputeCentroidAndBounds(latlons []map[string]string) (map[string]string, *geo.Bound) {
	hasEntrance, entranceLatLon, entranceBounds := getEntrance(latlons)

	if hasEntrance {
		return entranceLatLon, entranceBounds
	}

	// convert lat/lon map to geo.PointSet
	points := LatLngMapToPointSet(latlons)

	// determine if the way is a closed centroid or a linestring
	// by comparing first and last coordinates.
	isClosed := IsPointSetClosed(points)

	// compute the centroid using one of two different algorithms
	var compute *geo.Point
	if isClosed {
		compute = GetPolygonCentroid(points)
	} else {
		compute = GetLineCentroid(points)
	}

	// return point as lat/lon map
	centroid := PointToLatLon(compute)

	return centroid, points.Bound()
}
