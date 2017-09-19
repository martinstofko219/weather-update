package weatherupdate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	weddingYear  = 2017
	weddingMonth = time.September
	weddingDay   = 23
	weddingHour  = 15

	twilioNumber = "+18457648204"
	passphrase   = "wedding weather"

	weatherHistoryKey = "history1"
)

var (
	weatherAPI       = "http://api.openweathermap.org/data/2.5/forecast?zip=22304&units=imperial&APPID=" + os.Getenv("WEATHER_API_KEY")
	twilioSendSMSAPI = "https://api.twilio.com/2010-04-01/Accounts/" + os.Getenv("TWILIO_ACCOUNT_SID") + "/Messages"
)

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/checkNow", checkNowHandler)
}

type weatherResponse struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp     float32 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	} `json:"list"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	results, err := fetchWeather(ctx, weatherAPI)
	if err != nil {
		http.Error(w, "could not fetch weather data", http.StatusInternalServerError)
		log.Errorf(ctx, "could not fetch weather data: %v", err)
		return
	}

	var history = new(weatherHistory)
	history.load(ctx)

	if history.changed(*results) {
		msgBody := messageBody(*results)
		history.update(ctx, *results)

		if err := sendMessage(ctx, msgBody, "+18456496408"); err != nil {
			http.Error(w, "could not send sms", http.StatusInternalServerError)
			log.Errorf(ctx, "could not send sms: %v", err)
			return
		}
	}
}

const smsAllowContent = `<?xml version="1.0" encoding="UTF-8"?>
	<Response>
		<Message>
			<Body>%s</Body>
		</Message>
	</Response>`

const smsDenyContent = `<?xml version="1.0" encoding="UTF-8"?>
	<Response>
		<Message>
			<Media>%s</Media>
		</Message>
	</Response>`

func checkNowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/xml")

	body := r.PostFormValue("Body")
	if !strings.EqualFold(body, passphrase) {
		media := "http://www.icge.co.uk/languagesciencesblog/wp-content/uploads/2014/04/you_shall_not_pass1.jpg"
		fmt.Fprintf(w, smsDenyContent, media)
		return
	}

	results, err := fetchWeather(ctx, weatherAPI)
	if err != nil {
		http.Error(w, "could not fetch weather data", http.StatusInternalServerError)
		log.Errorf(ctx, "could not fetch weather data: %v", err)
		return
	}

	reply := messageBody(*results)
	fmt.Fprintf(w, smsAllowContent, reply)
}

func fetchWeather(ctx context.Context, url string) (*weatherResponse, error) {
	resp, err := urlfetch.Client(ctx).Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get weather api data: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}

	var results = new(weatherResponse)

	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("count not decode weather response: %v", err)
	}
	return results, nil
}

func sendMessage(ctx context.Context, body string, toNumber string) error {
	vals := url.Values{}
	vals.Set("To", toNumber)
	vals.Set("From", twilioNumber)
	vals.Set("Body", body)

	mb := strings.NewReader(vals.Encode())
	req, err := http.NewRequest("POST", twilioSendSMSAPI, mb)
	if err != nil {
		return fmt.Errorf("could not create sms request: %v", err)
	}
	req.SetBasicAuth(os.Getenv("TWILIO_ACCOUNT_SID"), os.Getenv("TWILIO_AUTH_TOKEN"))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err = urlfetch.Client(ctx).Do(req)
	if err != nil {
		return fmt.Errorf("could not send sms request: %v", err)
	}
	return nil
}

func messageBody(results weatherResponse) string {
	msg := "Wedding Weather Report\n"
	for _, r := range results.List {
		t := time.Unix(r.Dt, 0)
		if isWeddingDay(t) {
			msg += fmt.Sprintf("Temp: %d degrees\nHumidity: %d%%\n%s: %s",
				int(r.Main.Temp), r.Main.Humidity, r.Weather[0].Main, r.Weather[0].Description)
		}
	}
	return msg
}

func isWeddingDay(d time.Time) bool {
	return d.Year() == weddingYear && d.Month() == weddingMonth &&
		d.Day() == weddingDay && d.Hour() == weddingHour
}
