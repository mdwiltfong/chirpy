package utils

import (
	"log"
	"os"
	"sync"
)

type DataBaseClient struct {
	Path string
	Mux  *sync.RWMutex
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DataBaseClient, error) {
	// Path towards database file
	_, err := os.ReadFile(path)
	if err != nil {
		dbTemplate, _ := os.ReadFile("../database/template.json")
		writeError := os.WriteFile("../database/database.json", dbTemplate, 0644)
		if writeError != nil {
			log.Println(writeError)
			return nil, writeError
		}
	}
	return &DataBaseClient{Path: "../database/database.json"}, nil

}
