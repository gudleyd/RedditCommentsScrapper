package handlers

import (
	"hse_practise/redditbot/config"
	"hse_practise/redditbot/models"
	"hse_practise/redditbot/dbmanager"
	"hse_practise/redditbot/mylog"
	"hse_practise/redditbot/logic"

	"net/http"
	"github.com/julienschmidt/httprouter"

	"github.com/turnage/graw/reddit"
)

type Handler struct {
	Config config.Configuration
	Bot *reddit.Bot
	CommentRepo *dbmanager.CommentRepo
}

func (h Handler) StartObserving(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	newSubreddit := models.SubredditToObserve{Name: ps.ByName("subredditName")}
	if isNew, err := h.CommentRepo.AddSubredditToObserve(newSubreddit); err == nil {
		if isNew {
			go logic.StartSubredditObservation(newSubreddit.Name, h.Config, h.Bot, h.CommentRepo)
			w.WriteHeader(http.StatusAccepted)
			_, writeErr := w.Write([]byte("202 - subreddit accepted for observing"))
			if writeErr != nil {
				mylog.Logf(2, "Error on write %s", writeErr)
			}
		} else {
			w.WriteHeader(http.StatusAlreadyReported)
			_, writeErr := w.Write([]byte("208 - this subreddit is already being observed"))
			if writeErr != nil {
				mylog.Logf(2, "Error on write %s", writeErr)
			}
		}
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write([]byte("500 - cannot add subreddit to observables"))
		if writeErr != nil {
			mylog.Logf(2, "Error on write %s", writeErr)
		}
		mylog.Logf(3, "Cannot add subreddit %s to observables", newSubreddit.Name)
	}
}

func (h Handler) StopObserving(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	subredditName := ps.ByName("subredditName")
	if wasObservable, err := h.CommentRepo.RemoveSubredditFromObservables(subredditName); err == nil {
		if wasObservable {
			mylog.Logf(2, "Stop observing %s", subredditName)
			w.WriteHeader(http.StatusOK)
			_, writeErr := w.Write([]byte("200 - this will have effect only after reload"))
			if writeErr != nil {
				mylog.Logf(2, "Error on write %s", writeErr)
			}
		} else {
			w.WriteHeader(http.StatusOK)
			_, writeErr := w.Write([]byte("202 - subreddit wasn't observed"))
			if writeErr != nil {
				mylog.Logf(2, "Error on write %s", writeErr)
			}
		}
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write([]byte("500 - cannot stop observing subreddit"))
		if writeErr != nil {
			mylog.Logf(2, "Error on write %s", writeErr)
		}
		mylog.Logf(3, "Cannot stop observing subreddit %s", subredditName)
	}
}
