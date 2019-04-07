package fileloader

import (
	"testing"
)

func TestLoadFile(t *testing.T) {

	r := JSONData{}

	// LoadJSONFile will load Jsondata from the filename
	LoadJSONFile(&r, "./dmvInfo.json")
	if len(r) != 174 {
		t.Errorf(" Returned DMV json object number wrong, got: %v, expecting: %v.", len(r), 174)
	}

	r2, _ := LoadFile("./dmvInfo.json")
	if len(r2) != 17417 {
		t.Errorf(" Returned DMV numbers wrong, got: %v, expecting: %v.", len(r2), 17417)
	}
}
