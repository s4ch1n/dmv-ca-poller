package dmv

import (
	"testing"
)

func TestDMVSearch(t *testing.T) {

	l := Loc{
		Lat: 37.3785351,
		Lng: -122.0887737,
	}

	r, _ := GetQueryDMVs(l, 30)

	if len(r) != 11 {
		t.Errorf(" Returned DMV numbers wrong, got: %v, expecting: %v.", len(r), 11)
	}
}
