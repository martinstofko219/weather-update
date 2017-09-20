package weatherupdate

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type weatherHistory struct {
	Main        string
	Description string
}

func (wh *weatherHistory) load(ctx context.Context) {
	key := datastore.NewKey(ctx, "WeatherHistory", weatherHistoryKey, 0, nil)
	if err := datastore.Get(ctx, key, wh); err != nil {
		log.Errorf(ctx, "could not load weather history from datastore: %v", err)
	}
}

func (wh *weatherHistory) update(ctx context.Context, results weatherResponse) {
	if wh == nil {
		log.Errorf(ctx, "invalid weather history")
		return
	}

	for _, r := range results.List {
		t := time.Unix(r.Dt, 0)
		if isWeddingDay(t) {
			wh.Main = r.Weather[0].Main
			wh.Description = r.Weather[0].Description
		}
	}

	key := datastore.NewKey(ctx, "WeatherHistory", weatherHistoryKey, 0, nil)
	if _, err := datastore.Put(ctx, key, wh); err != nil {
		log.Errorf(ctx, "could not update weather history in datastore: %v", err)
	}
}

func (wh *weatherHistory) changed(results weatherResponse) bool {
	if wh == nil {
		return false
	}

	for _, r := range results.List {
		t := time.Unix(r.Dt, 0)
		if isWeddingDay(t) {
			return r.Weather[0].Main != wh.Main || r.Weather[0].Description != wh.Description
		}
	}
	return false
}
