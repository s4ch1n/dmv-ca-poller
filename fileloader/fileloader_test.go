package fileloader

import (
	"testing"
)

func TestLoadFile(t *testing.T) {

	r := JSONData{}

	// LoadJSONFile will load Jsondata from the filename
	err1 := LoadJSONFile(&r, "./dmvInfo.json")
	if len(r) != 174 {
		t.Errorf(" Returned DMV json object number wrong, got: %v, expecting: %v.", len(r), 174)
	}

	if err1 != nil {
		t.Errorf(" Returned error %s", err1)
	}

	r2, err2 := LoadFile("./dmvInfo.json")
	if len(r2) != 17417 {
		t.Errorf(" Returned DMV numbers wrong, got: %v, expecting: %v.", len(r2), 17417)
	}
	if err2 != nil {
		t.Errorf(" Returned error %s", err2)
	}

	_, err3 := LoadFile("./wrongfile.json")
	if err3 == nil {
		t.Errorf(" Expecting file not found error")
	}

}
