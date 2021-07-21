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
	"github.com/syndtr/goleveldb/leveldb/opt"
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
					cacheQueueNode(batch, v)
					if batch.Len() > config.BatchSize {
						cacheFlush(db, batch, true)
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
						cacheFlush(db, batch, true)
					}
				}

				// ----------------
				// write to leveldb
				// note: only write relation member ways
				// ----------------
				if masks.RelWays.Has(v.ID) {

					// write in batches
					cacheQueueWay(batch, v)
					if batch.Len() > config.BatchSize {
						cacheFlush(db, batch, true)
					}
				}

				// bitmask indicates if this is a way of interest
				// if so, print it
				if masks.Ways.Has(v.ID) {

					// lookup from leveldb
					latlons, err := cacheLookupNodes(db, v)

					// skip ways which fail to denormalize
					if err != nil {
						break
					}

					// compute centroid
					centroid, bounds := computeCentroidAndBounds(latlons)

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
						cacheFlush(db, batch, true)
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

					// best centroid and bounds to use
					var largestArea = 0.0
					var centroid map[string]string
					var bounds *geo.Bound

					// iterate over each way, selecting the largest way to use
					// for the centroid and bbox
					for _, latlons := range memberWayLatLons {

						// compute centroid
						wayCentroid, wayBounds := computeCentroidAndBounds(latlons)

						// if for any reason we failed to find a valid bounds
						if nil == wayBounds {
							log.Println("[warn] failed to calculate bounds for relation member way")
							continue
						}

						area := math.Max(wayBounds.GeoWidth(), 0.000001) * math.Max(wayBounds.GeoHeight(), 0.000001)

						// find the way with the largest area
						if area > largestArea {
							largestArea = area
							centroid = wayCentroid
							bounds = wayBounds
						}
					}

					// if for any reason we failed to find a valid bounds
					if nil == bounds {
						log.Println("[warn] denormalize failed for relation:", v.ID, "no valid bounds")
						continue
					}

					// use 'admin_centre' node centroid where available
					// note: only applies to 'boundary=administrative' relations
					// see: https://github.com/pelias/pbf2json/pull/98
					if v.Tags["boundary"] == "administrative" {
						for _, member := range v.Members {
							if member.Type == 0 && member.Role == "admin_centre" {
								if latlons, err := cacheLookupNodeByID(db, member.ID); err == nil {
									latlons["type"] = "admin_centre"
									centroid = latlons
									break
								}
							}
						}
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
func findMemberWayLatLons(db *leveldb.DB, v *osmpbf.Relation) [][]map[string]string {
	var memberWayLatLons [][]map[string]string

	for _, mem := range v.Members {
		if mem.Type == 1 {

			// lookup from leveldb
			latlons, err := cacheLookupWayNodes(db, mem.ID)

			// skip way if it fails to denormalize
			if err != nil {
				break
			}

			memberWayLatLons = append(memberWayLatLons, latlons)
		}
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

// queue a leveldb write in a batch
func cacheQueueNode(batch *leveldb.Batch, node *osmpbf.Node) {
	id, val := nodeToBytes(node)
	batch.Put([]byte(id), []byte(val))
}

// queue a leveldb write in a batch
func cacheQueueWay(batch *leveldb.Batch, way *osmpbf.Way) {
	id, val := wayToBytes(way)
	batch.Put([]byte(id), []byte(val))
}

// flush a leveldb batch to database and reset batch to 0
func cacheFlush(db *leveldb.DB, batch *leveldb.Batch, sync bool) {
	var writeOpts = &opt.WriteOptions{
		NoWriteMerge: true,
		Sync:         sync,
	}

	err := db.Write(batch, writeOpts)
	if err != nil {
		log.Fatal(err)
	}
	batch.Reset()
}

func cacheLookupNodeByID(db *leveldb.DB, id int64) (map[string]string, error) {
	stringid := strconv.FormatInt(id, 10)

	data, err := db.Get([]byte(stringid), nil)
	if err != nil {
		log.Println("[warn] fetch failed for node ID:", stringid)
		return make(map[string]string, 0), err
	}

	return bytesToLatLon(data), nil
}

func cacheLookupNodes(db *leveldb.DB, way *osmpbf.Way) ([]map[string]string, error) {

	var container []map[string]string

	for _, each := range way.NodeIDs {
		stringid := strconv.FormatInt(each, 10)

		data, err := db.Get([]byte(stringid), nil)
		if err != nil {
			log.Println("[warn] denormalize failed for way:", way.ID, "node not found:", stringid)
			return make([]map[string]string, 0), err
		}

		container = append(container, bytesToLatLon(data))
	}

	return container, nil
}

func cacheLookupWayNodes(db *leveldb.DB, wayid int64) ([]map[string]string, error) {

	// prefix the key with 'W' to differentiate it from node ids
	stringid := "W" + strconv.FormatInt(wayid, 10)

	// look up way bytes
	reldata, err := db.Get([]byte(stringid), nil)
	if err != nil {
		log.Println("[warn] lookup failed for way:", wayid, "noderefs not found:", stringid)
		return make([]map[string]string, 0), err
	}

	// generate a way object
	var way = &osmpbf.Way{
		ID:      wayid,
		NodeIDs: bytesToIDSlice(reldata),
	}

	return cacheLookupNodes(db, way)
}

// decode bytes to a 'latlon' type object (map[string]string)
func bytesToLatLon(data []byte) map[string]string {
	latlon := make(map[string]string, 4)
	enc := encoding(data)

	lat64, lon64 := enc.getCoords()
	latlon["lat"] = strconv.FormatFloat(lat64, 'f', 7, 64)
	latlon["lon"] = strconv.FormatFloat(lon64, 'f', 7, 64)

	// check for the bitmask byte which indicates things like an
	// entrance and the level of wheelchair accessibility
	meta := enc.getMetadata()
	if meta.EntranceType > 0 {
		latlon["entrance"] = fmt.Sprintf("%d", meta.EntranceType)
		latlon["wheelchair"] = fmt.Sprintf("%d", meta.AccessibilityType)
	}

	return latlon
}

// encode a node as bytes (between 12 & 13 bytes used)
func nodeToBytes(node *osmpbf.Node) (string, []byte) {
	stringid := strconv.FormatInt(node.ID, 10)

	enc := make(encoding, 14)
	enc.setCoords(node.Lat, node.Lon)

	len := enc.setMetadata(encodingMetadata{
		EntranceType:      entranceType(node),
		AccessibilityType: accessibilityType(node),
	})

	// return variable length encoding
	return stringid, enc[:len]
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

// compute the centroid of a way and its bbox
func computeCentroidAndBounds(latlons []map[string]string) (map[string]string, *geo.Bound) {

	// check to see if there is a tagged entrance we can use.
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
		return selectEntrance(entrances), points.Bound()
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

	return centroid, points.Bound()
}
