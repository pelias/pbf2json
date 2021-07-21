package main

import (
	"strings"

	"github.com/qedus/osmpbf"
)

const (
	EntranceNone uint8 = iota
	EntranceNormal
	EntranceMain
)

const (
	WheelchairAccessibleNo uint8 = iota
	WheelchairAccessibleImplicitYes
	WheelchairAccessibleExplicitYes
)

// determine if the node is for an entrance
// https://wiki.openstreetmap.org/wiki/Key:entrance
func entranceType(node *osmpbf.Node) uint8 {
	if val, ok := node.Tags["entrance"]; ok {
		var norm = strings.ToLower(val)
		switch norm {
		case "main":
			return EntranceMain
		case "yes", "home", "staircase":
			return EntranceNormal
		}
	}
	return EntranceNone
}

// determine if the node is accessible for wheelchair users
// https://wiki.openstreetmap.org/wiki/Key:entrance
func accessibilityType(node *osmpbf.Node) uint8 {
	if val, ok := node.Tags["wheelchair"]; ok {
		var norm = strings.ToLower(val)
		switch norm {
		case "yes":
			return WheelchairAccessibleExplicitYes
		case "no":
			return WheelchairAccessibleNo
		default:
			return WheelchairAccessibleImplicitYes
		}
	}
	return WheelchairAccessibleNo
}
