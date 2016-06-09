
package main

import (
  "encoding/json"
  "fmt"
  "flag"
  "bytes"
  "os"
  "log"
  "io"
  "time"
  "runtime"
  "sort"
  "strings"
  "strconv"
  "github.com/qedus/osmpbf"
  "github.com/syndtr/goleveldb/leveldb"
  "github.com/paulmach/go.geo"
  "math"
)

type Settings struct {
  PbfPath           string
  LevedbPath        string
  Tags              map[string][]string
  Ids               []string
  BatchSize         int
}

func getSettings() Settings {

  // command line flags
  leveldbPath := flag.String("leveldb", "/tmp", "path to leveldb directory")
  tagList := flag.String("tags", "", "comma-separated list of valid tags, group AND conditions with a +")
  idList := flag.String("ids", "", "comma-separated list of valid ids")
  batchSize := flag.Int("batch", 500, "batch leveldb writes in batches of this size")

  flag.Parse()
  args := flag.Args();

  tagsConditions := make(map[string][]string)
  idsConditions := make([]string,0,0)

  if len( args ) < 1 {
    log.Fatal("invalid args, you must specify a PBF file")
  }

  if len(*tagList) > 0 && len(*idList) > 0 {
    log.Fatal("Please use either the -tags option or the -ids option, but not both at the same time")
  } else if len(*tagList) < 1 && len(*idList) < 1 {
    if len(*tagList) < 1 {
      log.Fatal("Nothing to do, you must specify tags to match against")
    } else { // len(*idList) < 1
      log.Fatal("Nothing to do, you must specify id to match against")
    }
  } else if len(*tagList) > 0 {
    // parse tag conditions
    for _, group := range strings.Split(*tagList,",") {
      tagsConditions[group] = strings.Split(group,"+")
    }
  } else {
    // parse id conditions
    for _, value := range strings.Split(*idList,",") {
      idsConditions = append(idsConditions, value)
      sort.Strings(idsConditions)
    }
  }

  // fmt.Println(tagsConditions, len(tagsConditions))
  // fmt.Println(idsConditions, len(idsConditions))
  // os.Exit(1)

  return Settings{ args[0], *leveldbPath, tagsConditions, idsConditions, *batchSize }
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

func run(d *osmpbf.Decoder, db *leveldb.DB, config Settings){

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
          // handle conditions
          // ----------------

          if len(config.Ids) != 0 {
            i := sort.SearchStrings(config.Ids, strconv.FormatInt(v.ID,10))
            if i < len(config.Ids) && config.Ids[i] == strconv.FormatInt(v.ID,10) {
              onNode(v)
            }
          } else {
            if !hasTags(v.Tags) { break }

            v.Tags = trimTags(v.Tags)
            if containsValidTags( v.Tags, config.Tags ) {
              onNode(v)
            }
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

          if len(config.Ids) != 0 {
            i := sort.SearchStrings(config.Ids, strconv.FormatInt(v.ID,10))
            if i < len(config.Ids) && config.Ids[i] == strconv.FormatInt(v.ID,10) {
              // lookup from leveldb
              latlons, err := cacheLookup(db, v)

              // skip ways which fail to denormalize
              if err != nil { break }

              // compute centroid
              var centroid = computeCentroid(latlons);

              onWay(v,latlons,centroid)
            }
          } else {
            if !hasTags(v.Tags) { break }

            v.Tags = trimTags(v.Tags)
            if containsValidTags( v.Tags, config.Tags ) {

              // lookup from leveldb
              latlons, err := cacheLookup(db, v)

              // skip ways which fail to denormalize
              if err != nil { break }

              // compute centroid
              var centroid = computeCentroid(latlons);

              onWay(v,latlons,centroid)
            }
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

type JsonNode struct {
  ID        int64               `json:"id"`
  Type      string              `json:"type"`
  Lat       float64             `json:"lat"`
  Lon       float64             `json:"lon"`
  Tags      map[string]string   `json:"tags"`
  Timestamp time.Time           `json:"timestamp"`
}

func onNode(node *osmpbf.Node){
  marshall := JsonNode{ node.ID, "node", node.Lat, node.Lon, node.Tags, node.Info.Timestamp }
  json, _ := json.Marshal(marshall)
  fmt.Println(string(json))
}

type JsonWay struct {
  ID        int64               `json:"id"`
  Type      string              `json:"type"`
  Tags      map[string]string   `json:"tags"`
  // NodeIDs   []int64             `json:"refs"`
  Centroid  map[string]string   `json:"centroid"`
  Nodes     []map[string]string `json:"nodes"`
  Timestamp time.Time           `json:"timestamp"`
}

func onWay(way *osmpbf.Way, latlons []map[string]string, centroid map[string]string){
  marshall := JsonWay{ way.ID, "way", way.Tags/*, way.NodeIDs*/, centroid, latlons, way.Info.Timestamp }
  json, _ := json.Marshal(marshall)
  fmt.Println(string(json))
}

func onRelation(relation *osmpbf.Relation){
  // do nothing (yet)
}

// write to leveldb immediately
func cacheStore(db *leveldb.DB, node *osmpbf.Node){
  id, val := formatLevelDB(node)
  err := db.Put([]byte(id), []byte(val), nil)
  if err != nil {
    log.Fatal(err)
  }
}

// queue a leveldb write in a batch
func cacheQueue(batch *leveldb.Batch, node *osmpbf.Node){
  id, val := formatLevelDB(node)
  batch.Put([]byte(id), []byte(val))
}

// flush a leveldb batch to database and reset batch to 0
func cacheFlush(db *leveldb.DB, batch *leveldb.Batch){
  err := db.Write(batch, nil)
  if err != nil {
    log.Fatal(err)
  }
  batch.Reset()
}

func cacheLookup(db *leveldb.DB, way *osmpbf.Way) ([]map[string]string, error) {

  var container []map[string]string

  for _, each := range way.NodeIDs {
    stringid := strconv.FormatInt(each,10)

    data, err := db.Get([]byte(stringid), nil)
    if err != nil {
      log.Println("denormalize failed for way:", way.ID, "node not found:", stringid)
      return container, err
    }

    s := string(data)
    spl := strings.Split(s, ":");

    latlon := make(map[string]string)
    lat, lon := spl[0], spl[1]
    latlon["lat"] = lat
    latlon["lon"] = lon

    container = append(container, latlon)

  }

  return container, nil

  // fmt.Println(way.NodeIDs)
  // fmt.Println(container)
  // os.Exit(1)
}

func formatLevelDB(node *osmpbf.Node) (id string, val []byte){

  stringid := strconv.FormatInt(node.ID,10)

  var bufval bytes.Buffer
  bufval.WriteString(strconv.FormatFloat(node.Lat,'f',6,64))
  bufval.WriteString(":")
  bufval.WriteString(strconv.FormatFloat(node.Lon,'f',6,64))
  byteval := []byte(bufval.String())
  
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

    feature := strings.Split(name,"~")
    foundVal, foundKey := tags[feature[0]]

    // key check
    if !foundKey {
      return false
    }

    // value check
    if len( feature ) > 1 {
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
    if matchTagsAgainstCompulsoryTagList( tags, list ){
      return true
    }
  }
  return false
}

// trim leading/trailing spaces from keys and values
func trimTags(tags map[string]string) map[string]string {
  trimmed := make(map[string]string)
  for k, v := range tags {
    trimmed[strings.TrimSpace(k)] = strings.TrimSpace(v);
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

// compute the centroid of a way
func computeCentroid(latlons []map[string]string) map[string]string {

  points := geo.PointSet{}
  for _, each := range latlons {
    var lon, _ = strconv.ParseFloat( each["lon"], 64 );
    var lat, _ = strconv.ParseFloat( each["lat"], 64 );
    points.Push( geo.NewPoint( lon, lat ))
  }

  var compute = getCentroid(points);

  var centroid = make(map[string]string)
  centroid["lat"] = strconv.FormatFloat(compute.Lat(),'f',6,64)
  centroid["lon"] = strconv.FormatFloat(compute.Lng(),'f',6,64)

  return centroid
}

// compute the centroid of a polygon set
// using a spherical co-ordinate system
func getCentroid(ps geo.PointSet) *geo.Point {

  X := 0.0
  Y := 0.0
  Z := 0.0

  var toRad = math.Pi / 180
  var fromRad = 180 / math.Pi

  for _, point := range ps {

    var lon = point[0] * toRad
    var lat = point[1] * toRad

    X += math.Cos(lat) * math.Cos(lon)
    Y += math.Cos(lat) * math.Sin(lon)
    Z += math.Sin(lat)
  }

  numPoints := float64(len(ps))
  X = X / numPoints
  Y = Y / numPoints
  Z = Z / numPoints

  var lon = math.Atan2(Y, X)
  var hyp = math.Sqrt(X * X + Y * Y)
  var lat = math.Atan2(Z, hyp)

  var centroid = geo.NewPoint(lon * fromRad, lat * fromRad)

  return centroid;
}
