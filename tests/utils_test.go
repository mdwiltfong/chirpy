package tests

import (
	"encoding/json"
	"github.com/mdwiltfong/chirpy/utils"
	"github.com/mdwiltfong/chirpy/utils/types"
	"os"
	"testing"
)

func cleanUp(t *testing.T) {
	delErr := os.Remove("../database/database.json")
	if delErr != nil {
		t.Fatal(delErr)
	}
}

func TestDbFunctions(t *testing.T) {
	dbClient, err := utils.NewDB("../database/database.json")
	if err != nil {
		t.Fatalf("There was an issue creating a DB connection: %s", err)
	}
	if dbClient.Path != "../database/database.json" {
		t.Fatal("Db path is incorrect")
	}
	// Check that a database.json file was actually made
	dataBytes, readErr := os.ReadFile("../database/database.json")
	tempStruct := types.Database{Chirps: make(map[int]types.Chirp), Users: make(map[int]types.User)}
	json.Unmarshal(dataBytes, tempStruct)
	if readErr != nil {
		t.Fatalf("There was an issue in finding the database file: %s", readErr.Error())
	}
	if tempStruct.Chirps == nil {
		t.Fatal("Unable to store chirps")
	}
	if tempStruct.Users == nil {
		t.Fatal("Unable to store users")
	}
	cleanUp(t)
}

func TestLoadDB(t *testing.T) {
	dbClient, _ := utils.NewDB("../database/database.json")
	_, err := dbClient.LoadDB()
	if err != nil {
		t.Fatal(err.Error())
	}
	cleanUp(t)
}

func TestWriteDB(t *testing.T) {
	dbClient, _ := utils.NewDB("../database/database.json")
	dbData, _ := dbClient.LoadDB()
	dbData.Chirps[3] = types.Chirp{ID: 3, Body: "Test Chirp"}
	err := dbClient.WriteDB(dbData)
	if err != nil {
		t.Fatal(err.Error())
	}
	// Check that db file has new chirp
	dataBytes, err := os.ReadFile("../database/database.json")
	tempStruct := types.Database{}
	json.Unmarshal(dataBytes, &tempStruct)
	if tempStruct.Chirps[3].ID != 3 {
		t.Fatal("Incorrect ID stored")
	}
	if tempStruct.Chirps[3].Body != "Test Chirp" {
		t.Fatal("Incorrect Body stored")
	}
	cleanUp(t)
}
func TestEnsureDB(t *testing.T) {
	dbClient, _ := utils.NewDB("../database/database.json")
	os.Remove("../database/database.json")
	dbClient.EnsureDB()
	data, _ := os.ReadFile("../database/database.json")
	tempData := types.Database{}
	json.Unmarshal(data, &tempData)

	cleanUp(t)

}

func TestGetChirps(t *testing.T) {
	dbClient, _ := utils.NewDB("../database/database.json")
	chirps, err := dbClient.GetChirps()
	if err != nil {
		t.Fatal(err.Error())
	}
	if chirps == nil {
		t.Fatal("Unable to create chirps array")
	}
	cleanUp(t)
}

func TestCreateChirps(t *testing.T) {
	dbClient, _ := utils.NewDB("../database/database.json")
	chirp, err := dbClient.CreateChirp("Test Chirp")
	if chirp.Body != "Test Chirp" {
		t.Fatalf("Chirp body is incorrect. Was expecting: %s , but got %s instead", "Test Chirp", chirp.Body)
	}
	if err != nil {
		t.Fatal(err.Error())
	}
	chirps, _ := dbClient.GetChirps()
	if chirps[0].Body != "Test Chirp" {
		t.Fatal("New chirp isn't being saved into disk")
	}
	cleanUp(t)
}
