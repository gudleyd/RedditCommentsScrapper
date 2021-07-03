package main

import (
	"github.com/turnage/graw/reddit"

	"gorm.io/gorm"
)

type CommentsObserver struct {
	bot reddit.Bot
	subredditName string
	db *gorm.DB
}

func (c *CommentsObserver) Comment(comment *reddit.Comment) error {
	newComment := Comment{FullName: comment.Name, Id: comment.ID, Author: comment.Author, Body: comment.Body, Score: comment.Ups - comment.Downs, SubredditID: comment.SubredditID}
	var count int64
	if c.db.Model(&newComment).Where("full_name = ?", newComment.FullName).Count(&count); count == 0 {
		c.db.Create(&newComment)
	}
	Logf(0, "Processed new comment on %s with (Id, ParentId) = (%s, %s)\n", comment.Subreddit, comment.ID, comment.ParentID)
	return nil
}