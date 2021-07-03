package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/turnage/graw/reddit"

	"gorm.io/gorm"
)



func main() {
	config := GetConfig()

	var db *gorm.DB
	db = OpenDatabase(config)

	bot, botErr := reddit.NewBotFromAgentFile("scrapper.agent", 0)
	if botErr != nil {
		panic("Failed to create bot handle")
	}

	StartObservations(config, &bot, db)

	Logf(1, "Starting server...\n\n")
	router := httprouter.New()
	router.GET("/start_observing/:subredditName", Handler{config, db, &bot}.StartObserving)
	router.GET("/stop_observing/:subredditName", Handler{config, db, &bot}.StopObserving)
	if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatal(err)
    }
}


