package main

import (
	"encoding/binary"
	"math"
)

// an empty byte slice used for zeroing
var zeros = make([]byte, 64)

// node encoding
type encoding []byte

type encodingMetadata struct {
	EntranceType      uint8
	AccessibilityType uint8
}

// encode lat/lon as 64 bit floats packed in to 8 bytes,
// each float is then truncated to 6 bytes because we don't
// need the additional precision (> 8 decimal places)
func (e encoding) setCoords(lat float64, lon float64) {
	binary.BigEndian.PutUint64(e[0:8], math.Float64bits(lat))
	binary.BigEndian.PutUint64(e[6:14], math.Float64bits(lon))
	copy(e[12:14], zeros)
}

// decode lat/lon as truncated 64 bit floats from the first 12 bytes
func (e encoding) getCoords() (float64, float64) {
	buffer := make([]byte, 8)

	copy(buffer, e[:6])
	lat := math.Float64frombits(binary.BigEndian.Uint64(buffer))

	copy(buffer, e[6:12])
	lon := math.Float64frombits(binary.BigEndian.Uint64(buffer))

	return lat, lon
}

// leftmost two bits are for the entrance, next two bits are accessibility
// remaining 4 rightmost bits are reserved for future use.
func (e encoding) setMetadata(meta encodingMetadata) int {
	if meta.EntranceType > 0 {
		e[12] = ((meta.EntranceType & 0b00000011) << 6)       // values [0,1,2] (stored in leftmost two bits)
		e[12] |= ((meta.AccessibilityType & 0b00000011) << 4) // values [0,1,2] (stored in next two bits)

		// 13 byte encoding
		return 13
	}

	// 12 byte encoding
	return 12
}

func (e encoding) getMetadata() encodingMetadata {
	meta := encodingMetadata{}

	if len(e) > 12 {
		meta.EntranceType = (e[12] & 0b11000000) >> 6
		meta.AccessibilityType = (e[12] & 0b00110000) >> 4
	}

	return meta
}
