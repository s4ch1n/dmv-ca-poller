package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/GeorgeZhai/dmv-ca-poller/dmv"
	"github.com/GeorgeZhai/dmv-ca-poller/fileloader"
	"github.com/GeorgeZhai/dmv-ca-poller/notification"
	"github.com/GeorgeZhai/dmv-ca-poller/parsedmvresp"
)

func main() {
	// user settings
	us := fileloader.JSONData{}
	fileloader.LoadJSONFile(&us, "usersettings.json")
	homeloc := dmv.Loc{}
	mapstructure.Decode(us["homelocation"], &homeloc)

	// default configurations
	dc := fileloader.JSONData{}
	fileloader.LoadJSONFile(&dc, "defaultconf.json")

	searchDistance := flag.Int("distance", int(us["distance"].(float64)), "DMV distance to your honme")
	notifyForApptInDays := flag.Int("days", int(us["notifyForApptInDays"].(float64)), "How many days in the future?")
	pauseDMV := flag.Int("pauseDMVSec", 5, "minimum seconds to pause between requests")
	pauseRound := flag.Int("pauseRoundMin", 20, "gap between rounds in minutes")

	var queryErrorCountAllowed int
	flag.IntVar(&queryErrorCountAllowed, "errorAllowed", 20, "Number of query errors allow in the run")

	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	c := 0

	email := notification.NewEmailClient(us["senderEmail"].(string), us["senderPW"].(string), us["receiverEmail"].(string), us["smtpServerHost"].(string), int(us["smtpServerPort"].(float64)))
	reqClient := dmv.NewReqClient(us["mode_OfficeVisit_or_DriveTest"].(string), us["dlNumber"].(string), us["firstName"].(string), us["lastName"].(string), us["birthMonth"].(string), us["birthDay"].(string), us["birthYear"].(string), us["telArea"].(string), us["telPrefix"].(string), us["telSuffix"].(string), dc["CaptchaResponse"].(string), dc["GRecaptchaResponse"].(string))

	dmvstoquery := dmv.GetQueryDMVs(homeloc, float64(*searchDistance))

	for {
		c++
		log.Printf("Starting scan round %d", c)

		for _, d := range dmvstoquery {
			n := rand.Intn(10)
			log.Printf("working ... Query DMV %s (%d) %s in %d seconds, errorcount left %d , try to get appointment in %d days...", d.Name, d.ID, us["mode_OfficeVisit_or_DriveTest"].(string), n+(*pauseDMV), queryErrorCountAllowed, *notifyForApptInDays)
			time.Sleep(time.Duration(n+(*pauseDMV)) * time.Second)
			res, err := reqClient.RequestDMV(d)

			if err != nil {
				log.Println("error:", err)
				queryErrorCountAllowed--
				if queryErrorCountAllowed <= 0 {
					log.Fatal("quit programm, too many errors...:")
				}
				continue
			}

			t, err := parsedmvresp.GetAppointmentTime(res)

			if err != nil {
				log.Println("error:", err)
				queryErrorCountAllowed--
				if queryErrorCountAllowed <= 0 {
					log.Fatal("quit programm, too many errors...:")
				}
				continue
			}

			delta := t.Sub(time.Now())
			deltaDays := delta.Hours() / 24
			log.Printf(" in  %f days \n", deltaDays)

			if deltaDays <= float64(*notifyForApptInDays) {
				emailcontent := fmt.Sprintf("Found DMV %s, ID: %d, Date: %s, in the next %f days", d.Name, d.ID, t.Format("2006-01-02 15:04 PM MST"), deltaDays)
				// notification.SendEmail(emailconf, emailcontent)
				email.SendEmail(emailcontent)
			}

		}
		log.Printf("Completed scan round %d", c)
		time.Sleep(time.Duration(*pauseRound) * time.Minute)
	}

}
