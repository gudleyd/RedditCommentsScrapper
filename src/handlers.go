package main

import (
	"net/http"
	"github.com/julienschmidt/httprouter"

	"github.com/turnage/graw/reddit"
	"gorm.io/gorm"
)

type Handler struct {
	config Configuration
	db *gorm.DB
	bot *reddit.Bot
}

func (h Handler) StartObserving(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	newSubredditToObserve := SubredditToObserve{Name: ps.ByName("subredditName")}
	var count int64
	if h.db.Model(&newSubredditToObserve).Where("name = ?", newSubredditToObserve.Name).Count(&count); count == 0 {
		go StartSubredditObservation(newSubredditToObserve.Name, h.config, h.bot, h.db)
		h.db.Create(&newSubredditToObserve)
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("202 - subreddit accepted for observing"))
	} else {
		w.WriteHeader(http.StatusAlreadyReported)
		w.Write([]byte("208 - this subreddit is already being observed"))
	}
}

func (h Handler) StopObserving(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var subredditsToDelete []SubredditToObserve
	h.db.Model(&subredditsToDelete).Where("name = ?", ps.ByName("subredditName")).Find(&subredditsToDelete)
	if len(subredditsToDelete) > 0 {
		h.db.Delete(&subredditsToDelete[0])
		Logf(2, "Stop observing %s", subredditsToDelete[0].Name)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 - this will have effect only after reload"))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("202 - subreddit wasn't observed"))
	}
}