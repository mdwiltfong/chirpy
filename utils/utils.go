package utils

import (
	"encoding/json"
	"errors"
	"github.com/mdwiltfong/chirpy/utils/types"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type DataBaseClient struct {
	Path string
	Mux  *sync.RWMutex
}

func NewDB(path string) (*DataBaseClient, error) {
	// Path towards database file
	_, err := os.ReadFile(path)
	if err != nil {
		templatePath := GetPath()
		dbTemplate, _ := os.ReadFile(templatePath)
		writeError := os.WriteFile(path, dbTemplate, 0644)
		if writeError != nil {
			log.Println(writeError)
			return nil, writeError
		}
	}
	return &DataBaseClient{Path: path, Mux: new(sync.RWMutex)}, nil

}

func (db *DataBaseClient) LoadDB() (types.Database, error) {
	dataBytes, err := os.ReadFile(db.Path)
	if err != nil {
		return types.Database{}, errors.New(err.Error())
	}
	tempStruct := types.Database{}
	unMarshalError := json.Unmarshal(dataBytes, &tempStruct)
	if unMarshalError != nil {
		return types.Database{}, errors.New(unMarshalError.Error())
	}
	return tempStruct, nil
}

func (db *DataBaseClient) WriteDB(dbStructure types.Database) error {
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
		dbTemplate, _ := os.ReadFile(GetPath())
		writeError := os.WriteFile(db.Path, dbTemplate, 0644)
		if writeError != nil {
			log.Println(writeError)
			return writeError
		}
	}
	return nil
}

func (db *DataBaseClient) GetChirps() ([]types.Chirp, error) {
	data, err := db.LoadDB()
	if err != nil {
		return nil, err
	}
	chirps := []types.Chirp{}
	for k := range data.Chirps {
		chirps = append(chirps, data.Chirps[k])
	}
	return chirps, nil
}

func (db *DataBaseClient) CreateChirp(body string) (types.Chirp, error) {
	dataStruct, _ := db.LoadDB()
	numOfChirps := len(dataStruct.Chirps)
	id := numOfChirps + 1
	newChirp := types.Chirp{ID: id, Body: body}
	dataStruct.Chirps[id] = newChirp
	err := db.WriteDB(dataStruct)
	if err != nil {
		return types.Chirp{}, err
	}
	return newChirp, nil
}

func (db *DataBaseClient) CreateUsers(email string, password []byte) (types.User, error) {
	dataStruct, _ := db.LoadDB()
	numOfUsers := len(dataStruct.Users)
	id := numOfUsers + 1
	newUser := types.User{ID: id, Email: email, Password: password}
	dataStruct.Users[id] = newUser
	err := db.WriteDB(dataStruct)
	if err != nil {
		return types.User{}, err
	}
	newUser.Password = nil
	return newUser, nil

}

func (db *DataBaseClient) GetUser(email string) (types.User, error) {
	dataStruct, _ := db.LoadDB()
	for _, user := range dataStruct.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return types.User{}, errors.New("Can't find user")
}

func GetPath() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	var relativePath string
	if strings.Contains(wd, "/tests") {
		relativePath = "../database/template.json"
	} else {
		relativePath = "./database/template.json"
	}

	return filepath.Join(wd, relativePath)
}
