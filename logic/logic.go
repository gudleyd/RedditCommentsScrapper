package logic

import (
	"hse_practise/redditbot/config"
	"hse_practise/redditbot/models"
	"hse_practise/redditbot/mylog"
	"hse_practise/redditbot/dbmanager"
	"hse_practise/redditbot/observer"

	"context"

	redditReadOnly "github.com/vartanbeno/go-reddit/v2/reddit"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

func StartObservations(config config.Configuration, bot *reddit.Bot, commentRepo *dbmanager.CommentRepo) {
	subredditsToObserve, err := commentRepo.GetSubredditsToObserve()
	if err != nil {
		mylog.Logf(3, "Cannot load subreddits to observe with %s. Skipping...\n", err)
	}
	for _, subredditToObserve := range subredditsToObserve {
		go StartSubredditObservation(subredditToObserve.Name, config, bot, commentRepo)
	}
}

func StartSubredditObservation(subredditName string, config config.Configuration, bot *reddit.Bot, commentRepo *dbmanager.CommentRepo) {

	SubredditCommentsPreload(subredditName, config, commentRepo)

	mylog.Logf(1, "Start observing %s\n", subredditName)
	cfg := graw.Config{SubredditComments: []string{subredditName}}
	handler := &observer.CommentsObserver{Bot: *bot, SubredditName: subredditName, CommentRepo: commentRepo}
	_, wait, observErr := graw.Run(handler, *bot, cfg)
	if observErr != nil {
		mylog.Logf(3, "Can not start observing %s\n", subredditName)
	} else {
		err := wait()
		if err != nil {
			mylog.Logf(3, "Observing failed with %s\n", err)
		}
	}
}

func SubredditCommentsPreload(subredditName string, config config.Configuration, commentRepo *dbmanager.CommentRepo) {
	mylog.Logf(1, "Start preloading for %s\n\n", subredditName)
	readOnlyClient, clientErr := redditReadOnly.NewReadonlyClient()
	if clientErr != nil {
		mylog.Logf(3, "Failed to create read only client for %s with error %s\n", subredditName, clientErr)
	}
	var posts []*redditReadOnly.Post
	var err error
	processed, newOnes, errors := 0, 0, 0
	for config.MaxPostsPreload > 0 && errors < config.MaxErrors  {
		posts, _, err = readOnlyClient.Subreddit.NewPosts(context.Background(), subredditName, &redditReadOnly.ListOptions{
			Limit: int(min(100, config.MaxPostsPreload)),
		})
		if err != nil {
			mylog.Logf(2, "Preloading for subreddit %s failed with error %s. Skipping...\n", subredditName, err)
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
				mylog.Logf(2, "Preloading for post %s failed with error %s. Skipping...\n", post.FullID, err)
				errors += 1
				continue
			}
			_, loadMoreErr := readOnlyClient.Post.LoadMoreComments(context.Background(), postAndComments)
			if loadMoreErr != nil {
				mylog.Logf(2, "Preloading for post %s failed with error %s. Skipping...\n", post.FullID, err)
				errors += 1
				continue
			}
			for _, comment := range postAndComments.Comments {
				var isNew bool
				isNew, err = commentRepo.Upsert(models.Comment{FullName: comment.FullID, Id: comment.ID, Author: comment.Author, Body: comment.Body, Score: int32(comment.Score), SubredditID: comment.SubredditID, ParentID: comment.ParentID})
				if err != nil {
					mylog.Logf(2, "Error occured on comment saving")
				} else if isNew {
					mylog.Logf(0, "Processed new comment on %s with (Id, ParentId) = (%s, %s)\n", subredditName, comment.ID, comment.ParentID)
					newOnes += 1
				}
				processed += 1
			}
			config.MaxPostsPreload -= 1
		}
	}
	if errors >= config.MaxErrors {
		mylog.Logf(3, "Preloading for %s finished with %d errors. %d comments were processed. %d of them were new ones.\n", subredditName, errors, processed, newOnes)
	} else {
		mylog.Logf(1, "Preloading for %s finished. %d comments were processed. %d of them were new ones.\n", subredditName, processed, newOnes)
	}
}
