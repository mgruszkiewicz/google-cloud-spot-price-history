package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var PRICING_DIR = "data/"

func main() {
	files, err := os.ReadDir(PRICING_DIR)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
		file_open, _ := os.ReadFile(fmt.Sprintf("%s/%s", PRICING_DIR, file.Name()))
		file_content := string(file_open)

		var data map[string]interface{}

		err := yaml.Unmarshal([]byte(file_content), &data)
		if err != nil {
			log.Fatal(err)
		}
		timestamp, timestamp_ok := getTimestamp(data)

		if timestamp_ok {
			fmt.Println(timestamp)
		}

	}

}

func getTimestamp(data map[string]interface{}) (int, bool) {
	if about, ok := data["about"]; ok {
		if aboutMap, ok := about.(map[string]interface{}); ok {
			if timestamp, ok := aboutMap["timestamp"]; ok {
				return timestamp.(int), true
			}
		}
	}
	return 0, false
}
