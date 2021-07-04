package main

import (
	"os"
	"encoding/json"
)

type Configuration struct {
    MaxPostsPreload int `json:"maxPostsPreload"`
	Dsn string `json:"dsn"`
	MaxErrors int `json:"maxErrors"`
}

func GetConfig() Configuration {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Configuration{}
	decErr := decoder.Decode(&config)
	if decErr != nil {
		panic("Configuration file is not parseable")
	}
	return config
}