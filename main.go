package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

type jsondata map[string]interface{}

// DMVinfo defines ID, coordinates and NAme
type DMVinfo struct {
	ID   int
	Lat  float64
	Lng  float64
	Name string
}

// Loc defines coordinates
type Loc struct {
	Lat float64
	Lng float64
}

var us = jsondata{}
var dc = jsondata{}
var dmvs = jsondata{}
var dmvinfolists = []DMVinfo{}
var testhtml string
var offlineDebugMode = bool(false)
var queryErrorCountAllowed = int(20)

// outlook.com needs custermized stmp.Auth
type loginAuth struct {
	username, password string
}

// LoginAuth returns an Auth that implements the LOGIN authentication
// mechanism as defined in RFC 4616.
func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	command := string(fromServer)
	command = strings.TrimSpace(command)
	command = strings.TrimSuffix(command, ":")
	command = strings.ToLower(command)

	if more {
		if command == "username" {
			return []byte(fmt.Sprintf("%s", a.username)), nil
		} else if command == "password" {
			return []byte(fmt.Sprintf("%s", a.password)), nil
		} else {
			// We've already sent everything.
			return nil, fmt.Errorf("unexpected server challenge: %s", command)
		}
	}
	return nil, nil
}

func loadjsonfile(d *jsondata, f string) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println("error:", err)
	}
	err = json.Unmarshal(bs, &d)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func loadtextfile(f string) (string, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}
	return string(bs), nil
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

func findDMVsByDistance(dmvs []DMVinfo, l Loc, d float64) []DMVinfo {
	r := []DMVinfo{}

	for _, dmv := range dmvs {
		dist := distance(dmv.Lat, dmv.Lng, l.Lat, l.Lng)
		if dist <= d {
			r = append(r, dmv)
		}
	}
	return r
}

func notify(us jsondata, s string) {
	fmt.Println(s)

	senderEmail := us["senderEmail"].(string)
	senderPW := us["senderPW"].(string)
	receiverEmail := us["receiverEmail"].(string)
	smtpServerHost := us["smtpServerHost"].(string)
	smtpServerPort := int(us["smtpServerPort"].(float64))

	fmt.Println("sending email to ", receiverEmail)

	auth := LoginAuth(senderEmail, senderPW)

	to := []string{receiverEmail}
	msgs := fmt.Sprintf("To: %s\r\n"+
		"Subject: DMV search notification\r\n"+
		"\r\n"+
		" %s \r\n", receiverEmail, s)
	msg := []byte(msgs)
	err := smtp.SendMail(smtpServerHost+":"+strconv.Itoa(smtpServerPort), auth, senderEmail, to, msg)
	if err != nil {
		log.Fatal(err)
	}

}

func querydmvs(dmvs []DMVinfo, us jsondata, dc jsondata) {

	rand.Seed(time.Now().UnixNano())
	notifyForApptInDays := us["notifyForApptInDays"].(float64)

	for _, dmv := range dmvs {
		minPause := 5
		n := rand.Intn(10)
		fmt.Printf("working ... Query DMV %s (%d) in %d seconds, errorcount left %d , try to get appointment in %f days...", dmv.Name, dmv.ID, n+minPause, queryErrorCountAllowed, notifyForApptInDays)
		time.Sleep(time.Duration(n+minPause) * time.Second)
		res, err := requestDMV(dmv, us["mode"].(string), int(us["numberItems"].(float64)), us["taskCID"].(string), us["firstName"].(string), us["lastName"].(string), us["telArea"].(string), us["telPrefix"].(string), us["telSuffix"].(string), dc["URLPost"].(string), dc["CaptchaResponse"].(string), dc["GRecaptchaResponse"].(string))

		if err != nil {
			fmt.Println("error:", err)
			queryErrorCountAllowed--
			if queryErrorCountAllowed <= 0 {
				fmt.Println("quit programm, too many errors...:")
				os.Exit(1)
			}
			continue
		}

		t, err := getAppointmentTime(res)

		if err != nil {
			fmt.Println("error:", err)
			queryErrorCountAllowed--
			if queryErrorCountAllowed <= 0 {
				fmt.Println("quit programm, too many errors...:")
				os.Exit(1)
			}
			continue
		}

		delta := t.Sub(time.Now())
		deltaDays := delta.Hours() / 24
		fmt.Printf(" in  %f days \n", deltaDays)

		if deltaDays <= notifyForApptInDays {
			notification := fmt.Sprintf("Found DMV %s, ID: %d, Date: %s, in the next %f days", dmv.Name, dmv.ID, t.Format("2006-01-02 15:04 PM MST"), deltaDays)
			notify(us, notification)
		}

	}

}

