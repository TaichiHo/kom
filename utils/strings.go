package utils

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func TruncateString(s string, length int) string {
	if utf8.RuneCountInString(s) <= length {
		return s
	}

	// Convert the string to a slice of runes
	runes := []rune(s)
	return string(runes[:length])
}

func ToInt(s string) int {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return id
}

func ToIntDefault(s string, i int) int {
	id, err := strconv.Atoi(s)
	if err != nil {
		return i
	}
	return id
}

func ToUInt(s string) uint {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	if id > math.MaxUint {
		return 0
	}
	return uint(id)
}

// ToIntSlice converts a comma-separated string of numbers into a []int slice
func ToIntSlice(ids string) []int {
	// Split the string
	strIds := strings.Split(ids, ",")
	var intIds []int

	// Iterate through the string array and convert to integers
	for _, strId := range strIds {
		strId = strings.TrimSpace(strId) // Remove leading and trailing spaces
		if id, err := strconv.Atoi(strId); err == nil {
			intIds = append(intIds, id)
		}
	}

	return intIds
}
func ToInt64Slice(ids string) []int64 {
	// Split the string
	strIds := strings.Split(ids, ",")
	var intIds []int64

	// Iterate through the string array and convert to integers
	for _, strId := range strIds {
		strId = strings.TrimSpace(strId) // Remove leading and trailing spaces
		if id, err := strconv.ParseInt(strId, 10, 64); err == nil {
			intIds = append(intIds, id)
		}
	}

	return intIds
}
func ToInt64(str string) int64 {
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
func IsTextFile(ob []byte) (bool, error) {

	n := len(ob)
	if n > 1024 {
		n = 1024
	}
	// Check for non-text characters
	if !utf8.Valid(ob[:n]) {
		return false, nil
	}

	// Check for null bytes (\x00), which typically indicate a binary file
	if bytes.Contains(ob[:n], []byte{0}) {
		return false, nil
	}

	return true, nil
}

// SanitizeFileName removes illegal characters from filenames and replaces them with underscores '_'
func SanitizeFileName(filename string) string {
	// Define regex for illegal characters, including \ / : * ? " < > | and parentheses ()
	reg := regexp.MustCompile(`[\\/:*?"<>|()]+`)

	// Replace illegal characters with underscore '_'
	sanitizedFilename := reg.ReplaceAllString(filename, "_")

	return sanitizedFilename
}
func NormalizeNewlines(input string) string {
	// Replace Windows-style newlines \r\n with Unix-style \n
	return strings.ReplaceAll(input, "\r\n", "\n")
}
func NormalizeToWindows(input string) string {
	// First normalize to \n then replace with \r\n to prevent duplicate replacements
	unixNormalized := strings.ReplaceAll(input, "\r\n", "\n")
	return strings.ReplaceAll(unixNormalized, "\n", "\r\n")
}
func TrimQuotes(str string) string {
	str = strings.TrimPrefix(str, "`")
	str = strings.TrimSuffix(str, "`")
	str = strings.TrimPrefix(str, "'")
	str = strings.TrimSuffix(str, "'")
	return str
}

func ParseTime(value string) (time.Time, error) {
	// Try different time formats
	layouts := []string{
		time.RFC3339,                    // "2006-01-02T15:04:05Z07:00"
		"2006-01-02",                    // "2006-01-02" (date)
		"2006-01-02 15:04:05",           // "2006-01-02 15:04:05" (no timezone)
		"2006-01-02 15:04:05 -0700 MST", // "2006-01-02 15:04:05 -0700 MST" (with timezone) 2024-12-05 14:11:44 +0000 UTC
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}

// StringListToSQLIn converts a string list to a format used in SQL IN clause, e.g., ('x', 'xx')
// For example, ['x', 'xx'] is converted to ('x', 'xx')
func StringListToSQLIn(strList []string) string {
	if len(strList) == 0 {
		return "()"
	}

	var parts []string
	for _, str := range strList {
		parts = append(parts, fmt.Sprintf("'%s'", str))
	}

	return "(" + strings.Join(parts, ", ") + ")"
}
