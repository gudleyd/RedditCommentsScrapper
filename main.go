package main

import (
	"hse_practise/redditbot/config"
	"hse_practise/redditbot/mylog"
	"hse_practise/redditbot/dbmanager"
	"hse_practise/redditbot/handlers"
	"hse_practise/redditbot/logic"

	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/turnage/graw/reddit"

	"gorm.io/gorm"
)



func main() {
	config, configErr := config.GetConfig()
	if configErr != nil {
		mylog.Logf(3, "Can not parse config with %s\n", configErr)
		return
	}

	var db *gorm.DB
	var dbErr error
	db, dbErr = dbmanager.OpenDatabase(config)
	if dbErr != nil {
		mylog.Logf(3, "Cannot connect to database with %s\n", dbErr)
		return
	}
	commentRepo := dbmanager.CommentRepo{Db: db}
	
	bot, botErr := reddit.NewBot(
		reddit.BotConfig{Agent: config.BotUserAgent, 
			App: reddit.App{ID: config.ClientID, 
				Secret: config.ClientSecret, 
				Username: config.RedditUsername, 
				Password: config.RedditPassword}})
	if botErr != nil {
		mylog.Logf(3, "Cannot create bot with %s\n", dbErr)
		return
	}

	logic.StartObservations(config, &bot, &commentRepo)

	mylog.Logf(1, "Starting server...\n\n")
	handler := handlers.Handler{Config: config, Bot: &bot, CommentRepo: &commentRepo}
	router := httprouter.New()
	router.GET("/start_observing/:subredditName", handler.StartObserving)
	router.GET("/stop_observing/:subredditName", handler.StopObserving)
	if err := http.ListenAndServe(":8080", router); err != nil {
        mylog.Logf(3, "%s", err)
    }
}
