package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// PathExists returns true if the path exists.
func PathExists(path string) bool {
	_, err := os.Lstat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

// StringInSlice returns true if the provided string is in the provided slice.
func StringInSlice(key string, list []string) bool {
	for _, entry := range list {
		if entry == key {
			return true
		}
	}

	return false
}

// Int64InSlice returns true if the provided int64 is in the provided slice.
func Int64InSlice(key int64, list []int64) bool {
	for _, entry := range list {
		if entry == key {
			return true
		}
	}

	return false
}

// ParseTags converts a serialized tags list to map[string]string.
func ParseTags(in string) (map[string]string, error) {
	out := map[string]string{}

	for _, entry := range strings.Split(in, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		fields := strings.SplitN(entry, ":", 2)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid tag: %s", entry)
		}

		_, ok := out[fields[0]]
		if ok {
			return nil, fmt.Errorf("duplicate tag: %s", entry)
		}

		out[fields[0]] = fields[1]
	}

	return out, nil
}

// PackTags converts map[string]string to a serialized tags list.
func PackTags(in map[string]string) string {
	tags := []string{}

	for k, v := range in {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	sort.Strings(tags)

	return strings.Join(tags, ",")
}
