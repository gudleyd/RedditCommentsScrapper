package observer

import (
	"hse_practise/redditbot/mylog"
	"hse_practise/redditbot/models"
	"hse_practise/redditbot/dbmanager"
	"github.com/turnage/graw/reddit"
)

type CommentsObserver struct {
	Bot reddit.Bot
	SubredditName string
	CommentRepo *dbmanager.CommentRepo
}

func (c *CommentsObserver) Comment(comment *reddit.Comment) error {
	_, err := c.CommentRepo.Upsert(
		models.Comment{FullName: comment.Name, 
			Id: comment.ID, Author: 
			comment.Author, 
			Body: comment.Body, 
			Score: comment.Ups - comment.Downs, 
			SubredditID: comment.SubredditID})
	if err != nil {
		mylog.Logf(3, "Comment processing ended with error %s\n", err)
	} else {
		mylog.Logf(0, "Processed new comment on %s with (Id, ParentId) = (%s, %s)\n", 
		comment.Subreddit, 
		comment.ID, 
		comment.ParentID)
	}
	return nil
}
