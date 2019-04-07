package dmv

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/GeorgeZhai/dmv-ca-poller/fileloader"
)

// OfflineDebugMode to enable the debugging without request DMV site
const OfflineDebugMode = bool(false)

// URLreferer hardcoded
const URLreferer = "https://www.dmv.ca.gov/wasapp/foa/findOfficeVisit.do"

// URLorigin hardcoded
const URLorigin = "https://www.dmv.ca.gov"

// URLPostOfficeVisit hardcoded
const URLPostOfficeVisit = "https://www.dmv.ca.gov/wasapp/foa/findOfficeVisit.do"

// URLPostDriveTest hardcoded
const URLPostDriveTest = "https://www.dmv.ca.gov/wasapp/foa/findDriveTest.do"

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

// ReqConfig contains configuration for RequestDMV
type ReqConfig struct {
	Mode               string
	DlNumber           string
	FirstName          string
	LastName           string
	BirthMonth         string
	BirthDay           string
	BirthYear          string
	TelArea            string
	TelPrefix          string
	TelSuffix          string
	CaptchaResponse    string
	GRecaptchaResponse string
}

// RequestDMV POST data to dmv webserver and get response string
func RequestDMV(d DMVinfo, c ReqConfig) (string, error) {

	url := URLPostOfficeVisit
	querydata := fmt.Sprintf("mode=OfficeVisit&captchaResponse=%s&officeId=%d&numberItems=1&taskCID=true&firstName=%s&lastName=%s&telArea=%s&telPrefix=%s&telSuffix=%s&resetCheckFields=true&g-recaptcha-response=%s", c.CaptchaResponse, d.ID, c.FirstName, c.LastName, c.TelArea, c.TelPrefix, c.TelSuffix, c.GRecaptchaResponse)
	if c.Mode == "DriveTest" {
		url = URLPostDriveTest
		querydata = fmt.Sprintf("mode=DriveTest&captchaResponse=%s&officeId=%d&numberItems=1&requestedTask=DT&firstName=%s&lastName=%s&dlNumber=%s&birthMonth=%s&birthDay=%s&birthYear=%s&telArea=%s&telPrefix=%s&telSuffix=%s&resetCheckFields=true&g-recaptcha-response=%s", c.CaptchaResponse, d.ID, c.FirstName, c.LastName, c.DlNumber, c.BirthMonth, c.BirthDay, c.BirthYear, c.TelArea, c.TelPrefix, c.TelSuffix, c.GRecaptchaResponse)
	}

	body := strings.NewReader(querydata)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("error:", err)
	}
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Origin", URLorigin)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Referer", URLreferer)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,zh-TW;q=0.6")

	if OfflineDebugMode {
		testdata, _ := fileloader.LoadFile("test/testdata.txt")
		return testdata, nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("error:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			defer reader.Close()
		default:
			reader = resp.Body
		}

		bodyBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Println("error:", err)
			return "", err
		}

		bodyString := string(bodyBytes)
		return bodyString, nil

	}

	return "", errors.New("POST status not ok: " + strconv.Itoa(resp.StatusCode))

}
