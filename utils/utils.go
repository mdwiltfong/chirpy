package utils

import (
	"encoding/json"
	"errors"
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
		writeError := os.WriteFile(path, dbTemplate, 0644)
		if writeError != nil {
			log.Println(writeError)
			return nil, writeError
		}
	}
	return &DataBaseClient{Path: path, Mux: new(sync.RWMutex)}, nil

}

func (db *DataBaseClient) LoadDB() (DBStructure, error) {
	dataBytes, err := os.ReadFile(db.Path)
	if err != nil {
		return DBStructure{}, errors.New(err.Error())
	}
	tempStruct := DBStructure{}
	unMarshalError := json.Unmarshal(dataBytes, &tempStruct)
	if unMarshalError != nil {
		return DBStructure{}, errors.New(unMarshalError.Error())
	}
	return tempStruct, nil
}

func (db *DataBaseClient) WriteDB(dbStructure DBStructure) error {
	db.Mux.Lock()
	defer db.Mux.Unlock()
	dataBytes, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	writeError := os.WriteFile(db.Path, dataBytes, 0644)
	if writeError != nil {
		return writeError
	}

	return nil
}

func (db *DataBaseClient) EnsureDB() error {
	_, err := os.ReadFile(db.Path)
	if err != nil {
		dbTemplate, _ := os.ReadFile("../database/template.json")
		writeError := os.WriteFile(db.Path, dbTemplate, 0644)
		if writeError != nil {
			log.Println(writeError)
			return writeError
		}
	}
	return nil
}

func (db *DataBaseClient) GetChirps() ([]Chirp, error) {
	data, err := db.LoadDB()
	if err != nil {
		return nil, err
	}
	chirps := []Chirp{}
	for k := range data.Chirps {
		chirps = append(chirps, data.Chirps[k])
	}
	return chirps, nil
}
