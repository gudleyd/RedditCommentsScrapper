package models

import (
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	FullName string `gorm:"primaryKey"`
	Id string 
	Author string
	Body string
	Score int32
	SubredditID string
	ParentID string
}

type SubredditToObserve struct {
	gorm.Model
	Name string `gorm:"primaryKey"`
}
