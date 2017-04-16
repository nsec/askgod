package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/nsec/askgod/internal/utils"
)

func setStructKey(obj interface{}, arg string) error {
	v := reflect.ValueOf(obj)

	fields := strings.SplitN(arg, "=", 2)
	if len(fields) != 2 {
		return fmt.Errorf("Bad key=value input: %s", arg)
	}

	field := v.Elem().FieldByNameFunc(func(name string) bool {
		if strings.ToLower(name) == strings.ToLower(fields[0]) {
			return true
		}

		return false
	})

	if !field.IsValid() {
		return fmt.Errorf("Invalid key: %s", fields[0])
	}

	if field.Type() == reflect.TypeOf("") {
		field.SetString(fields[1])
	} else if field.Type() == reflect.TypeOf(int64(0)) {
		intValue, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return err
		}

		field.SetInt(intValue)
	} else if field.Type() == reflect.TypeOf(map[string]string{}) {
		tags, err := utils.ParseTags(fields[1])
		if err != nil {
			return err
		}

		tagsValue := reflect.ValueOf(tags)
		field.Set(tagsValue)
	} else {
		return fmt.Errorf("Unsupported type for key: %s", fields[0])
	}

	return nil
}
