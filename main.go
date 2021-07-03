package main

import (
	"fmt"
	"os"
	"log"
	"net/http"
	"context"

	"encoding/json"

	"github.com/julienschmidt/httprouter"

	redditReadOnly "github.com/vartanbeno/go-reddit/v2/reddit"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	InfoColor    = "\033[32m"
	ColorReset   = "\033[0m"
)

type Configuration struct {
    MaxPostsPreload int `json:"maxPostsPreload"`
	Dsn string `json:"dsn"`
}

type Comment struct {
	gorm.Model
	FullName string `gorm:"primaryKey"`
	Id string 
	Author string
	Body string
	Score int32
	SubredditID string
}

type SubredditToObserve struct {
	gorm.Model
	Name string `gorm:"primaryKey"`
}

type commentsObserver struct {
	bot reddit.Bot
	subredditName string
	db *gorm.DB
}

func (c *commentsObserver) Comment(comment *reddit.Comment) error {
	newComment := Comment{FullName: comment.Name, Id: comment.ID, Author: comment.Author, Body: comment.Body, Score: comment.Ups - comment.Downs, SubredditID: comment.SubredditID}
	var count int64
	if c.db.Model(&newComment).Where("full_name = ?", newComment.FullName).Count(&count); count == 0 {
		c.db.Create(&newComment)
	}
	fmt.Printf("Processed new comment on %s\n", comment.Subreddit)
	return nil
}

func main() {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Configuration{}
	decErr := decoder.Decode(&config)
	if decErr != nil {
		panic("Configuration file is not parseable")
	}

	dsn := config.Dsn
	db, dbErr := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if dbErr != nil {
		panic("Failed to connect database")
	}
	db.AutoMigrate(&Comment{})
	db.AutoMigrate(&SubredditToObserve{})

	bot, botErr := reddit.NewBotFromAgentFile("scrapper.agent", 0)
	if botErr != nil {
		panic("Failed to create bot handle")
	}

	var subredditsToObserve []SubredditToObserve
	db.Find(&subredditsToObserve)
	for _, subredditToObserve := range subredditsToObserve {
		go StartBoard(subredditToObserve.Name, config.MaxPostsPreload, &bot, db)
	}

	fmt.Printf("%sStarting server...\n\n%s", InfoColor, ColorReset)
	router := httprouter.New()
	router.GET("/start_observing/:subredditName", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		newSubredditToObserve := SubredditToObserve{Name: ps.ByName("subredditName")}
		var count int64
		if db.Model(&newSubredditToObserve).Where("name = ?", newSubredditToObserve.Name).Count(&count); count == 0 {
			go StartBoard(newSubredditToObserve.Name, config.MaxPostsPreload, &bot, db)
			db.Create(&newSubredditToObserve)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - subreddit accepted for observing"))
		} else {
			w.WriteHeader(http.StatusAlreadyReported)
			w.Write([]byte("208 - this subreddit is already being observed"))
		}
    })
	router.GET("/end_observing/:subredditName", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var subredditsToDelete []SubredditToObserve
		db.Model(&subredditsToDelete).Where("name = ?", ps.ByName("subredditName")).Find(&subredditsToDelete)
		if len(subredditsToDelete) > 0 {
			db.Delete(&subredditsToDelete[0])
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("200 - this will have effect only after reload"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("202 - subreddit wasn't observed"))
		}
    })

	if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatal(err)
    }
}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

func StartBoard(subredditName string, maxPostsPreload int, bot *reddit.Bot, db *gorm.DB) {
	fmt.Printf("%sStarting preloading for %s\n\n%s", InfoColor, subredditName, ColorReset)
	readOnlyClient, _ := redditReadOnly.NewReadonlyClient()
	var posts []*redditReadOnly.Post
	var err error
	processed := 0
	newOnes := 0
	for maxPostsPreload > 0  {
		posts, _, err = readOnlyClient.Subreddit.NewPosts(context.Background(), subredditName, &redditReadOnly.ListOptions{
			Limit: int(min(100, maxPostsPreload)),
		})
		if err != nil {
			fmt.Printf("Preloading for subreddit %s failed with error %s. Skipping...\n", subredditName, err)
		}
		for _, post := range posts {
			if maxPostsPreload <= 0 {
				break
			}
			var postErr error
			var postAndComments *redditReadOnly.PostAndComments
			postAndComments, _, err = readOnlyClient.Post.Get(context.Background(), post.ID)
			if postErr != nil {
				fmt.Printf("Preloading for post %s failed with error %s. Skipping...\n", post.FullID, err)
				continue
			}
			_, loadMoreErr := readOnlyClient.Post.LoadMoreComments(context.Background(), postAndComments)
			if loadMoreErr != nil {
				fmt.Printf("Preloading for post %s failed with error %s. Skipping...\n", post.FullID, err)
				continue
			}
			for _, comment := range postAndComments.Comments {
				var newComment Comment
				newComment = Comment{FullName: comment.FullID, Id: comment.ID, Author: comment.Author, Body: comment.Body, Score: int32(comment.Score), SubredditID: comment.SubredditID}
				var count int64
				if db.Model(&newComment).Where("full_name = ?", newComment.FullName).Count(&count); count == 0 {
					db.Create(&newComment)
					newOnes += 1
				}
				processed += 1
			}
			maxPostsPreload -= 1
		}
	}
	fmt.Printf("%sPreloading for %s ended. %d comments were processed. %d of them were new ones.\nStarting observing %s...\n\n%s", InfoColor, subredditName, processed, newOnes, subredditName, ColorReset)

	cfg := graw.Config{SubredditComments: []string{subredditName}}
	handler := &commentsObserver{bot: *bot, subredditName: subredditName, db: db}
	_, wait, _ := graw.Run(handler, *bot, cfg)
	wait()
}
