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

var (
	searchDistance         = flag.Int("distance", 0, "DMV distance to your honme")
	notifyForApptInDays    = flag.Int("days", 0, "How many days in the future?")
	queryErrorCountAllowed = flag.Int("error_allowed", 20, "Number of query errors allow in the run")
	pauseDMV               = flag.Duration("pause_dmv", time.Duration(12*time.Second), "minimum seconds to pause between requests")
	pauseRound             = flag.Duration("pause_round", time.Duration(25*time.Minute), "gap between rounds in minutes")
)

func main() {

	flag.Parse()

	userSettings := fileloader.JSONData{}
	if err := fileloader.LoadJSONFile(&userSettings, "usersettings.json"); err != nil {
		log.Fatal(err)
	}
	homeloc := dmv.Loc{}
	if err := mapstructure.Decode(userSettings["homelocation"], &homeloc); err != nil {
		log.Fatal(err)
	}

	usDistance := int(userSettings["distance"].(float64))
	usNotifyForApptInDays := int(userSettings["notifyForApptInDays"].(float64))
	usMode := userSettings["mode_OfficeVisit_or_DriveTest"].(string)
	usDlNumber := userSettings["dlNumber"].(string)
	usFirstName := userSettings["firstName"].(string)
	usLastName := userSettings["lastName"].(string)
	usBirthMonth := userSettings["birthMonth"].(string)
	usBirthDay := userSettings["birthDay"].(string)
	usBirthYear := userSettings["birthYear"].(string)
	usTelArea := userSettings["telArea"].(string)
	usTelPrefix := userSettings["telPrefix"].(string)
	ustelSuffix := userSettings["telSuffix"].(string)
	usSenderEmail := userSettings["senderEmail"].(string)
	usSenderPW := userSettings["senderPW"].(string)
	usReceiverEmail := userSettings["receiverEmail"].(string)
	usSMTPServerHost := userSettings["smtpServerHost"].(string)
	usSMTPerverPort := int(userSettings["smtpServerPort"].(float64))

	if *searchDistance == 0 {
		*searchDistance = usDistance
	}
	if *notifyForApptInDays == 0 {
		*notifyForApptInDays = usNotifyForApptInDays
	}

	defaultConf := fileloader.JSONData{}
	if err := fileloader.LoadJSONFile(&defaultConf, "defaultconf.json"); err != nil {
		log.Fatal(err)
	}

	dcCaptchaResponse := defaultConf["CaptchaResponse"].(string)
	dcGRecaptchaResponse := defaultConf["GRecaptchaResponse"].(string)

	rand.Seed(time.Now().UnixNano())
	c := 0

	email := notification.NewEmailClient(usSenderEmail, usSenderPW, usReceiverEmail, usSMTPServerHost, usSMTPerverPort)
	reqClient := dmv.NewReqClient(usMode, usDlNumber, usFirstName, usLastName, usBirthMonth, usBirthDay, usBirthYear, usTelArea, usTelPrefix, ustelSuffix, dcCaptchaResponse, dcGRecaptchaResponse)

	dmvstoquery, errDmv := dmv.GetQueryDMVs(homeloc, float64(*searchDistance))
	if errDmv != nil {
		log.Fatal(errDmv)
	}

	for {
		c++
		log.Printf("Starting scan round %d", c)

		for _, d := range dmvstoquery {
			n := rand.Intn(10)
			log.Printf("working ... Query DMV %s (%d) %s in %v, errorcount left %d , try to get appointment in %d days...", d.Name, d.ID, userSettings["mode_OfficeVisit_or_DriveTest"].(string), (*pauseDMV + time.Duration(n)*time.Second), *queryErrorCountAllowed, *notifyForApptInDays)
			time.Sleep(*pauseDMV + time.Duration(n)*time.Second)
			res, err := reqClient.RequestDMV(d)

			if err != nil {
				log.Println("error:", err)
				*queryErrorCountAllowed--
				if *queryErrorCountAllowed <= 0 {
					log.Fatal("quit programm, too many errors...:")
				}
				continue
			}

			t, err := parsedmvresp.GetAppointmentTime(res, time.Now())

			if err != nil {
				log.Println("error:", err)
				*queryErrorCountAllowed--
				if *queryErrorCountAllowed <= 0 {
					log.Fatal("quit programm, too many errors...:")
				}
				continue
			}

			deltaDays := t.Sub(time.Now()).Hours() / 24
			log.Printf(" in  %f days \n", deltaDays)

			if deltaDays <= float64(*notifyForApptInDays) {

				if err := email.SendEmail(fmt.Sprintf("Found DMV %s, ID: %d, Date: %s, in the next %f days", d.Name, d.ID, t.Format("2006-01-02 15:04 PM MST"), deltaDays)); err != nil {
					log.Fatal(err)
				}
			}

		}
		log.Printf("Completed scan round %d", c)
		time.Sleep(*pauseRound)
	}

}
