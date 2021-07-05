package dbmanager

import (
	"hse_practise/redditbot/config"
	"hse_practise/redditbot/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenDatabase(config config.Configuration) (*gorm.DB, error) {
	db, dbErr := gorm.Open(postgres.Open(config.DatabaseConnectionString), &gorm.Config{})
	if dbErr != nil {
		return db, dbErr
	}
	commentsErr := db.AutoMigrate(&models.Comment{})
	if commentsErr != nil {
		return nil, commentsErr
	}
	subredditsErr := db.AutoMigrate(&models.SubredditToObserve{})
	if subredditsErr != nil {
		return nil, subredditsErr
	}
	return db, nil
}

type CommentRepo struct {
	Db *gorm.DB
}
 
func (c *CommentRepo) Upsert(newComment models.Comment) (bool, error) {
	var count int64
	if c.Db.Model(&newComment).Where("full_name = ?", newComment.FullName).Count(&count); count == 0 {
		err := c.Db.Create(&newComment).Error
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (c *CommentRepo) GetSubredditsToObserve() ([]models.SubredditToObserve, error) {
	var subredditsToObserve []models.SubredditToObserve
	err := c.Db.Find(&subredditsToObserve).Error
	if err != nil {
		return nil, err
	}
	return subredditsToObserve, nil
}

func (c *CommentRepo) AddSubredditToObserve(newSubredditToObserve models.SubredditToObserve) (bool, error) {
	var count int64
	if c.Db.Model(&newSubredditToObserve).Where("name = ?", newSubredditToObserve.Name).Count(&count); count == 0 {
		err := c.Db.Create(&newSubredditToObserve).Error
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (c *CommentRepo) RemoveSubredditFromObservables(subredditName string) (bool, error) {
	var subredditsToDelete []models.SubredditToObserve
	c.Db.Model(&subredditsToDelete).Where("name = ?", subredditName).Find(&subredditsToDelete)
	if len(subredditsToDelete) > 0 {
		err := c.Db.Delete(&subredditsToDelete[0]).Error
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
