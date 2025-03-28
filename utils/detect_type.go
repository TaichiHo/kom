package utils

import (
	"fmt"
	"strconv"
)

// Define string types
const (
	TypeNumber  = "number"
	TypeTime    = "time"
	TypeString  = "string"
	TypeBoolean = "boolean"
)

// DetectType detects the type of a string (number, time, string)
func DetectType(value interface{}) (string, interface{}) {

	if boolean, err := strconv.ParseBool(fmt.Sprintf("%v", value)); err == nil {
		return TypeBoolean, boolean
	}

	// 1. Try to parse as integer or float
	if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
		return TypeNumber, num
	}

	if t, err := ParseTime(fmt.Sprintf("%v", value)); err == nil {
		return TypeTime, t
	}

	// 3. Return string type by default
	return TypeString, value
}
