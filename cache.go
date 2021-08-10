package main

import (
	"github.com/qedus/osmpbf"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"log"
	"strconv"
)

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
