package utils

import (
	"crypto/rand"
	"encoding/hex"
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

func (db *DataBaseClient) GetUserByEmail(email string) (types.User, error) {
	dataStruct, _ := db.LoadDB()
	for _, user := range dataStruct.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return types.User{}, errors.New("Can't find user")
}

func (db *DataBaseClient) GetUserByID(id int) (types.User, error) {
	datastruct, _ := db.LoadDB()
	return datastruct.Users[id], nil
}

func (db *DataBaseClient) UpdateUser(id int, updateInformation types.User) (types.User, error) {
	dataStruct, err := db.LoadDB()
	if err != nil {
		return types.User{}, err
	}
	dataStruct.Users[id] = updateInformation

	return dataStruct.Users[id], db.WriteDB(dataStruct)

}
func (db *DataBaseClient) StoreUsersRefreshToken(token string, userId int) (types.RefreshToken, error) {
	dataStruct, err := db.LoadDB()
	if err != nil {
		log.Print(err.Error())
		return types.RefreshToken{}, err
	}
	numOfTokens := len(dataStruct.RefreshTokens)
	id := numOfTokens + 1
	refreshToken := types.RefreshToken{ID: id, Token: token, UserId: userId, IsValid: true}
	dataStruct.RefreshTokens[id] = refreshToken
	user := dataStruct.Users[userId]
	user.RefreshTokenId = id
	dataStruct.Users[userId] = user
	writeErr := db.WriteDB(dataStruct)
	if writeErr != nil {
		return types.RefreshToken{}, writeErr
	}
	return refreshToken, nil
}

func (db *DataBaseClient) GetRefreshToken(tokenId int) (types.RefreshToken, error) {
	dataStruct, err := db.LoadDB()
	if err != nil {
		return types.RefreshToken{}, errors.New("There was an issue loading the db")
	}
	if token, ok := dataStruct.RefreshTokens[tokenId]; ok == false {
		return types.RefreshToken{}, errors.New("Token not found")
	} else {
		return token, nil
	}
}

func (db *DataBaseClient) InvalidateToken(tokenId int) (types.RefreshToken, error) {
	token, err := db.GetRefreshToken(tokenId)
	if err != nil {
		return types.RefreshToken{}, err
	}
	if token.IsValid == false {
		return token, nil
	} else {
		token.IsValid = false
	}

	return token, nil
}

func (db *DataBaseClient) InvalidateUsersToken(userId int) error {
	foundUser, err := db.GetUserByID(userId)
	if err != nil {
		log.Print(err.Error())
		return errors.New("User not found")
	}
	token, getErr := db.GetRefreshToken(foundUser.RefreshTokenId)
	if getErr != nil {
		log.Print(getErr.Error())
		return errors.New("Couldn't find token")
	}
	invalidatedToken, invalidateErr := db.InvalidateToken(token.ID)
	if invalidatedToken.IsValid != false || invalidateErr != nil {
		log.Print(invalidateErr.Error())
		return errors.New("Could not invalidate token")
	}
	return nil
}

func (db *DataBaseClient) GenerateRefreshToken(userId int) (types.RefreshToken, error) {
	// Refresh Token
	c := 10
	rndByteArr := make([]byte, c)
	_, readErr := rand.Read(rndByteArr)
	if readErr != nil {
		return types.RefreshToken{}, readErr
	}
	encodedString := hex.EncodeToString(rndByteArr)
	refreshToken, storeErr := db.StoreUsersRefreshToken(encodedString, userId)
	if storeErr != nil {
		return types.RefreshToken{}, storeErr
	}
	return refreshToken, nil
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
