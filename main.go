package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

const (
	weddingYear  = 2017
	weddingMonth = time.September
	weddingDay   = 23
	weddingHour  = 15
)

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/sendMessage", messageHandler)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)

	msg := url.Values{}
	msg.Set("To", "+18456496408")
	msg.Set("From", "+18457648204")
	msg.Set("Body", "Hello from Go!")

	b := strings.NewReader(msg.Encode())

	req, _ := http.NewRequest("POST", "https://api.twilio.com/2010-04-01/Accounts/"+os.Getenv("TWILIO_ACCOUNT_SID")+"/Messages", b)
	req.SetBasicAuth(os.Getenv("TWILIO_ACCOUNT_SID"), os.Getenv("TWILIO_AUTH_TOKEN"))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, resp.Status)
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
	client := urlfetch.Client(ctx)

	resp, err := client.Get("http://api.openweathermap.org/data/2.5/forecast?zip=22304&units=imperial&APPID=" + os.Getenv("WEATHER_API_KEY"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, resp.Status, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(w, "could not read response body: %v\n", err)
		return
	}

	var results = new(weatherResponse)
	err = json.Unmarshal(b, &results)
	if err != nil {
		fmt.Fprintf(w, "count not parse response body: %v\n", err)
		return
	}

	for _, r := range results.List {
		t := time.Unix(r.Dt, 0)

		if isWeddingDay(t) {
			for _, weather := range r.Weather {
				fmt.Fprintln(w, weather.Main, weather.Description, r.Main.Temp, r.Main.Humidity, t)
			}
		}
	}
}

func isWeddingDay(d time.Time) bool {
	return d.Year() == weddingYear && d.Month() == weddingMonth &&
		d.Day() == weddingDay && d.Hour() == weddingHour
}
