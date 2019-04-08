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
func GetAppointmentTime(s string, now time.Time) (time.Time, error) {

	oneYearLater := now.AddDate(1, 0, 0)

	if strings.Contains(s, "Sorry, all appointments at this office are currently taken") || strings.Contains(s, "no appointment is available") {
		return oneYearLater, nil
	}

	r, _ := regexp.Compile("(\\w+), (\\w+) (\\d{1,2}), (\\d{4}) at (\\d{1,2}:\\d{2}) (AM|PM)")
	match := r.FindStringSubmatch(s)

	if len(match) == 0 {
		_ = logReponse(s, "debugReponse.html")
		return oneYearLater, errors.New("No datetime string found in return")

	}
	t, err := time.Parse("Monday, January 2, 2006 at 15:04 PM -0700", match[0]+" -0700")
	if err != nil {
		log.Println("error:", err)
		return oneYearLater, err
	}

	return t, nil
}

func logReponse(s string, fn string) error {
	if f, errCreate := os.Create(fn); errCreate != nil {
		log.Println(errCreate)
		return errCreate
	} else if _, errWrite := f.WriteString(s); errWrite != nil {
		log.Println(errWrite)
		f.Close()
		return errWrite
	} else {
		f.Close()
		return nil
	}
}
