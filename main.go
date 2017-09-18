package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// http://api.openweathermap.org/data/2.5/forecast?zip=22304&APPID=946b324f486a4ede799d2b98f38b1e34

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
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

	res, err := http.Get("http://api.openweathermap.org/data/2.5/forecast?zip=22304&APPID=946b324f486a4ede799d2b98f38b1e34")
	if err != nil {
		http.Error(w, "could not perform request", http.StatusInternalServerError)
		return
	}

	if res.StatusCode != http.StatusOK {
		http.Error(w, res.Status, res.StatusCode)
		return
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("could not read response body: %v", err)
		return
	}

	var results = new(weatherResponse)
	err = json.Unmarshal(b, &results)
	if err != nil {
		fmt.Printf("count not parse response body: %v", err)
		return
	}

	fmt.Print(results)
}
