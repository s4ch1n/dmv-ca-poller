package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
)

type ActionURL struct {
	Action string `json:"action"`
	HERF   string `json:"herf"`
	URL    string `json:"url"`
}

type Settings struct {
	domain    string
	baseurl   string
	mode      string
	firstName string
	lastName  string
	telArea   string
	telPrefix string
	telSuffix string
}

func loadSettings(s *Settings) {
	s.domain = "www.dmv.ca.gov"
	s.baseurl = "/portal/dmv/dmv/appointments"
	s.mode = "Office Visit Appointment" //"Office Visit Appointment" or "Behind-the-Wheel Drive Test"
	s.firstName = "John"
	s.lastName = "Doe"
	s.telArea = "408"
	s.telPrefix = "456"
	s.telSuffix = "7890"

}

func main() {
	var s Settings
	loadSettings(&s)
	actionurls, err := getActionURLs(&s)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, actionurl := range actionurls {
		if s.mode == "Office Visit Appointment" && s.mode == actionurl.Action {
			url := actionurl.URL
			iddmv := "516"
			t, err := getOfficeVisitAppointmentTime(&s, iddmv, url)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(fmt.Sprintf("appointment time for DMV id: %s is: %s", iddmv, t))
		} else if s.mode == "Behind-the-Wheel Drive Test" && s.mode == actionurl.Action {
			url := actionurl.URL
			iddmv := "516"
			t, err := getDriveTestAppointmentTime(&s, iddmv, url)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(fmt.Sprintf("appointment time for DMV id: %s is: %s", iddmv, t))
		}

	}

}

func getOfficeVisitAppointmentTime(s *Settings, iddmv string, url string) (string, error) {
	fmt.Println("using ferret to simulate web clicking may not work with the captcha!!!")

	err := errors.New("function getOfficeVisitAppointmentTime not implimented")
	return "", err
}

func getDriveTestAppointmentTime(s *Settings, iddmv string, url string) (string, error) {
	fmt.Println("using ferret to simulate web clicking may not work with the captcha!!!")

	err := errors.New("function getDriveTestAppointmentTime not implimented")
	return "", err
}

func getActionURLs(s *Settings) ([]*ActionURL, error) {

	query := fmt.Sprintf(`
	LET DOMAIN = '%s'
	LET BASEURL = '%s'
	LET doc = DOCUMENT('https://'+DOMAIN+BASEURL, true)
	WAIT_ELEMENT(doc, '.btn3', 5000)
	LET app_btns = ELEMENTS(doc, '.btn3')
	FOR btn IN app_btns
		FILTER TRIM(btn.innerText) == "Office Visit Appointment" || TRIM(btn.innerText) == "Behind-the-Wheel Drive Test"
		RETURN {action: TRIM(btn.innerText),herf: btn.attributes.href,url: 'https://'+DOMAIN+BASEURL+btn.attributes.href}
	`, s.domain, s.baseurl)
	// fmt.Printf("%s", query)

	comp := compiler.New()

	program, err := comp.Compile(query)

	if err != nil {
		return nil, err
	}

	// create a root context
	ctx := context.Background()

	// enable HTML drivers
	// by default, Ferret Runtime does not know about any HTML drivers
	// all HTML manipulations are done via functions from standard library
	// that assume that at least one driver is available
	ctx = drivers.WithContext(ctx, cdp.NewDriver())
	// ctx = drivers.WithContext(ctx, http.NewDriver(), drivers.AsDefault())

	out, err := program.Run(ctx)

	if err != nil {
		return nil, err
	}

	res := make([]*ActionURL, 0, 2)

	err = json.Unmarshal(out, &res)

	if err != nil {
		return nil, err
	}

	return res, nil

}
