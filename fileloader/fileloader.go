package fileloader

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// JSONData is a map[string] to hold any json object, expect root array object
type JSONData map[string]interface{}

// LoadJSONFile will load Jsondata from the filename
func LoadJSONFile(d *JSONData, f string) error {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bs, &d)
	if err != nil {
		return err
	}
	return nil
}

// LoadFile will load text file and return a string
func LoadFile(f string) (string, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		log.Println("error:", err)
		return "", err
	}
	return string(bs), nil
}
