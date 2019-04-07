package dmv

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// URLreferer hardcoded
const URLreferer = "https://www.dmv.ca.gov/wasapp/foa/findOfficeVisit.do"

// URLorigin hardcoded
const URLorigin = "https://www.dmv.ca.gov"

// URLPostOfficeVisit hardcoded
const URLPostOfficeVisit = "https://www.dmv.ca.gov/wasapp/foa/findOfficeVisit.do"

// URLPostDriveTest hardcoded
const URLPostDriveTest = "https://www.dmv.ca.gov/wasapp/foa/findDriveTest.do"

// ReqConfig contains configuration for RequestDMV
type ReqClient struct {
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

// NewReqClient create a new ReqClient
func NewReqClient(Mode string, DlNumber string, FirstName string, LastName string, BirthMonth string, BirthDay string, BirthYear string, TelArea string, TelPrefix string, TelSuffix string, CaptchaResponse string, GRecaptchaResponse string) *ReqClient {
	return &ReqClient{
		Mode:               Mode,
		DlNumber:           DlNumber,
		FirstName:          FirstName,
		LastName:           LastName,
		BirthMonth:         BirthMonth,
		BirthDay:           BirthDay,
		BirthYear:          BirthYear,
		TelArea:            TelArea,
		TelPrefix:          TelPrefix,
		TelSuffix:          TelSuffix,
		CaptchaResponse:    CaptchaResponse,
		GRecaptchaResponse: CaptchaResponse,
	}

}

// RequestDMV POST data to dmv webserver and get response string
func (c *ReqClient) RequestDMV(d DMVinfo) (string, error) {

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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
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

	return "", errors.New("psot status not ok: " + strconv.Itoa(resp.StatusCode))

}
