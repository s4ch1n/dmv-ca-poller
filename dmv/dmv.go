package dmv

import (
	"math"

	"github.com/mitchellh/mapstructure"

	"github.com/GeorgeZhai/dmv-ca-poller/fileloader"
)

// DMVinfo defines ID, coordinates and NAme
type DMVinfo struct {
	ID   int
	Lat  float64
	Lng  float64
	Name string
}

// Loc defines struct to hold coordinates
type Loc struct {
	Lat float64
	Lng float64
}

func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64, unit ...string) float64 {
	const PI float64 = 3.141592653589793
	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)
	theta := float64(lng1 - lng2)
	radtheta := float64(PI * theta / 180)
	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515

	if len(unit) > 0 {
		if unit[0] == "K" {
			dist = dist * 1.609344
		} else if unit[0] == "N" {
			dist = dist * 0.8684
		}
	}

	return dist
}

// FindDMVsByDistance filter a list a DNVinfo with home location and distance range. it returns a new slice of DMVinfo
func FindDMVsByDistance(dmvs []DMVinfo, l Loc, d float64) []DMVinfo {
	r := []DMVinfo{}
	for _, dmv := range dmvs {
		dist := distance(dmv.Lat, dmv.Lng, l.Lat, l.Lng)
		if dist <= d {
			r = append(r, dmv)
		}
	}
	return r
}

// FindAllDMVs give all the DMVs in CA
func FindAllDMVs() []DMVinfo {
	r := []DMVinfo{}
	dmvs := fileloader.JSONData{}
	fileloader.LoadJSONFile(&dmvs, "./dmvinfo.json")

	for k, v := range dmvs {
		dmv := DMVinfo{}
		mapstructure.Decode(v, &dmv)
		(&dmv).Name = k
		r = append(r, dmv)
	}

	return r
}

// GetQueryDMVs return a slice of DMVs for query
func GetQueryDMVs(l Loc, d float64) []DMVinfo {
	return FindDMVsByDistance(FindAllDMVs(), l, d)
}
