package utils

import (
	"encoding/json"
	"log"
)

// ToJSON converts any struct to a formatted JSON string
func ToJSON(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("Error converting to JSON: %v", err)
	}
	return string(jsonData)
}

// DeepCopy accepts a parameter of any type and returns a deep copy of it
func DeepCopy[T any](src T) (T, error) {
	var dst T
	data, err := json.Marshal(src)
	if err != nil {
		return dst, err
	}
	err = json.Unmarshal(data, &dst)
	return dst, err
}
