package main

import (
	"context"

	redditReadOnly "github.com/vartanbeno/go-reddit/v2/reddit"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"

	"gorm.io/gorm"
)

func StartObservations(config Configuration, bot *reddit.Bot, db *gorm.DB) {
	var subredditsToObserve []SubredditToObserve
	db.Find(&subredditsToObserve)
	for _, subredditToObserve := range subredditsToObserve {
		go StartSubredditObservation(subredditToObserve.Name, config, bot, db)
	}
}

func StartSubredditObservation(subredditName string, config Configuration, bot *reddit.Bot, db *gorm.DB) {

	SubredditCommentsPreload(subredditName, config, db)

	Logf(1, "Start observing %s\n", subredditName)
	cfg := graw.Config{SubredditComments: []string{subredditName}}
	handler := &CommentsObserver{bot: *bot, subredditName: subredditName, db: db}
	_, wait, observErr := graw.Run(handler, *bot, cfg)
	if observErr == nil {
		wait()
	} else {
		Logf(3, "Can not start observing %s\n", subredditName)
	}
}

func SubredditCommentsPreload(subredditName string, config Configuration, db *gorm.DB) {
	Logf(1, "Start preloading for %s\n\n", subredditName)
	readOnlyClient, clientErr := redditReadOnly.NewReadonlyClient()
	if clientErr != nil {
		Logf(3, "Failed to create read only client for %s with error %s\n", subredditName, clientErr)
	}
	var posts []*redditReadOnly.Post
	var err error
	processed, newOnes, errors := 0, 0, 0
	for config.MaxPostsPreload > 0 && errors < config.MaxErrors  {
		posts, _, err = readOnlyClient.Subreddit.NewPosts(context.Background(), subredditName, &redditReadOnly.ListOptions{
			Limit: int(min(100, config.MaxPostsPreload)),
		})
		if err != nil {
			Logf(2, "Preloading for subreddit %s failed with error %s. Skipping...\n", subredditName, err)
			errors += 1
		}
		for _, post := range posts {
			if config.MaxPostsPreload <= 0 || errors >= config.MaxErrors {
				break
			}
			var postErr error
			var postAndComments *redditReadOnly.PostAndComments
			postAndComments, _, err = readOnlyClient.Post.Get(context.Background(), post.ID)
			if postErr != nil {
				Logf(2, "Preloading for post %s failed with error %s. Skipping...\n", post.FullID, err)
				errors += 1
				continue
			}
			_, loadMoreErr := readOnlyClient.Post.LoadMoreComments(context.Background(), postAndComments)
			if loadMoreErr != nil {
				Logf(2, "Preloading for post %s failed with error %s. Skipping...\n", post.FullID, err)
				errors += 1
				continue
			}
			for _, comment := range postAndComments.Comments {
				var newComment Comment
				newComment = Comment{FullName: comment.FullID, Id: comment.ID, Author: comment.Author, Body: comment.Body, Score: int32(comment.Score), SubredditID: comment.SubredditID, ParentID: comment.ParentID}
				var count int64
				if db.Model(&newComment).Where("full_name = ?", newComment.FullName).Count(&count); count == 0 {
					Logf(0, "Processed new comment on %s with (Id, ParentId) = (%s, %s)\n", subredditName, comment.ID, comment.ParentID)
					db.Create(&newComment)
					newOnes += 1
				}
				processed += 1
			}
			config.MaxPostsPreload -= 1
		}
	}
	if errors >= config.MaxErrors {
		Logf(3, "Preloading for %s finished with %d errors. %d comments were processed. %d of them were new ones.\n", subredditName, errors, processed, newOnes)
	} else {
		Logf(1, "Preloading for %s finished. %d comments were processed. %d of them were new ones.\n", subredditName, processed, newOnes)
	}
}