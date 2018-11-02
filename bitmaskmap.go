package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
)

// BitmaskMap - struct to hold common masks
type BitmaskMap struct {
	Nodes       *Bitmask
	Ways        *Bitmask
	Relations   *Bitmask
	WayRefs     *Bitmask
	RelNodes    *Bitmask
	RelWays     *Bitmask
	RelRelation *Bitmask
}

// NewBitmaskMap - constructor
func NewBitmaskMap() *BitmaskMap {
	return &BitmaskMap{
		Nodes:       NewBitMask(),
		Ways:        NewBitMask(),
		Relations:   NewBitMask(),
		WayRefs:     NewBitMask(),
		RelNodes:    NewBitMask(),
		RelWays:     NewBitMask(),
		RelRelation: NewBitMask(),
	}
}

// WriteTo - write to destination
func (m *BitmaskMap) WriteTo(sink io.Writer) (int64, error) {
	encoder := gob.NewEncoder(sink)
	err := encoder.Encode(m)
	return 0, err
}

// ReadFrom - read from destination
func (m *BitmaskMap) ReadFrom(tap io.Reader) (int64, error) {
	decoder := gob.NewDecoder(tap)
	err := decoder.Decode(m)
	return 0, err
}

// WriteToFile - write to disk
func (m *BitmaskMap) WriteToFile(path string) {
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	m.WriteTo(file)
	log.Println("wrote bitmask:", path)
}

// ReadFromFile - read from disk
func (m *BitmaskMap) ReadFromFile(path string) {

	// bitmask file doesn't exist
	if _, err := os.Stat(path); err != nil {
		fmt.Println("bitmask file not found:", path)
		os.Exit(1)
	}

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	m.ReadFrom(file)
	log.Println("read bitmask:", path)
}

// Print -- print debug stats
func (m BitmaskMap) Print() {
	k := reflect.TypeOf(m)
	v := reflect.ValueOf(m)
	for i := 0; i < k.NumField(); i++ {
		key := k.Field(i).Name
		val := v.Field(i).Interface()
		fmt.Printf("%s: %v\n", key, (val.(*Bitmask)).Len())
	}
}
