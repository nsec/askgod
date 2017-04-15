package utils

import (
	"os"
)

// PathExists returns true if the path exists
func PathExists(path string) bool {
	_, err := os.Lstat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

// StringInSlice returns true if the provided string is in the provided slice
func StringInSlice(key string, list []string) bool {
	for _, entry := range list {
		if entry == key {
			return true
		}
	}
	return false
}
