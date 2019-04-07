package main

import (
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

	rand.Seed(time.Now().UnixNano())
	notifyForApptInDays := us["notifyForApptInDays"].(float64)
	queryErrorCountAllowed := int(20)
	c := 0

	emailconf := notification.EmailConfig{
		SenderEmail:   us["senderEmail"].(string),
		SenderPW:      us["senderPW"].(string),
		ReceiverEmail: us["receiverEmail"].(string),
		ServerHost:    us["smtpServerHost"].(string),
		ServerPort:    int(us["smtpServerPort"].(float64)),
	}
	reqConf := dmv.ReqConfig{
		Mode:               us["mode_OfficeVisit_or_DriveTest"].(string),
		DlNumber:           us["dlNumber"].(string),
		FirstName:          us["firstName"].(string),
		LastName:           us["lastName"].(string),
		BirthMonth:         us["birthMonth"].(string),
		BirthDay:           us["birthDay"].(string),
		BirthYear:          us["birthYear"].(string),
		TelArea:            us["telArea"].(string),
		TelPrefix:          us["telPrefix"].(string),
		TelSuffix:          us["telSuffix"].(string),
		CaptchaResponse:    dc["CaptchaResponse"].(string),
		GRecaptchaResponse: dc["GRecaptchaResponse"].(string),
	}

	dmvstoquery := dmv.GetQueryDMVs(homeloc, us["distance"].(float64))

	for {
		c++
		log.Printf("Starting scan round %d", c)

		for _, d := range dmvstoquery {
			minPause := 5
			n := rand.Intn(10)
			log.Printf("working ... Query DMV %s (%d) %s in %d seconds, errorcount left %d , try to get appointment in %f days...", d.Name, d.ID, us["mode_OfficeVisit_or_DriveTest"].(string), n+minPause, queryErrorCountAllowed, notifyForApptInDays)
			time.Sleep(time.Duration(n+minPause) * time.Second)
			res, err := dmv.RequestDMV(d, reqConf)

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

			if deltaDays <= notifyForApptInDays {
				emailcontent := fmt.Sprintf("Found DMV %s, ID: %d, Date: %s, in the next %f days", d.Name, d.ID, t.Format("2006-01-02 15:04 PM MST"), deltaDays)
				notification.SendEmail(emailconf, emailcontent)
			}

		}
		log.Printf("Completed scan round %d", c)
		time.Sleep(time.Duration(20) * time.Minute)
	}

}
