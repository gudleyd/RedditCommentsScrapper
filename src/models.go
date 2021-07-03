package main

import (
	"gorm.io/driver/postgres"
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

func OpenDatabase(config Configuration) *gorm.DB {
	dsn := config.Dsn
	db, dbErr := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if dbErr != nil {
		panic("Failed to connect database")
	}
	db.AutoMigrate(&Comment{})
	db.AutoMigrate(&SubredditToObserve{})
	return db
}