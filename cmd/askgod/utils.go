package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/nsec/askgod/internal/utils"
)

func getStructField(base reflect.Value, key string) (reflect.Value, error) {
	field := base

	found := false
	for i := 0; i < field.NumField(); i++ {
		f := field.Field(i)
		ft := field.Type().Field(i)
		if strings.ToLower(ft.Name) == strings.ToLower(key) {
			field = f
			found = true
			break
		}

		if ft.Tag.Get("yaml") == strings.ToLower(key) {
			field = f
			found = true
			break
		}

		if ft.Tag.Get("yaml") == ",inline" {
			subField, err := getStructField(f, key)
			if err == nil {
				field = subField
				found = true
				break
			}
		}
	}

	if !found || !field.IsValid() {
		return reflect.Value{}, fmt.Errorf("Invalid key: %s", key)
	}

	return field, nil
}

func setStructKey(obj interface{}, arg string) error {
	fields := strings.SplitN(arg, "=", 2)
	if len(fields) != 2 {
		return fmt.Errorf("Bad key=value input: %s", arg)
	}

	path := strings.Split(fields[0], ".")
	field := reflect.ValueOf(obj).Elem()

	var err error
	for _, e := range path {
		field, err = getStructField(field, e)
		if err != nil {
			return err
		}
	}

	if field.Type() == reflect.TypeOf("") {
		field.SetString(fields[1])
	} else if field.Type() == reflect.TypeOf(true) {
		if fields[1] == "false" {
			field.SetBool(false)
		} else if fields[1] == "true" {
			field.SetBool(true)
		} else {
			return fmt.Errorf("Bad boolean: %s", fields[1])
		}
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
