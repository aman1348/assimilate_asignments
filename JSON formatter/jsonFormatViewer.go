package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
)

func prettyPrint(data interface{}, indent string) {
	d := reflect.ValueOf(data)

	switch d.Kind() {
	case reflect.Map:
		// map to store data with nested map/slice
		nestedValues := make(map[string]interface{})
		// print all items without any nested data
		for _, key := range d.MapKeys() {
			value := d.MapIndex(key).Interface()
			if (reflect.ValueOf(value).Kind() == reflect.Map) || (reflect.ValueOf(value).Kind() == reflect.Slice) {
				nestedValues[key.Interface().(string)] = value
			}else {
				fmt.Printf("%s%s: %v\n", indent, key.Interface(), d.MapIndex(key).Interface())
			}
        }
		// print nested data using recursion
		for key, val := range nestedValues {
			fmt.Printf("%s%s: \n", indent, key)
			prettyPrint(val, indent+"  ")
		}
	case reflect.Slice:
		// if the data is in a slice itterate the slice and recursively print it
		for i := 0; i < d.Len(); i++ {
			prettyPrint(d.Index(i).Interface(), indent+"  ")
        }
	default:
		fmt.Printf("%s%v\n", indent, d)
	}
}

func main() {
	
	configPath := "cfg\\config.json"

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", configPath, err)
	}

	// validate if the json
	isJsonValid := json.Valid(data)

	if !isJsonValid {
		log.Fatalf("invalid JSON file  - %s", configPath)
	}

	// Parse into interface{}
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Print formatted output
	fmt.Println("=== JSON Configuration ===")
	prettyPrint(result, "  ")
}
