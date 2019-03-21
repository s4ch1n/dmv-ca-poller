package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type DefaultConf struct {
	URLPost            string
	CaptchaResponse    string
	GRecaptchaResponse string
}

type UserSettings struct {
	homeloclat       float64 `json:"homeloclat"`
	homeloclng       float64 `json:"homeloclng"`
	mode             string  `json:"mode"`
	numberItems      string  `json:"numberItems"`
	taskCID          string  `json:"taskCID"`
	firstName        string  `json:"firstName"`
	lastName         string  `json:"lastName"`
	telArea          string  `json:"telArea"`
	telPrefix        string  `json:"telPrefix"`
	telSuffix        string  `json:"telSuffix"`
	resetCheckFields string  `json:"resetCheckFields"`
}

type DMVInfo struct {
	name string  `json:"name"`
	id   int     `json:"id"`
	lat  float64 `json:"lat"`
	lng  float64 `json:"lng"`
}

var dc = DefaultConf{}

// var us = UserSettings{}

type jsondata interface{}

var us map[string]string

var dmv = DMVInfo{}

func loadDefaultConf(d *DefaultConf) {
	f, _ := os.Open("defaultconf.json")
	defer f.Close()
	decoder := json.NewDecoder(f)
	err := decoder.Decode(d)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func loadUserSettings(d *jsondata) {
	bs, err := ioutil.ReadFile("usersettings.json")
	if err != nil {
		fmt.Println("error:", err)
	}
	err = json.Unmarshal(bs, &d)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func loadDMVLoc(dmv *DMVInfo) {
	f, _ := os.Open("dmvInfo.json")
	defer f.Close()
	decoder := json.NewDecoder(f)
	err := decoder.Decode(dmv)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func requestDMV() (string, error) {

	body := strings.NewReader(`mode=OfficeVisit&captchaResponse=03AOLTBLR3efGSZ3wQ_Wtxsvu4TOyUpWNAtEzJ8BUWj450rgbx_M-277VsZKQWkQqZWH4lHTj_lGOckdsOTlZfwby63B_yPKUilY3IyC3zWDCZkUNYa8GvNiWAwlkTDrupNtDbN-aC88HmINrozTaKx1U-_rsB6A1B7ozohB_KeVJVod5YRbA8ElwRM79zINXoxgVKAYbumnez_IbbV1aXG5fm95uWwWz9JuP3LxfLmrobBC9K2A6GkcRywnhUZkZO5IGob_1D4CjtmLx37XBrt3JBH9ht5-oKLDR5crPXH2eahSwsO28WHodwF1z_qBs6kD4jRlgrpD0g&officeId=570&numberItems=1&taskCID=true&firstName=JASON&lastName=WHITE&telArea=408&telPrefix=122&telSuffix=4336&resetCheckFields=true&g-recaptcha-response=03AOLTBLR3efGSZ3wQ_Wtxsvu4TOyUpWNAtEzJ8BUWj450rgbx_M-277VsZKQWkQqZWH4lHTj_lGOckdsOTlZfwby63B_yPKUilY3IyC3zWDCZkUNYa8GvNiWAwlkTDrupNtDbN-aC88HmINrozTaKx1U-_rsB6A1B7ozohB_KeVJVod5YRbA8ElwRM79zINXoxgVKAYbumnez_IbbV1aXG5fm95uWwWz9JuP3LxfLmrobBC9K2A6GkcRywnhUZkZO5IGob_1D4CjtmLx37XBrt3JBH9ht5-oKLDR5crPXH2eahSwsO28WHodwF1z_qBs6kD4jRlgrpD0g`)
	req, err := http.NewRequest("POST", "https://www.dmv.ca.gov/wasapp/foa/findOfficeVisit.do", body)
	if err != nil {
		fmt.Println("error:", err)
	}
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Origin", "https://www.dmv.ca.gov")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Referer", "https://www.dmv.ca.gov/wasapp/foa/findOfficeVisit.do")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,zh-TW;q=0.6")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {

		// Check that the server actually sent compressed data
		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			fmt.Println("gzipped data found:")
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

	// r, _ := regexp.Compile(".*, .* \\d{1,2}, \\d{4} at \\d{1,2}:\\d{2} (AM|PM)")
	r, _ := regexp.Compile("(\\w+), (\\w+) (\\d{1,2}), (\\d{4}) at (\\d{1,2}:\\d{2}) (AM|PM)")

	match := r.FindStringSubmatch(s)

	if len(match) == 0 {
		return time.Now(), errors.New("No datetime string found in return")
	}

	fmt.Println(match[0])

	t, err := time.Parse("Monday, Jan 2, 2006 at 15:04 PM -0700", match[0]+" -0700")
	if err != nil {
		fmt.Println("error:", err)
		return time.Now(), err
	}

	return t, nil
}
func main() {

	loadDefaultConf(&dc)
	loadUserSettings(&us)
	// loadDMVLoc(&dmv)
	fmt.Println(dc)
	fmt.Println(us)
	fmt.Println(us["homeloclat"])

	// res, err := requestDMV()
	// if err != nil {
	// 	fmt.Println("error:", err)
	// 	os.Exit(1)
	// }

	res := testhtml
	// fmt.Println(res)

	t, err := getAppointmentTime(res)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println(t.Format("2006-01-02 15:04 PM MST"))

	delta := t.Sub(time.Now())
	fmt.Println(delta.Hours() / 24)

}

const testhtml = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">








<html lang="en">
<head>
<meta charset="windows-1252" />
<meta http-equiv="Content-Type" content="text/html; windows-1252" />
<meta http-equiv='Pragma' content='no-cache' />
<meta http-equiv='Cache-Control' content='no-cache' />
<meta http-equiv="X-UA-Compatible" content="IE=edge" />

<meta name="HandheldFriendly" content="True" />
<meta name="MobileOptimized" content="320" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, minimum-scale=1.0" />

<title>Appointment</title>

<link rel="stylesheet" href="/wasapp/../imageserver/theme/css/colorscheme-oceanside.css" />
<link rel="stylesheet" href="/wasapp/../imageserver/theme/css/cagov.core.css" />
<link rel="stylesheet" href="/wasapp/../imageserver/theme/css/wasapp.css" />
<script src="/wasapp/../imageserver/theme/js/jquery.min.js" type="text/javascript"></script>
<script src="/wasapp/../imageserver/theme/js/modernizr-2.0.6.min.js"></script>
<script src="/wasapp/../imageserver/theme/js/wasapp.js"></script>
<script type="text/javascript" src="javascript/JsFunctions.js"></script>
<script type="text/javascript" src="javascript/map.js"></script>
<script type="text/javascript" src="javascript/comm_map.js"></script>

<style type="text/css">
        @media print {
                .noPrint, #heading, #footer {
                    display: none !important;
                    height: 0;
                }
         }

        @media only screen {
                .noScreen,
                .hidden-screen {
                  display: none;
                }
        }

        @media print {
                .col-sm-1, .col-sm-2, .col-sm-3, .col-sm-4, .col-sm-5, .col-sm-6,.col-sm-7, .col-sm-8, .col-sm-9, .col-sm-10, .col-sm-11, .col-sm-12, .col-xs-12 {
                  float: left;
                }
                .col-sm-12, .col-xs-12 {
                  width: 100%;
                }
                .col-sm-11 {
                  width: 91.66666667%;
                }
                .col-sm-10 {
                  width: 83.33333333%;
                }
                .col-sm-9 {
                  width: 75%;
                }
                .col-sm-8 {
                  width: 66.66666667%;
                }
                .col-sm-7 {
                  width: 58.33333333%;
                }
                .col-sm-6 {
                  width: 66.66666667%;
                }
                .col-sm-5 {
                  width: 41.66666667%;
                }
                .col-sm-4 {
                  width: 33.33333333%;
                }
                .col-sm-3 {
                  width: 25%;
                }
                .col-sm-2 {
                  width: 25%;
                }
                .col-sm-1 {
                  width: 8.33333333%;
                }
                .col-sm-pull-12 {
                  right: 100%;
                }
                .col-sm-pull-11 {
                  right: 91.66666667%;
                }
                .col-sm-pull-10 {
                  right: 83.33333333%;
                }
                .col-sm-pull-9 {
                  right: 75%;
                }
                .col-sm-pull-8 {
                  right: 66.66666667%;
                }
                .col-sm-pull-7 {
                  right: 58.33333333%;
                }
                .col-sm-pull-6 {
                  right: 50%;
                }
                .col-sm-pull-5 {
                  right: 41.66666667%;
                }
                .col-sm-pull-4 {
                  right: 33.33333333%;
                }
                .col-sm-pull-3 {
                  right: 25%;
                }
                .col-sm-pull-2 {
                  right: 16.66666667%;
                }
                .col-sm-pull-1 {
                  right: 8.33333333%;
                }
                .col-sm-pull-0 {
                  right: auto;
                }
                .col-sm-push-12 {
                  left: 100%;
                }
                .col-sm-push-11 {
                  left: 91.66666667%;
                }
                .col-sm-push-10 {
                  left: 83.33333333%;
                }
                .col-sm-push-9 {
                  left: 75%;
                }
                .col-sm-push-8 {
                  left: 66.66666667%;
                }
                .col-sm-push-7 {
                  left: 58.33333333%;
                }
                .col-sm-push-6 {
                  left: 50%;
                }
                .col-sm-push-5 {
                  left: 41.66666667%;
                }
                .col-sm-push-4 {
                  left: 33.33333333%;
                }
                .col-sm-push-3 {
                  left: 25%;
                }
                .col-sm-push-2 {
                  left: 16.66666667%;
                }
                .col-sm-push-1 {
                  left: 8.33333333%;
                }
                .col-sm-push-0 {
                  left: auto;
                }
                .col-sm-offset-12 {
                  margin-left: 100%;
                }
                .col-sm-offset-11 {
                  margin-left: 91.66666667%;
                }
                .col-sm-offset-10 {
                  margin-left: 83.33333333%;
                }
                .col-sm-offset-9 {
                  margin-left: 75%;
                }
                .col-sm-offset-8 {
                  margin-left: 66.66666667%;
                }
                .col-sm-offset-7 {
                  margin-left: 58.33333333%;
                }
                .col-sm-offset-6 {
                  margin-left: 50%;
                }
                .col-sm-offset-5 {
                  margin-left: 41.66666667%;
                }
                .col-sm-offset-4 {
                  margin-left: 33.33333333%;
                }
                .col-sm-offset-3 {
                  margin-left: 25%;
                }
                .col-sm-offset-2 {
                  margin-left: 16.66666667%;
                }
                .col-sm-offset-1 {
                  margin-left: 8.33333333%;
                }
                .col-sm-offset-0 {
                  margin-left: 0%;
                }
        }

        @media print {
                .col-md-1, .col-md-2, .col-md-3, .col-md-4, .col-md-5, .col-md-6,.col-md-7, .col-md-8, .col-md-9, .col-md-10, .col-md-11, .col-md-12 {
                  float: left;
                }
                .col-md-12 {
                  width: 100%;
                }
                .col-md-11 {
                  width: 91.66666667%;
                }
                .col-md-10 {
                  width: 83.33333333%;
                }
                .col-md-9 {
                  width: 75%;
                }
                .col-md-8 {
                  width: 66.66666667%;
                }
                .col-md-7 {
                  width: 58.33333333%;
                }
                .col-md-6 {
                  width: 50%;
                }
                .col-md-5 {
                  width: 41.66666667%;
                }
                .col-md-4 {
                  width: 100%;
                }
                .col-md-3 {
                  width:35%;
                }
                .col-md-2 {
                  width: 16.66666667%;
                }
                .col-md-1 {
                  width: 8.33333333%;
                }
                .col-md-pull-12 {
                  right: 100%;
                }
                .col-md-pull-11 {
                  right: 91.66666667%;
                }
                .col-md-pull-10 {
                  right: 83.33333333%;
                }
                .col-md-pull-9 {
                  right: 75%;
                }
                .col-md-pull-8 {
                  right: 66.66666667%;
                }
                .col-md-pull-7 {
                  right: 58.33333333%;
                }
                .col-md-pull-6 {
                  right: 50%;
                }
                .col-md-pull-5 {
                  right: 41.66666667%;
                }
                .col-md-pull-4 {
                  right: 33.33333333%;
                }
                .col-md-pull-3 {
                  right: 25%;
                }
                .col-md-pull-2 {
                  right: 16.66666667%;
                }
                .col-md-pull-1 {
                  right: 8.33333333%;
                }
                .col-md-pull-0 {
                  right: auto;
                }
                .col-md-push-12 {
                  left: 100%;
                }
                .col-md-push-11 {
                  left: 91.66666667%;
                }
                .col-md-push-10 {
                  left: 83.33333333%;
                }
                .col-md-push-9 {
                  left: 75%;
                }
                .col-md-push-8 {
                  left: 66.66666667%;
                }
                .col-md-push-7 {
                  left: 58.33333333%;
                }
                .col-md-push-6 {
                  left: 50%;
                }
                .col-md-push-5 {
                  left: 41.66666667%;
                }
                .col-md-push-4 {
                  left: 33.33333333%;
                }
                .col-md-push-3 {
                  left: 25%;
                }
                .col-md-push-2 {
                  left: 16.66666667%;
                }
                .col-md-push-1 {
                  left: 8.33333333%;
                }
                .col-md-push-0 {
                  left: auto;
                }
                .col-md-offset-12 {
                  margin-left: 100%;
                }
                .col-md-offset-11 {
                  margin-left: 91.66666667%;
                }
                .col-md-offset-10 {
                  margin-left: 83.33333333%;
                }
                .col-md-offset-9 {
                  margin-left: 75%;
                }
                .col-md-offset-8 {
                  margin-left: 66.66666667%;
                }
                .col-md-offset-7 {
                  margin-left: 58.33333333%;
                }
                .col-md-offset-6 {
                  margin-left: 50%;
                }
                .col-md-offset-5 {
                  margin-left: 41.66666667%;
                }
                .col-md-offset-4 {
                  margin-left: 33.33333333%;
                }
                .col-md-offset-3 {
                  margin-left: 25%;
                }
                .col-md-offset-2 {
                  margin-left: 16.66666667%;
                }
                .col-md-offset-1 {
                  margin-left: 8.33333333%;
                }
                .col-md-offset-0 {
                  margin-left: 0%;
                }
        }

        @media print {
                  body, .body, form-control, .form-control {
                        font-size: 12px;
                    background-color: white;
                  }
    }

        @media print {
                h1, .h1, h2, .h2, h3, .h3 {
                        color: black;
                }
        }

        @media print {
                h1, .h1 {
                        font-size: 1.5em;
                }
        }

        @media print {
                input:disabled+label {
                        color: silver;
                }
        }

        @media print {
                input[ ?checkbox ?]:disabled+label {
                        color: silver;
                }
        }
}
</style>   

<style type="text/css">
input:disabled+label {
        color: silver;
}

input[ ?checkbox ?]:disabled+label {
        color: silver;
}
</style>



<script type="text/javascript">
        window.setTimeout("parent.location='/portal/dmv/timeout';",
                        605000);
</script>

<!--   Google Analytics   -->
<script type="text/javascript">
    var _gaq = _gaq || [];
    _gaq.push([ '_setAccount', 'UA-3419582-2' ]); // ca.gov  google analytics profile code
    _gaq.push([ '_setDomainName', '.ca.gov' ]);
    _gaq.push([ '_trackPageview' ]);

    (function() {
        var ga = document.createElement('script');
        ga.type = 'text/javascript';
        ga.async = true;
        ga.src = ('https:' == document.location.protocol ? 'https://ssl'
                                        : 'http://www')
                                        + '.google-analytics.com/ga.js';
        var s = document.getElementsByTagName('script')[0];
        s.parentNode.insertBefore(ga, s);
    })();
</script>
<script type="text/javascript">
    var _gaq = _gaq || [];
    _gaq.push([ '_setAccount', 'UA-3419582-34' ]); // DMV's google analytics profile code
    _gaq.push([ '_setDomainName', '.ca.gov' ]);
    _gaq.push([ '_trackPageview' ]);

    (function() {
        var ga = document.createElement('script');
        ga.type = 'text/javascript';
        ga.async = true;
        ga.src = ('https:' == document.location.protocol ? 'https://ssl'
                                        : 'http://www')
                                        + '.google-analytics.com/ga.js';
        var s = document.getElementsByTagName('script')[0];
        s.parentNode.insertBefore(ga, s);
    })();
</script>

<!--   Google Chrome Browser Autocomplete Disablement   -->
<script>
    $(document).ready(function () {
        try {
            $("input[type='text']").each(function(){
                $(this).attr("autocomplete","none");
               });
        }
        catch (e)
        { }
    });
</script>

</head>


<body class="app javascript_on ">
    

    <div id="heading">

        
            
            
                <header id="header" class="global-header">

    <div id="skip_to_content" style="visibility: visible"></div>

    <!-- logo and organization banner -->
    <div class="branding">
        <div class="header-cagov-logo">
            <img src="/wasapp/../imageserver/theme/images/header-ca.gov.png" alt="CA.gov" />
        </div>
    
        <div class="header-organization-banner">
            <img src="/wasapp/../imageserver/theme/images/header-organization.png" alt="California Department of Motor Vehicles">
        </div>
    </div>

    <div class="mobile-controls">
        <span class="mobile-control toggle-menu"><span class="ca-gov-icon-menu" aria-hidden="true"></span><span class="sr-only">Menu</span></span>
    </div>
    
            
        

        
            
            
                
            
        

        
            
                
                    <!-- UserInfo, logged out -->
                    <div id="UserInfo">
                        <p>
                            <a href="/wasapp/../portal/mydmv?lang=en">
                                Login
                            </a>
                            <a href="/wasapp/shoppingcart/shoppingCartApplication.do?localeName=en">
                                Shopping Cart
                                
                            </a>
                        </p>
                    </div>
                
                
            
        
        

        <!-- main navigation -->
        <nav id="navigation" class="main-navigation singlelevelnav auto-highlight mobile-closed">
        <ul id="nav_list" class="top-level-nav">
            <!-- DMV Home -->
            <li id="home-link" class="nav-item">
                <a href="/wasapp/../portal/dmv" class="first-level-link">
                    Home
                </a>
            </li>
        </ul>
        </nav>

        <div class="header-decoration"></div>
        </header>

    </div>
    <!-- closes heading div -->

    <a name="content"></a>

    <div id="main-content" class="main-content">
        <div id="app_header">
            
<h1>Appointment System</h1>
        </div>
        <div id="app_content">
            







<noscript>
        <p class="alert alert-danger">
                Javascript is off on your browser. This application requires Javascript to be enabled.
        </p>
</noscript>

<script>
        function submitForm_1() {
                document.getElementById("formId_1").submit(); 
        }
</script>
<p>
        You have selected to do the following task(s) for
        <strong>JASON WHITE</strong>:
</p>
<ul>


                <li>
                        Apply for, replace or renew a California driver license or identification card
                </li>


</ul>

<p>
        <strong>NOTE:</strong>
        This appointment is NOT for a behind-the-wheel driving test. If you want to schedule a driving test, click
        <a href="changeToDriveTest.do">
                here.
        </a>
</p>

<form id="formId_1" method="post" action="/wasapp/foa/checkForOfficeVisitConflicts.do">
        <!-- Selected Office -->
        <div class="panel panel-default">
                <div class="panel-heading">
                        <strong>You have selected the following office:</strong>
                </div>

                <div class="r-table col-xs-12">
                        <table class="col-sm-12 table-bordered table-condensed">
                                <thead>
                                        <tr>
                                                <th>Select</th>
                                                <th>Office</th>
                                                <th>Appointment</th>
                                        </tr>
                                </thead>
                                <tbody>
                                        <tr>
                                                <td data-title="Select">
                                                                <p class="radio full-label center-align">
                                                                        <label>
                                                                                <input type="radio" name="chosenOfficeId" value="0" checked="checked" id="office" />
                                                                        </label>
                                                                </p>
                                                        </td>
                                                <td data-title="Office">
                                                                <p class="no-margin-bottom">
                                                                        AUBURN
                                                                        <br />
                                                                        11722 Enterprise Drive
                                                                        <br />
                                                                        AUBURN, CA
                                                                </p>
                                                         </td>
                                                <td data-title="Appointment">
                                                         

                                                        <p class="no-margin-bottom">
                                                                The first available appointment for this office is on:


                                                        </p> 

                                                                        <p class="no-margin-bottom">
                                                                                <strong> Thursday, May 9, 2019 at 9:20 AM
                                                                                </strong>
                                                                        </p>

                                                        </td>
                                        </tr>
                                </tbody>
                        </table>
                </div>
        </div>

        <!-- Nearby Offices -->

</form>

<!-- Calendar -->
<form id="ApptForm" method="post" action="/wasapp/foa/findOfficeVisit.do" class="form-horizontal">
        <input type="hidden" id="showToday" value="true" />



        <fieldset>
                <p><em>Your selection must be for a time later than the first available appointment displayed above.</em></p>
                <div class="col-xs-12 col-md-3">
                        <label for="formattedRequestedDate">Date</label>
                        <br>
                        <input type="text" id="formattedRequestedDate" name="formattedRequestedDate" class="form-control" onchange="enableCheckAvail(this.name);">
                </div>
                <div class="col-xs-12 col-md-1">
                        and/or
                </div>
                <div class="col-xs-12 col-md-2">
                        <label for="requestedTime">Time</label>
                        <br>
                        <select name="requestedTime" id="requestedTime" class="form-control inline" onchange="enableCheckAvail(this.name);">
                                <option value="" selected> </option>
                                <option value ="0800">8:00 AM</option>
                                <option value ="0830">8:30 AM</option>
                                <option value ="0900">9:00 AM</option>
                                <option value ="0930">9:30 AM</option>
                                <option value ="1000">10:00 AM</option>
                                <option value ="1030">10:30 AM</option>
                                <option value ="1100">11:00 AM</option>
                                <option value ="1130">11:30 AM</option>
                                <option value ="1200">12:00 PM</option>
                                <option value ="1230">12:30 PM</option>
                                <option value ="1300">1:00 PM</option>
                                <option value ="1330">1:30 PM</option>
                                <option value ="1400">2:00 PM</option>
                                <option value ="1430">2:30 PM</option>
                                <option value ="1500">3:00 PM</option>
                                <option value ="1530">3:30 PM</option>
                                <option value ="1600">4:00 PM</option>
                                <option value ="1630">4:30 PM</option>
                        </select>
                </div>
                <div class="col-xs-12 col-md-6">
                        <br><input type="submit" id="checkAvail" name="checkAvail" class="btn btn-default" value="Check for Availability" /><br><br>
                </div>
                <br>
                <br>
                <p>NOTE:</p>
                <ul>
                        <li>A search by <i><b>date and time</b></i> will look for an appointment for that date only and the closest time on or after your request.</li>
                        <li>A search by <i><b>date only</b></i> will look for an appointment for that date only and the first available time on that date.</li>
                        <li>A search by <i><b>time only</b></i> will look for an appointment for the first available date with an appointment time on or after your request.</li>
                </ul>
        </fieldset>
        <script type="text/javascript">
            var ng_config = {
                assests_dir: 'calendar/'        // the path to the assets directory
            };
        </script>
        <script type="text/javascript" src="javascript/ng_all.js"></script>
        <script type="text/javascript" src="javascript/calendar.js"></script>
        <script type="text/javascript">
                var my_cal;
                var start = new Date('March 19, 2019');
                var today = document.getElementById("showToday");

                if (today != null) {
                        start.setDate(start.getDate() + 0);
                } else {
                        start.setDate(start.getDate() + 1);
                }

                var end = new Date('March 19, 2019');
                end.setDate(end.getDate() + 90);
                var holidays = [{date:1, month:3},{date:27, month:4}];

                ng.ready(function(){
                        document.getElementById("checkAvail").disabled = true;

                // creating the calendar
                // changed calendar to display 4 months for 90 days availability 12/2013
                my_cal = new ng.Calendar({
                    input: 'formattedRequestedDate',
                    num_months: 1,
                    num_col: 1,
                    dates_off: holidays,
                    weekend: [0,7],
                    date_format: 'D, d M Y',
                    server_date_format: 'D, d M Y',
                        start_date: start,
                        end_date: end,
                        events: { 
                                        onClose:function(dt){enableCheckAvail('formattedRequestedDate');}
                                },
                                offset:{x:0, y:0},
                                calendar_img: ng_config.assests_dir+'components/calendar/images/cal.gif'
                });
                
                // Format css for input field and icon
                $(".ng_cal_input_field").addClass("form-control").css({
                "border-top-right-radius": "0",
                "border-bottom-right-radius": "0"
            });
                $(".ng-button-icon-span img").css({
                        "padding": "8px 12px",
                        "border-top-right-radius": "4px",
                        "border-bottom-right-radius": "4px",
                        "border-left": "0",
                        "background-color": "#EEE",
                        "border-color": "#CCC"
                });
            });
        </script>
</form>

<div class="col-xs-12 form-group button-group xs-expand centered">


                                <a href="javascript: submitForm_1()" class="btn btn-primary">
                                        Continue
                                </a>
                                <a href="startOfficeVisit.do" class="btn btn-default">
                                        Back
                                </a>
                                <a href="../../portal/dmv/detail/portal/foa/welcome" class="btn btn-default">
                                        Cancel
                                </a>
            


            
            

</div>

<p class="foot_note">
        <em><strong>Note:</strong> If you choose <strong>&quot;Back&quot;</strong> or <strong>&quot;Cancel&quot;</strong> this appointment will NOT be submitted and the information previously entered will not be retained.</em>
</p>
        </div>
    </div>

    <div class="cleaner"></div>

    <div id="footer">
        
            
            
                <div class="footer-copyright global-footer">
    <div class="container">
        <div class="row">
        
            <div class="col-md-2">&nbsp;</div>
            
            <div class="col-md-4">

                <div class="foot1">
                    <a href="/wasapp/../portal/dmv/dmvfooter2/accessibility">Accessibility</a>
                    <br>
                    <a href="/wasapp/../portal/dmv/dmvfooter2/conditionsofuse">Conditions of Use</a>
                    <br>
                    <a href="/wasapp/../portal/dmv/dmvfooter1/disabilityservices">Disability Services</a>
                    <br>
                    <a href="/wasapp/../portal/dmv/dmvfooter1/footerhelp">Help</a>
                </div>
            </div>

            <div class="col-md-4">
                <div class="head2">
                    <a href="/wasapp/../portal/dmv">Home</a>
                    <br>
                    <a href="/wasapp/../portal/dmv/dmvfooter2/privacypolicy">Privacy Policy</a>
                    <br>
                    <a href="/wasapp/../portal/dmv/dmvfooter1/sitemap">Site Map</a>
                    <br>
                    <a href="/wasapp/../portal/dmv/dmvfooter1/technicalsupport">Technical Support</a>
                </div>
            </div>
            
            <div class="col-md-2">&nbsp;</div>

        </div>
    </div>
    <div class="row">
        <div class="copyrt">
            Copyright &copy; 
            <span id="copyright"> <script type="text/javascript">document.getElementById('copyright').appendChild(document.createTextNode(new Date().getFullYear()))
                        </script>
            </span> State of California
        </div>
    </div>
</div>
            
        
        <div
            style="padding: 0px; overflow: hidden; visibility: hidden; position: absolute; width: 1279px; left: -1280px; top: 0px; z-index: 1010;"
            id="WzTtDiV"></div>
    </div>

    <script src="/wasapp/../imageserver/theme/js/cagov.core.js"></script>

</body>
</html>`
