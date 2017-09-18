package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", handler)
}

type weatherResponse struct {
	List []struct {
		Dt      int `json:"dt"`
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
	resp, err := client.Get("http://api.openweathermap.org/data/2.5/forecast?zip=22304&APPID=946b324f486a4ede799d2b98f38b1e34")
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
		fmt.Fprintf(w, "could not read response body: %v", err)
		return
	}

	var results = new(weatherResponse)
	err = json.Unmarshal(b, &results)
	if err != nil {
		fmt.Fprintf(w, "count not parse response body: %v", err)
		return
	}

	fmt.Fprint(w, results)
}