func requestDMV(dmvs DMVinfo, mode string, numberItems int, taskCID string, firstName string, lastName string, telArea string, telPrefix string, telSuffix string, url string, captchaResponse string, gRecaptchaResponse string) (string, error) {

	referer := "https://www.dmv.ca.gov/wasapp/foa/findOfficeVisit.do"
	origin := "https://www.dmv.ca.gov"

	querydata := fmt.Sprintf("mode=%s&captchaResponse=%s&officeId=%d&numberItems=%d&taskCID=%s&firstName=%s&lastName=%s&telArea=%s&telPrefix=%s&telSuffix=%s&resetCheckFields=true&g-recaptcha-response=%s", mode, captchaResponse, dmvs.ID, numberItems, taskCID, firstName, lastName, telArea, telPrefix, telSuffix, gRecaptchaResponse)
	body := strings.NewReader(querydata)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println("error:", err)
	}
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Origin", origin)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,zh-TW;q=0.6")

	if offlineDebugMode {
		testhtml, _ = loadtextfile("testhtml.html")
		return testhtml, nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error:", err)
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
			fmt.Println("error:", err)
			return "", err
		}

		bodyString := string(bodyBytes)
		return bodyString, nil

	}

	return "", errors.New("POST status not ok: " + strconv.Itoa(resp.StatusCode))

}

func getAppointmentTime(s string) (time.Time, error) {
	now := time.Now()
	oneYearLater := now.Add(time.Hour * time.Duration(24*365))

	if strings.Contains(s, "Sorry, all appointments at this office are currently taken") || strings.Contains(s, "no appointment is available") {
		return oneYearLater, nil
	}

	r, _ := regexp.Compile("(\\w+), (\\w+) (\\d{1,2}), (\\d{4}) at (\\d{1,2}:\\d{2}) (AM|PM)")
	match := r.FindStringSubmatch(s)

	if len(match) == 0 {

		f, err := os.Create("notimestampe.html")
		if err != nil {
			fmt.Println(err)
		}
		l, err := f.WriteString(s)
		fmt.Println(l)
		if err != nil {
			fmt.Println(err)
			f.Close()
		}
		f.Close()
		return oneYearLater, errors.New("No datetime string found in return")

	}
	t, err := time.Parse("Monday, January 2, 2006 at 15:04 PM -0700", match[0]+" -0700")
	if err != nil {
		fmt.Println("error:", err)
		return oneYearLater, err
	}

	return t, nil
}
func main() {

	loadjsonfile(&dc, "defaultconf.json")
	loadjsonfile(&us, "usersettings.json")

	loadjsonfile(&dmvs, "dmvinfo.json")

	for k, v := range dmvs {
		dmv := DMVinfo{}
		mapstructure.Decode(v, &dmv)
		(&dmv).Name = k
		dmvinfolists = append(dmvinfolists, dmv)
	}

	homeloc := Loc{}
	mapstructure.Decode(us["homelocation"], &homeloc)

	dmvstoquery := findDMVsByDistance(dmvinfolists, homeloc, us["distance"].(float64))
	c := 0

	for {
		c++
		fmt.Printf("Starting scan round %d from: %s \n", c, time.Now().Format("2006-01-02 15:04 PM MST"))
		querydmvs(dmvstoquery, us, dc)
		fmt.Printf("Completed round %d from: %s \n", c, time.Now().Format("2006-01-02 15:04 PM MST"))

		time.Sleep(time.Duration(20) * time.Minute)
	}

}
