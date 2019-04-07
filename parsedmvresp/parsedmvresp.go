package parsedmvresp

import (
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// GetAppointmentTime will parse DMV returned HTML content string to extract time. it returns time.Time, error
func GetAppointmentTime(s string) (time.Time, error) {
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
			log.Println(err)
		}
		l, err := f.WriteString(s)
		log.Println(l)
		if err != nil {
			log.Println(err)
			f.Close()
		}
		f.Close()
		return oneYearLater, errors.New("No datetime string found in return")

	}
	t, err := time.Parse("Monday, January 2, 2006 at 15:04 PM -0700", match[0]+" -0700")
	if err != nil {
		log.Println("error:", err)
		return oneYearLater, err
	}

	return t, nil
}
